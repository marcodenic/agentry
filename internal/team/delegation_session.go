package team

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

type delegationSession struct {
	team    *Team
	agentID string
	input   string

	timer     *Timer
	startTime time.Time

	agent            *Agent
	workspaceContext string
	timeout          time.Duration
	notifier         delegationNotifier

	ctx    context.Context
	runCtx context.Context
	cancel context.CancelFunc
}

func newDelegationSession(team *Team, agentID, input string) *delegationSession {
	return &delegationSession{team: team, agentID: agentID, input: input}
}

func (s *delegationSession) Run(ctx context.Context) (string, error) {
	s.ctx = ctx
	s.timer = StartTimer(fmt.Sprintf("Call(%s)", s.agentID))
	defer s.timer.Stop()

	s.logDelegationStart()
	if err := s.ensureAgent(); err != nil {
		return "", err
	}

	s.agent.SetStatus("working")
	s.logWorkStart()
	s.logDelegationFile()
	s.prepareWorkspaceContext()
	s.configureTimeout()
	if s.cancel != nil {
		defer s.cancel()
	}
	s.publishStartEvent()

	s.startTime = time.Now()
	debugPrintf("🔧 Call: About to call runAgent for %s", s.agentID)
	result, err := runAgent(s.runCtx, s.agent.Agent, s.augmentedInput(), s.agentID, s.team.GetAgents())
	duration := time.Since(s.startTime)
	s.timer.Checkpoint("runAgent completed")
	debugPrintf("🔧 Call: runAgent completed for %s in %s", s.agentID, duration)

	outcome, outcomeErr := s.processOutcome(result, err)
	s.timer.Checkpoint("cleanup completed")
	return outcome, outcomeErr
}

func (s *delegationSession) logDelegationStart() {
	debugPrintf("\n🔄 AGENT DELEGATION: Agent 0 -> %s\n", s.agentID)
	debugPrintf("📝 Task: %s\n", s.input)
	debugPrintf("⏰ Timestamp: %s\n", time.Now().Format("15:04:05"))

	s.notifier.User("🔄 Delegating to %s agent...\n", s.agentID)

	s.team.LogCoordinationEvent("delegation", "agent_0", s.agentID, s.input, map[string]interface{}{
		"task_length": len(s.input),
		"agent_type":  s.agentID,
	})
	s.timer.Checkpoint("coordination logged")
}

func (s *delegationSession) ensureAgent() error {
	s.team.mutex.RLock()
	agent, exists := s.team.agentsByName[s.agentID]
	s.team.mutex.RUnlock()
	s.timer.Checkpoint("agent lookup completed")

	if !exists {
		debugPrintf("🆕 Creating new agent: %s\n", s.agentID)
		s.notifier.User("🆕 Creating %s agent...\n", s.agentID)

		spawnedAgent, err := s.team.SpawnAgent(s.ctx, s.agentID, s.agentID)
		if err != nil {
			debugPrintf("❌ Failed to spawn agent %s: %v\n", s.agentID, err)
			return fmt.Errorf("failed to spawn agent %s: %w", s.agentID, err)
		}

		s.timer.Checkpoint("new agent spawned")
		agent = spawnedAgent
		debugPrintf("✅ Agent %s created and ready\n", s.agentID)
		s.notifier.User("✅ %s agent ready\n", s.agentID)
	} else {
		s.timer.Checkpoint("existing agent found")
		debugPrintf("♻️  Using existing agent: %s (Status: %s)\n", s.agentID, agent.Status)
	}

	s.agent = agent
	return nil
}

func (s *delegationSession) logWorkStart() {
	debugPrintf("🚀 Starting task execution on agent %s...\n", s.agentID)
	s.notifier.User("🚀 %s agent working on task...\n", s.agentID)
}

func (s *delegationSession) logDelegationFile() {
	s.notifier.File("DELEGATION: Agent 0 -> %s | Task: %s", s.agentID, s.input)
	s.timer.Checkpoint("logging completed")
}

func (s *delegationSession) prepareWorkspaceContext() {
	events := s.team.GetWorkspaceEvents(5)
	s.timer.Checkpoint("inbox processing completed")
	if len(events) == 0 {
		s.workspaceContext = ""
		s.timer.Checkpoint("context and events prepared")
		return
	}

	var sb strings.Builder
	sb.WriteString("\n\nRECENT WORKSPACE EVENTS:\n")
	for _, e := range events {
		ts := ""
		if !e.Timestamp.IsZero() {
			ts = e.Timestamp.Format("15:04:05")
		}
		sb.WriteString("- ")
		if ts != "" {
			sb.WriteString("[" + ts + "] ")
		}
		if e.AgentID != "" {
			sb.WriteString(e.AgentID + " | ")
		}
		sb.WriteString(e.Type)
		if e.Description != "" {
			sb.WriteString(": " + e.Description)
		}
		sb.WriteString("\n")
	}
	s.workspaceContext = sb.String()
	s.timer.Checkpoint("context and events prepared")
}

