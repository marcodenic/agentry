package team

import (
    "context"
    "errors"
    "fmt"
    "os"
    "strings"
    "sync"
    "time"

    "golang.org/x/sync/errgroup"
)

// Call delegates work to the named agent with enhanced communication logging.
func (t *Team) Call(ctx context.Context, agentID, input string) (string, error) {
    timer := StartTimer(fmt.Sprintf("Call(%s)", agentID))
    defer timer.Stop()

    // Log explicit agent-to-agent communication
    debugPrintf("\n🔄 AGENT DELEGATION: Agent 0 -> %s\n", agentID)
    debugPrintf("📝 Task: %s\n", input)
    debugPrintf("⏰ Timestamp: %s\n", time.Now().Format("15:04:05"))

    // Always show delegation progress to user (not just debug mode)
    if !isTUI() {
        fmt.Fprintf(os.Stderr, "🔄 Delegating to %s agent...\n", agentID)
    }

    // Log coordination event
    t.LogCoordinationEvent("delegation", "agent_0", agentID, input, map[string]interface{}{
        "task_length": len(input),
        "agent_type":  agentID,
    })
    timer.Checkpoint("coordination logged")

    t.mutex.RLock()
    agent, exists := t.agentsByName[agentID]
    t.mutex.RUnlock()
    timer.Checkpoint("agent lookup completed")

    if !exists {
        debugPrintf("🆕 Creating new agent: %s\n", agentID)
        if !isTUI() {
            fmt.Fprintf(os.Stderr, "🆕 Creating %s agent...\n", agentID)
        }
        // Create missing agent using SpawnAgent
        spawnedAgent, err := t.SpawnAgent(ctx, agentID, agentID)
        if err != nil {
            debugPrintf("❌ Failed to spawn agent %s: %v\n", agentID, err)
            return "", fmt.Errorf("failed to spawn agent %s: %w", agentID, err)
        }
        agent = spawnedAgent
        timer.Checkpoint("new agent spawned")
        debugPrintf("✅ Agent %s created and ready\n", agentID)
        if !isTUI() {
            fmt.Fprintf(os.Stderr, "✅ %s agent ready\n", agentID)
        }
    } else {
        timer.Checkpoint("existing agent found")
        debugPrintf("♻️  Using existing agent: %s (Status: %s)\n", agentID, agent.Status)
    }

    // Update agent status
    agent.SetStatus("working")

    // Log delegation start
    debugPrintf("🚀 Starting task execution on agent %s...\n", agentID)
    if !isTUI() {
        fmt.Fprintf(os.Stderr, "🚀 %s agent working on task...\n", agentID)
    }

    // Log the communication to file as well
    logMessage := fmt.Sprintf("DELEGATION: Agent 0 -> %s | Task: %s", agentID, input)
    logToFile(logMessage)
    timer.Checkpoint("logging completed")

    // Collect unread inbox messages without mutating the agent prompt (thread-safe approach)
    inbox := t.GetAgentInbox(agentID)
    unread := make([]map[string]interface{}, 0, len(inbox))
    for _, m := range inbox {
        if read, ok := m["read"].(bool); !ok || !read {
            unread = append(unread, m)
        }
    }
    var inboxContext string
    if len(unread) > 0 {
        var sb strings.Builder
        sb.WriteString("\n\nINBOX CONTEXT (Unread Messages):\n")
        for _, m := range unread {
            from, _ := m["from"].(string)
            msg, _ := m["message"].(string)
            ts := ""
            if tv, ok := m["timestamp"].(time.Time); ok {
                ts = tv.Format("15:04:05")
            }
            sb.WriteString("- ")
            if ts != "" { sb.WriteString("["+ts+"] ") }
            if from != "" { sb.WriteString(from+": ") }
            sb.WriteString(msg)
            sb.WriteString("\n")
        }
        inboxContext = sb.String()
    }
    timer.Checkpoint("inbox processing completed")

    // Execute the input on the core agent with a reasonable timeout for complex development tasks
    timeout := 15 * time.Minute
    if v := os.Getenv("AGENTRY_DELEGATION_TIMEOUT"); v != "" {
        if d, err := time.ParseDuration(v); err == nil && d > 0 { timeout = d }
    }
    debugPrintf("🔧 Call: Creating context with timeout %s for agent %s", timeout, agentID)
    dctx, cancel := context.WithTimeout(ctx, timeout)
    defer cancel()

    // Skip workspace event publishing in TUI mode
    if !isTUI() {
        t.PublishWorkspaceEvent("agent_0", "delegation_started", fmt.Sprintf("Delegated to %s", agentID), map[string]interface{}{"agent": agentID, "timeout": timeout.String()})
    }
    timer.Checkpoint("context and events prepared")

    debugPrintf("🔧 Call: About to call runAgent for %s", agentID)
    startTime := time.Now()
    augmentedInput := input
    if inboxContext != "" {
        augmentedInput = input + inboxContext + "\n(Consider the above unread messages in your response.)"
    }
    result, err := runAgent(dctx, agent.Agent, augmentedInput, agentID, t.GetAgents())
    duration := time.Since(startTime)
    timer.Checkpoint("runAgent completed")
    debugPrintf("🔧 Call: runAgent completed for %s in %s", agentID, duration)

    if err != nil {
        debugPrintf("❌ Call: runAgent failed for %s: %v", agentID, err)
        if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
            // Check if work was actually completed despite timeout
            workCompleted := t.checkWorkCompleted(agentID, input)
            if workCompleted {
                msg := fmt.Sprintf("✅ %s agent completed the work successfully (response generation timed out after %s but files were created)", agentID, timeout)
                if !isTUI() {
                    fmt.Fprintf(os.Stderr, "✅ %s agent completed work successfully (response timed out)\n", agentID)
                }
                t.LogCoordinationEvent("delegation_success_timeout", agentID, "agent_0", msg, map[string]interface{}{"timeout": timeout.String()})
                return msg, nil
            }
            // Actual timeout without work completion
            msg := fmt.Sprintf("⏳ Delegation to '%s' timed out after %s without completing work. Consider simplifying the task, choosing a different agent, or increasing AGENTRY_DELEGATION_TIMEOUT.", agentID, timeout)
            if !isTUI() {
                fmt.Fprintf(os.Stderr, "⏳ %s agent timed out without completing work\n", agentID)
            }
            t.LogCoordinationEvent("delegation_timeout", agentID, "agent_0", msg, map[string]interface{}{"timeout": timeout.String()})
            if !isTUI() {
                t.PublishWorkspaceEvent("agent_0", "delegation_timeout", msg, map[string]interface{}{"agent": agentID})
            }
            return msg, nil
        }
    }

    // Mark inbox messages as read after processing
    if len(unread) > 0 { t.MarkMessagesAsRead(agentID) }
    timer.Checkpoint("cleanup completed")

    // Update agent status and handle errors gracefully
    if err != nil {
        agent.SetStatus("error")
        debugPrintf("❌ Agent %s failed: %v\n", agentID, err)
        logToFile(fmt.Sprintf("DELEGATION FAILED: %s | Error: %v", agentID, err))
        t.LogCoordinationEvent("delegation_failed", agentID, "agent_0", err.Error(), map[string]interface{}{"error": err.Error()})
        // Return as feedback instead of propagating error
        errorFeedback := fmt.Sprintf("❌ Agent '%s' encountered an error: %v\n\nSuggestions:\n- Try a different approach\n- Simplify the request\n- Use alternative tools\n- Break the task into smaller steps", agentID, err)
        return errorFeedback, nil
    }

    agent.SetStatus("ready")
    debugPrintf("✅ Agent %s completed successfully\n", agentID)
    if !isTUI() {
        fmt.Fprintf(os.Stderr, "✅ %s agent completed task\n", agentID)
    }
    debugPrintf("📤 Result length: %d characters\n", len(result))
    t.LogCoordinationEvent("delegation_success", agentID, "agent_0", "Task completed", map[string]interface{}{"result_length": len(result), "agent_type": agentID})
    // Store last result
    t.SetSharedData(fmt.Sprintf("last_result_%s", agentID), result)
    t.SetSharedData(fmt.Sprintf("last_task_%s", agentID), input)
    debugPrintf("🏁 Delegation complete: Agent 0 <- %s\n\n", agentID)
    return result, nil
}