func (s *delegationSession) configureTimeout() {
	timeout := 15 * time.Minute
	if v := os.Getenv("AGENTRY_DELEGATION_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			timeout = d
		}
	}

	s.timeout = timeout
	debugPrintf("🔧 Call: Creating context with timeout %s for agent %s", timeout, s.agentID)
	s.runCtx, s.cancel = context.WithTimeout(s.ctx, timeout)
}

func (s *delegationSession) publishStartEvent() {
	if isTUI() {
		return
	}

	s.team.PublishWorkspaceEvent(
		"agent_0",
		"delegation_started",
		fmt.Sprintf("Delegated to %s", s.agentID),
		map[string]interface{}{"agent": s.agentID, "timeout": s.timeout.String()},
	)
}

func (s *delegationSession) augmentedInput() string {
	if strings.TrimSpace(s.workspaceContext) == "" {
		return s.input
	}
	return s.input + s.workspaceContext
}

func (s *delegationSession) processOutcome(result string, runErr error) (string, error) {
	if runErr != nil {
		return s.handleError(runErr)
	}
	return s.handleSuccess(result), nil
}

func (s *delegationSession) handleError(runErr error) (string, error) {
	debugPrintf("❌ Call: runAgent failed for %s: %v", s.agentID, runErr)
	if errors.Is(runErr, context.DeadlineExceeded) || errors.Is(runErr, context.Canceled) {
		if s.team.checkWorkCompleted(s.agentID, s.input) {
			msg := fmt.Sprintf("✅ %s agent completed the work successfully (response generation timed out after %s but files were created)", s.agentID, s.timeout)
			s.notifier.User("✅ %s agent completed work successfully (response timed out)\n", s.agentID)
			s.team.LogCoordinationEvent("delegation_success_timeout", s.agentID, "agent_0", msg, map[string]interface{}{"timeout": s.timeout.String()})
			return msg, nil
		}

		msg := fmt.Sprintf("⏳ Delegation to '%s' timed out after %s without completing work. Consider simplifying the task, choosing a different agent, or increasing AGENTRY_DELEGATION_TIMEOUT.", s.agentID, s.timeout)
		s.notifier.User("⏳ %s agent timed out without completing work\n", s.agentID)
		if !isTUI() {
			s.team.PublishWorkspaceEvent("agent_0", "delegation_timeout", msg, map[string]interface{}{"agent": s.agentID})
		}
		s.team.LogCoordinationEvent("delegation_timeout", s.agentID, "agent_0", msg, map[string]interface{}{"timeout": s.timeout.String()})
		return "", errors.New(msg)
	}

	s.agent.SetStatus("error")
	debugPrintf("❌ Agent %s failed: %v\n", s.agentID, runErr)
	logToFile(fmt.Sprintf("DELEGATION FAILED: %s | Error: %v", s.agentID, runErr))
	s.team.LogCoordinationEvent("delegation_failed", s.agentID, "agent_0", runErr.Error(), map[string]interface{}{"error": runErr.Error()})
	errorFeedback := fmt.Sprintf("❌ Agent '%s' encountered an error: %v\n\nSuggestions:\n- Try a different approach\n- Simplify the request\n- Use alternative tools\n- Break the task into smaller steps", s.agentID, runErr)
	return "", errors.New(errorFeedback)
}

func (s *delegationSession) handleSuccess(result string) string {
	s.agent.SetStatus("ready")
	debugPrintf("✅ Agent %s completed successfully\n", s.agentID)
	s.notifier.User("✅ %s agent completed task\n", s.agentID)
	debugPrintf("📤 Result length: %d characters\n", len(result))
	s.team.LogCoordinationEvent("delegation_success", s.agentID, "agent_0", "Task completed", map[string]interface{}{"result_length": len(result), "agent_type": s.agentID})
	s.team.SetSharedData(fmt.Sprintf("last_result_%s", s.agentID), result)
	s.team.SetSharedData(fmt.Sprintf("last_task_%s", s.agentID), s.input)
	debugPrintf("🏁 Delegation complete: Agent 0 <- %s\n\n", s.agentID)
	return result
}