// CallParallel executes multiple agent tasks in parallel for improved efficiency
func (t *Team) CallParallel(ctx context.Context, tasks []interface{}) (string, error) {
    if len(tasks) == 0 {
        return "", errors.New("no tasks provided")
    }
    eg, ctxGroup := errgroup.WithContext(ctx)
    results := make([]string, len(tasks))
    var mu sync.Mutex
    for i, taskInterface := range tasks {
        idx := i
        taskValue := taskInterface
        eg.Go(func() error {
            if ctxGroup.Err() != nil { return ctxGroup.Err() }
            task, ok := taskValue.(map[string]interface{})
            if !ok { return fmt.Errorf("task %d: invalid task format", idx) }
            agentName, ok := task["agent"].(string)
            if !ok || agentName == "" { return fmt.Errorf("task %d: agent name is required", idx) }
            input, ok := task["input"].(string)
            if !ok || input == "" { return fmt.Errorf("task %d: input is required", idx) }
            trimmed := input
            if len(trimmed) > 50 { trimmed = trimmed[:50] }
            debugPrintf("🚀 Starting parallel task %d: %s -> %s", idx, agentName, trimmed)
            res, err := t.Call(ctxGroup, agentName, input)
            if err != nil { return fmt.Errorf("task %d (%s) failed: %w", idx, agentName, err) }
            mu.Lock(); results[idx] = res; mu.Unlock()
            return nil
        })
    }
    if err := eg.Wait(); err != nil { return "", err }
    var combinedResult strings.Builder
    combinedResult.WriteString("📋 **Parallel Agent Execution Results:**\n\n")
    for i, result := range results {
        taskInterface := tasks[i]
        task := taskInterface.(map[string]interface{})
        agentName := task["agent"].(string)
        combinedResult.WriteString(fmt.Sprintf("**Agent %d (%s):**\n", i+1, agentName))
        combinedResult.WriteString(result)
        if i < len(results)-1 { combinedResult.WriteString("\n\n---\n\n") }
    }
    debugPrintf("✅ Parallel execution completed successfully with %d agents", len(tasks))
    return combinedResult.String(), nil
}
