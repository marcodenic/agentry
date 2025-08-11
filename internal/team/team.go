package team

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/memstore"
	"github.com/marcodenic/agentry/internal/tool"
)

// Team manages a multi-agent conversation step by step.
// This is a simplified version that consolidates the functionality
// from converse.Team and maintains compatibility.
type Team struct {
	parent       *core.Agent
	agents       map[string]*Agent // Changed to use Agent type
	agentsByName map[string]*Agent // Changed to use Agent type
	names        []string
	tasks        map[string]*Task
	messages     []Message
	roles        map[string]*RoleConfig
	portRange    PortRange
	name         string
	maxTurns     int
	mutex        sync.RWMutex
	// ENHANCED: Shared memory and communication tracking
	sharedMemory map[string]interface{} // Shared data between agents
	store        memstore.SharedStore   // Durable-backed store (in-memory by default)
	coordination []CoordinationEvent    // Log of coordination events
}

// NewTeam creates a new team with the given parent agent.
func NewTeam(parent *core.Agent, maxTurns int, name string) (*Team, error) {
	team := &Team{
		parent:       parent,
		maxTurns:     maxTurns,
		name:         name,
		agents:       make(map[string]*Agent),
		agentsByName: make(map[string]*Agent),
		tasks:        make(map[string]*Task),
		messages:     make([]Message, 0),
		roles:        make(map[string]*RoleConfig),
		portRange:    PortRange{Start: 9000, End: 9099}, // ENHANCED: Initialize shared memory and coordination tracking
		sharedMemory: make(map[string]interface{}),
		store:        memstore.Get(),
		coordination: make([]CoordinationEvent, 0),
	}

	// Kick off default GC for the store (once-per-process)
	memstore.StartDefaultGC(60 * time.Second)

	// Best-effort: load persisted coordination events for this team
	team.loadCoordinationFromStore()

	return team, nil
}

// NewTeamWithRoles creates a new team with the given parent agent and loads role configurations.
func NewTeamWithRoles(parent *core.Agent, maxTurns int, name string, includePaths []string, configDir string) (*Team, error) {
	team, err := NewTeam(parent, maxTurns, name)
	if err != nil {
		return nil, err
	}

	// Load role configurations from include paths
	if len(includePaths) > 0 {
		roles, err := LoadRolesFromIncludePaths(includePaths, configDir)
		if err != nil {
			// Don't fail completely, just warn and continue with empty roles
			fmt.Printf("Warning: failed to load some roles: %v\n", err)
		}

		// Add loaded roles to team
		for name, role := range roles {
			team.roles[name] = role
			if os.Getenv("AGENTRY_TUI_MODE") != "1" {
				fmt.Fprintf(os.Stderr, "📋 Team role loaded: %s\n", name)
			}
		}
	}

	return team, nil
}

// GetRoles returns the loaded role configurations by name.
func (t *Team) GetRoles() map[string]*RoleConfig {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	out := make(map[string]*RoleConfig, len(t.roles))
	for k, v := range t.roles {
		out[k] = v
	}
	return out
}

// Add registers ag under name so it can be addressed via Call.
func (t *Team) Add(name string, ag *core.Agent) {
	// CRITICAL: Remove the "agent" tool from added agents to prevent delegation cascading
	if _, hasAgent := ag.Tools["agent"]; hasAgent {
		// Create a new registry without the agent tool
		newTools := make(tool.Registry)
		for toolName, toolInstance := range ag.Tools {
			if toolName != "agent" {
				newTools[toolName] = toolInstance
			}
		}
		ag.Tools = newTools

	}

	// Create wrapper
	agent := &Agent{
		ID:        name,
		Name:      name,
		Agent:     ag,
		Status:    "ready",
		StartedAt: time.Now(),
		LastSeen:  time.Now(),
		Metadata:  make(map[string]string),
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.agents[name] = agent
	t.agentsByName[name] = agent
	t.names = append(t.names, name)
}

// AddAgent creates a new agent and adds it to the team.
// Returns the core agent and its assigned name.
func (t *Team) AddAgent(name string) (*core.Agent, string) {
	// FIXED: Create agent with FULL tool registry instead of inheriting restricted parent tools
	// This prevents the tool inheritance bug where spawned agents get Agent 0's restricted tools
	registry := tool.DefaultRegistry() // Get all available tools

	// Create new agent with full capabilities, not inherited restrictions
	// Pass parent's tracer to enable proper trace events for spawned agents
	coreAgent := core.New(t.parent.Client, t.parent.ModelName, registry, memory.NewInMemory(), memory.NewInMemoryVector(), t.parent.Tracer)

	// Set role-appropriate prompt
	coreAgent.Prompt = fmt.Sprintf("You are a %s agent specialized in %s tasks. You have access to all necessary tools to complete your assignments.", name, name)

	// Configure error handling for resilience
	coreAgent.ErrorHandling.TreatErrorsAsResults = true
	coreAgent.ErrorHandling.MaxErrorRetries = 3
	coreAgent.ErrorHandling.IncludeErrorContext = true

	// Remove ONLY the "agent" tool to prevent delegation cascading
	// Keep all other tools (create, write, edit_range, etc.) so agents can actually work
	delete(coreAgent.Tools, "agent")

	// Note: Cost manager is already initialized in core.New()

	// Create wrapper
	agent := &Agent{
		ID:        name, // Use name as ID for simplicity
		Name:      name,
		Agent:     coreAgent,
		Status:    "ready",
		StartedAt: time.Now(),
		LastSeen:  time.Now(),
		Metadata:  make(map[string]string),
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.agents[name] = agent
	t.agentsByName[name] = agent
	t.names = append(t.names, name)

	return coreAgent, name
}

// Call implements the Caller interface for compatibility with existing code.
// It delegates work to the named agent with enhanced communication logging.
func (t *Team) Call(ctx context.Context, agentID, input string) (string, error) {
	// ENHANCED: Log explicit agent-to-agent communication
	debugPrintf("\n🔄 AGENT DELEGATION: Agent 0 -> %s\n", agentID)
	debugPrintf("📝 Task: %s\n", input)
	debugPrintf("⏰ Timestamp: %s\n", time.Now().Format("15:04:05"))

	// Log coordination event
	t.LogCoordinationEvent("delegation", "agent_0", agentID, input, map[string]interface{}{
		"task_length": len(input),
		"agent_type":  agentID,
	})

	t.mutex.RLock()
	agent, exists := t.agentsByName[agentID]
	t.mutex.RUnlock()

	if !exists {
		debugPrintf("🆕 Creating new agent: %s\n", agentID)
		// If agent doesn't exist, create it using SpawnAgent for proper model selection
		spawnedAgent, err := t.SpawnAgent(ctx, agentID, agentID)
		if err != nil {
			debugPrintf("❌ Failed to spawn agent %s: %v\n", agentID, err)
			return "", fmt.Errorf("failed to spawn agent %s: %w", agentID, err)
		}
		agent = spawnedAgent
		debugPrintf("✅ Agent %s created and ready\n", agentID)
	} else {
		debugPrintf("♻️  Using existing agent: %s (Status: %s)\n", agentID, agent.Status)
	}

	// Update agent status
	agent.SetStatus("working")

	// Log delegation start
	debugPrintf("🚀 Starting task execution on agent %s...\n", agentID)

	// Log the communication to file as well
	logMessage := fmt.Sprintf("DELEGATION: Agent 0 -> %s | Task: %s", agentID, input)
	logToFile(logMessage)

	// Inject inbox into prompt for this turn (lightweight option)
	originalPrompt := agent.Agent.Prompt
	// Collect unread inbox messages
	inbox := t.GetAgentInbox(agentID)
	unread := make([]map[string]interface{}, 0, len(inbox))
	for _, m := range inbox {
		if read, ok := m["read"].(bool); !ok || !read {
			unread = append(unread, m)
		}
	}
	if len(unread) > 0 {
		// Build an INBOX section appended to the system prompt
		var sb strings.Builder
		sb.WriteString(originalPrompt)
		sb.WriteString("\n\nINBOX: You have ")
		sb.WriteString(fmt.Sprintf("%d unread message(s). Read and consider them before continuing.\n", len(unread)))
		for _, m := range unread {
			from, _ := m["from"].(string)
			msg, _ := m["message"].(string)
			ts := ""
			if tv, ok := m["timestamp"].(time.Time); ok {
				ts = tv.Format("15:04:05")
			}
			sb.WriteString("- ")
			if ts != "" { sb.WriteString("["); sb.WriteString(ts); sb.WriteString("] ") }
			if from != "" { sb.WriteString(from); sb.WriteString(": ") }
			sb.WriteString(msg)
			sb.WriteString("\n")
		}
		agent.Agent.Prompt = sb.String()
	}

	// Execute the input on the core agent with a bounded timeout to avoid indefinite hangs
	timeout := 120 * time.Second
	if v := os.Getenv("AGENTRY_DELEGATION_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			timeout = d
		}
	}
	debugPrintf("🔧 Call: Creating context with timeout %s for agent %s", timeout, agentID)
	dctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Skip workspace event publishing in TUI mode to prevent console interference
	if os.Getenv("AGENTRY_TUI_MODE") != "1" {
		t.PublishWorkspaceEvent("agent_0", "delegation_started", fmt.Sprintf("Delegated to %s", agentID), map[string]interface{}{"agent": agentID, "timeout": timeout.String()})
	}

	debugPrintf("🔧 Call: About to call runAgent for %s", agentID)
	startTime := time.Now()
	result, err := runAgent(dctx, agent.Agent, input, agentID, t.names)
	duration := time.Since(startTime)
	debugPrintf("🔧 Call: runAgent completed for %s in %s", agentID, duration)
	
	if err != nil {
		debugPrintf("❌ Call: runAgent failed for %s: %v", agentID, err)
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			// Explicit timeout handling
			msg := fmt.Sprintf("⏳ Delegation to '%s' timed out after %s. Consider simplifying the task, choosing a different agent, or increasing AGENTRY_DELEGATION_TIMEOUT.", agentID, timeout)
			t.LogCoordinationEvent("delegation_timeout", agentID, "agent_0", msg, map[string]interface{}{"timeout": timeout.String()})
			// Skip workspace event publishing in TUI mode
			if os.Getenv("AGENTRY_TUI_MODE") != "1" {
				t.PublishWorkspaceEvent("agent_0", "delegation_timeout", msg, map[string]interface{}{"agent": agentID})
			}
			return msg, nil
		}
	}

	// Restore original prompt and mark inbox messages as read after processing
	agent.Agent.Prompt = originalPrompt
	if len(unread) > 0 {
		t.MarkMessagesAsRead(agentID)
	}

	// Update agent status and handle errors gracefully
	if err != nil {
		agent.SetStatus("error")
		debugPrintf("❌ Agent %s failed: %v\n", agentID, err)
		logToFile(fmt.Sprintf("DELEGATION FAILED: %s | Error: %v", agentID, err))
		t.LogCoordinationEvent("delegation_failed", agentID, "agent_0", err.Error(), map[string]interface{}{
			"error": err.Error(),
		})

		// Instead of returning the error directly, format it as feedback for the parent agent
		errorFeedback := fmt.Sprintf("❌ Agent '%s' encountered an error: %v\n\nSuggestions:\n- Try a different approach\n- Simplify the request\n- Use alternative tools\n- Break the task into smaller steps",
			agentID, err)

		// Return the error as feedback instead of propagating it
		return errorFeedback, nil
	} else {
		agent.SetStatus("ready")
		debugPrintf("✅ Agent %s completed successfully\n", agentID)
		debugPrintf("📤 Result length: %d characters\n", len(result))
		debugPrintf("🧮 Agent %s final token count: %d\n", agentID, func() int {
			if agent.Agent.Cost != nil {
				return agent.Agent.Cost.TotalTokens()
			}
			return 0
		}())
		if len(result) > 100 {
			debugPrintf("📄 Result preview: %.100s...\n", result)
		} else {
			debugPrintf("📄 Result: %s\n", result)
		}
		logToFile(fmt.Sprintf("DELEGATION SUCCESS: %s | Result length: %d", agentID, len(result)))
		t.LogCoordinationEvent("delegation_success", agentID, "agent_0", "Task completed", map[string]interface{}{
			"result_length": len(result),
			"agent_type":    agentID,
		})

		// Store result in shared memory for other agents to access
		t.SetSharedData(fmt.Sprintf("last_result_%s", agentID), result)
		t.SetSharedData(fmt.Sprintf("last_task_%s", agentID), input)
	}
	debugPrintf("🏁 Delegation complete: Agent 0 <- %s\n\n", agentID)

	return result, nil
}

// runAgent executes an agent with the given input, similar to converse.runAgent
func runAgent(ctx context.Context, ag *core.Agent, input, name string, peers []string) (string, error) {
	// Attach agent name into context for builtins to use sensible defaults
	ctx = context.WithValue(ctx, tool.AgentNameContextKey, name)
	// Use the standard agent.Run() method instead of custom logic
	// This ensures that all tracing, token counting, and other instrumentation works correctly
	debugPrintf("🚀 runAgent: About to call ag.Run for agent %s with input length %d", name, len(input))
	debugPrintf("🚀 runAgent: Agent %s context timeout: %v", name, ctx.Err())
	debugPrintf("🚀 runAgent: Agent %s cost manager: %p, tokens before: %d", name, ag.Cost, func() int {
		if ag.Cost != nil {
			return ag.Cost.TotalTokens()
		}
		return 0
	}())
	
	result, err := ag.Run(ctx, input)
	
	debugPrintf("🏁 runAgent: ag.Run completed for agent %s", name)
	debugPrintf("🏁 runAgent: Result length: %d", len(result))
	debugPrintf("🏁 runAgent: Error: %v", err)
	debugPrintf("🏁 runAgent: Agent %s tokens after: %d", name, func() int {
		if ag.Cost != nil {
			return ag.Cost.TotalTokens()
		}
		return 0
	}())
	debugPrintf("🏁 runAgent: Agent %s context final state: %v", name, ctx.Err())
	
	return result, err
}

// GetAgents returns a list of all agent names in the team.
func (t *Team) GetAgents() []string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return append([]string(nil), t.names...)
}

// Names returns a list of all agent names in the team.
func (t *Team) Names() []string {
	return t.GetAgents()
}

// GetTeamAgents returns a list of all team agents with role information.
func (t *Team) GetTeamAgents() []*Agent {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	agents := make([]*Agent, 0, len(t.agents))
	for _, agent := range t.agents {
		agents = append(agents, agent)
	}
	return agents
}

// logToFile logs the message to a file (only if not in TUI mode)
func logToFile(message string) {
	if os.Getenv("AGENTRY_TUI_MODE") == "1" {
		return
	}

	file, err := os.OpenFile("agent_communication.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer file.Close()

	logger := log.New(file, "", log.LstdFlags)
	logger.Println(message)
}

// debugPrintf prints debug information only when not in TUI mode
func debugPrintf(format string, v ...interface{}) {
	if os.Getenv("AGENTRY_TUI_MODE") != "1" {
		fmt.Fprintf(os.Stderr, format, v...)
	}
}

// ENHANCED: Shared Memory and Coordination Methods

// SetSharedData stores data in shared memory accessible to all agents
func (t *Team) SetSharedData(key string, value interface{}) {
	t.mutex.Lock()
	t.sharedMemory[key] = value
	t.mutex.Unlock()

	// Persist a JSON representation to the shared store (best-effort)
	if t.store != nil {
		if b, err := json.Marshal(value); err == nil {
			_ = t.store.Set(t.name, key, b, 0)
		} else {
			// Fallback to string formatting to avoid losing data entirely
			_ = t.store.Set(t.name, key, []byte(fmt.Sprintf("%v", value)), 0)
		}
	}

	// Log the shared memory update
	event := CoordinationEvent{
		ID:        fmt.Sprintf("shared_%d", time.Now().Unix()),
		Type:      "shared_memory_update",
		From:      "system",
		To:        "*",
		Content:   fmt.Sprintf("Updated shared data: %s", key),
		Timestamp: time.Now(),
		Metadata:  map[string]interface{}{"key": key, "value_type": fmt.Sprintf("%T", value)},
	}
	t.mutex.Lock()
	t.coordination = append(t.coordination, event)
	t.mutex.Unlock()
	debugPrintf("📊 Shared memory updated: %s\n", key)
}

// GetSharedData retrieves data from shared memory
func (t *Team) GetSharedData(key string) (interface{}, bool) {
	t.mutex.RLock()
	value, exists := t.sharedMemory[key]
	t.mutex.RUnlock()
	if exists {
		return value, true
	}

	// Try backing store if not present in in-memory map
	if t.store != nil {
		if b, ok, err := t.store.Get(t.name, key); err == nil && ok {
			var out interface{}
			if err := json.Unmarshal(b, &out); err != nil {
				// treat as plain string
				out = string(b)
			}
			// normalize common JSON generic types into typed Go forms used by callers
			out = normalizeShared(out)
			// cache in memory for quick typed access in-session
			t.mutex.Lock()
			t.sharedMemory[key] = out
			t.mutex.Unlock()
			return out, true
		}
	}

	return nil, false
}

// normalizeShared converts generic []any of map[string]any into []map[string]any
// to satisfy existing callers that assert concrete types.
func normalizeShared(v interface{}) interface{} {
	switch vv := v.(type) {
	case []interface{}:
		// Check if it's a slice of maps; convert to []map[string]interface{}
		converted := make([]map[string]interface{}, 0, len(vv))
		for _, it := range vv {
			if m, ok := it.(map[string]interface{}); ok {
				converted = append(converted, m)
			} else {
				// Not homogeneous; return original
				return v
			}
		}
		return converted
	default:
		return v
	}
}

// GetAllSharedData returns all shared memory data
func (t *Team) GetAllSharedData() map[string]interface{} {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	result := make(map[string]interface{})
	for k, v := range t.sharedMemory {
		result[k] = v
	}
	return result
}

// LogCoordinationEvent adds a coordination event to the log
func (t *Team) LogCoordinationEvent(eventType, from, to, content string, metadata map[string]interface{}) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	event := CoordinationEvent{
		ID:        fmt.Sprintf("%s_%d", eventType, time.Now().UnixNano()),
		Type:      eventType,
		From:      from,
		To:        to,
		Content:   content,
		Timestamp: time.Now(),
		Metadata:  metadata,
	}
	t.coordination = append(t.coordination, event)

	// Persist the event (best-effort)
	if t.store != nil {
		if b, err := json.Marshal(event); err == nil {
			_ = t.store.Set(t.name, "coord-"+event.ID, b, 0)
		}
	}

	// Enhanced console logging
	debugPrintf("📝 COORDINATION EVENT: %s -> %s | %s: %s\n", from, to, eventType, content)
	logToFile(fmt.Sprintf("COORDINATION: %s -> %s | %s: %s", from, to, eventType, content))
}

// GetCoordinationHistory returns the coordination event history
func (t *Team) GetCoordinationHistory() []CoordinationEvent {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	result := make([]CoordinationEvent, len(t.coordination))
	copy(result, t.coordination)
	return result
}

// loadCoordinationFromStore loads persisted coordination events at startup.
func (t *Team) loadCoordinationFromStore() {
	if t.store == nil {
		return
	}
	keys, err := t.store.Keys(t.name)
	if err != nil || len(keys) == 0 {
		return
	}
	// Collect coord-* keys
	events := make([]CoordinationEvent, 0)
	for _, k := range keys {
		if len(k) < 6 || k[:6] != "coord-" {
			continue
		}
		if b, ok, err := t.store.Get(t.name, k); err == nil && ok {
			var ev CoordinationEvent
			if err := json.Unmarshal(b, &ev); err == nil {
				events = append(events, ev)
			}
		}
	}
	if len(events) == 0 {
		return
	}
	// Append to in-memory log, keep order by timestamp
	t.mutex.Lock()
	t.coordination = append(t.coordination, events...)
	t.mutex.Unlock()
}

// GetCoordinationSummary returns a summary of recent coordination events
func (t *Team) GetCoordinationSummary() string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	if len(t.coordination) == 0 {
		return "No coordination events recorded."
	}

	recent := t.coordination
	if len(recent) > 10 {
		recent = recent[len(recent)-10:] // Last 10 events
	}

	summary := fmt.Sprintf("Recent Coordination Events (%d total):\n", len(t.coordination))
	for _, event := range recent {
		summary += fmt.Sprintf("- %s: %s -> %s | %s\n",
			event.Timestamp.Format("15:04:05"),
			event.From,
			event.To,
			event.Content)
	}

	return summary
}

// CoordinationHistoryStrings returns formatted lines of coordination events.
// If limit <= 0, returns all events; otherwise returns the last 'limit' events.
func (t *Team) CoordinationHistoryStrings(limit int) []string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	if len(t.coordination) == 0 {
		return nil
	}
	start := 0
	if limit > 0 && len(t.coordination) > limit {
		start = len(t.coordination) - limit
	}
	res := make([]string, 0, len(t.coordination)-start)
	for _, e := range t.coordination[start:] {
		res = append(res, fmt.Sprintf("%s %s -> %s | %s", e.Timestamp.Format("15:04:05"), e.From, e.To, e.Content))
	}
	return res
}

// ENHANCED: Direct agent-to-agent communication methods

// SendMessageToAgent enables direct communication between agents
func (t *Team) SendMessageToAgent(ctx context.Context, fromAgentID, toAgentID, message string) error {
	t.mutex.RLock()
	_, fromExists := t.agentsByName[fromAgentID]
	_, toExists := t.agentsByName[toAgentID]
	t.mutex.RUnlock()

	if !fromExists {
		return fmt.Errorf("sender agent %s not found", fromAgentID)
	}
	if !toExists {
		return fmt.Errorf("recipient agent %s not found", toAgentID)
	}

	// Log the direct communication
	fmt.Fprintf(os.Stderr, "💬 DIRECT MESSAGE: %s → %s\n", fromAgentID, toAgentID)
	fmt.Fprintf(os.Stderr, "📝 Message: %s\n", message)

	// Store in coordination events
	t.LogCoordinationEvent("direct_message", fromAgentID, toAgentID, message, map[string]interface{}{
		"message_type": "agent_to_agent",
		"timestamp":    time.Now(),
	})

	// Add to agent's inbox (shared memory)
	inboxKey := fmt.Sprintf("inbox_%s", toAgentID)
	inbox, exists := t.GetSharedData(inboxKey)
	var messages []map[string]interface{}

	if exists {
		if existingMessages, ok := inbox.([]map[string]interface{}); ok {
			messages = existingMessages
		}
	}

	// Add new message
	newMessage := map[string]interface{}{
		"from":      fromAgentID,
		"message":   message,
		"timestamp": time.Now(),
		"read":      false,
	}
	messages = append(messages, newMessage)

	t.SetSharedData(inboxKey, messages)
	fmt.Fprintf(os.Stderr, "📬 Message delivered to %s's inbox\n", toAgentID)

	return nil
}

// BroadcastToAllAgents sends a message to all agents
func (t *Team) BroadcastToAllAgents(ctx context.Context, fromAgentID, message string) error {
	t.mutex.RLock()
	agentNames := append([]string(nil), t.names...)
	t.mutex.RUnlock()

	fmt.Fprintf(os.Stderr, "📢 BROADCAST from %s: %s\n", fromAgentID, message)

	for _, agentName := range agentNames {
		if agentName != fromAgentID { // Don't send to self
			err := t.SendMessageToAgent(ctx, fromAgentID, agentName, message)
			if err != nil {
				fmt.Fprintf(os.Stderr, "❌ Failed to broadcast to %s: %v\n", agentName, err)
			}
		}
	}

	return nil
}

// GetAgentInbox returns unread messages for an agent
func (t *Team) GetAgentInbox(agentID string) []map[string]interface{} {
	inboxKey := fmt.Sprintf("inbox_%s", agentID)
	inbox, exists := t.GetSharedData(inboxKey)

	if !exists {
		return []map[string]interface{}{}
	}

	if messages, ok := inbox.([]map[string]interface{}); ok {
		return messages
	}

	return []map[string]interface{}{}
}

// MarkMessagesAsRead marks messages in an agent's inbox as read
func (t *Team) MarkMessagesAsRead(agentID string) {
	inboxKey := fmt.Sprintf("inbox_%s", agentID)
	inbox, exists := t.GetSharedData(inboxKey)

	if exists {
		if messages, ok := inbox.([]map[string]interface{}); ok {
			for i := range messages {
				messages[i]["read"] = true
			}
			t.SetSharedData(inboxKey, messages)
		}
	}
}

// ENHANCED: Collaborative Planning and Workspace Awareness

// WorkspaceEvent represents an event in the shared workspace
type WorkspaceEvent struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // "file_created", "task_started", "task_completed", "question", "help_request"
	AgentID     string                 `json:"agent_id"`
	Description string                 `json:"description"`
	Timestamp   time.Time              `json:"timestamp"`
	Data        map[string]interface{} `json:"data"`
}

// PublishWorkspaceEvent publishes an event that all agents can see
func (t *Team) PublishWorkspaceEvent(agentID, eventType, description string, data map[string]interface{}) {
	event := WorkspaceEvent{
		ID:          fmt.Sprintf("%s_%s_%d", agentID, eventType, time.Now().Unix()),
		Type:        eventType,
		AgentID:     agentID,
		Description: description,
		Timestamp:   time.Now(),
		Data:        data,
	}

	// Store in shared memory
	eventsKey := "workspace_events"
	events, exists := t.GetSharedData(eventsKey)
	var eventList []WorkspaceEvent

	if exists {
		if existingEvents, ok := events.([]WorkspaceEvent); ok {
			eventList = existingEvents
		}
	}

	eventList = append(eventList, event)

	// Keep only last 50 events
	if len(eventList) > 50 {
		eventList = eventList[len(eventList)-50:]
	}

	t.SetSharedData(eventsKey, eventList)

	// Only log to stderr in non-TUI mode to avoid console interference
	if os.Getenv("AGENTRY_TUI_MODE") != "1" {
		fmt.Fprintf(os.Stderr, "📡 WORKSPACE EVENT: %s | %s: %s\n", agentID, eventType, description)
	}

	// Log coordination event
	t.LogCoordinationEvent("workspace_event", agentID, "*", fmt.Sprintf("%s: %s", eventType, description), map[string]interface{}{
		"event_type": eventType,
		"data":       data,
	})
}

// GetWorkspaceEvents returns recent workspace events
func (t *Team) GetWorkspaceEvents(limit int) []WorkspaceEvent {
	eventsKey := "workspace_events"
	events, exists := t.GetSharedData(eventsKey)

	if !exists {
		return []WorkspaceEvent{}
	}

	if eventList, ok := events.([]WorkspaceEvent); ok {
		if limit > 0 && len(eventList) > limit {
			return eventList[len(eventList)-limit:]
		}
		return eventList
	}

	return []WorkspaceEvent{}
}

// RequestHelp allows an agent to request help from other agents
func (t *Team) RequestHelp(ctx context.Context, agentID, helpDescription string, preferredHelper string) error {
	// Only log to stderr in non-TUI mode to avoid console interference
	if os.Getenv("AGENTRY_TUI_MODE") != "1" {
		fmt.Fprintf(os.Stderr, "🆘 HELP REQUEST from %s: %s\n", agentID, helpDescription)
	}

	// Skip workspace event publishing in TUI mode to prevent console interference
	if os.Getenv("AGENTRY_TUI_MODE") != "1" {
		t.PublishWorkspaceEvent(agentID, "help_request", helpDescription, map[string]interface{}{
			"preferred_helper": preferredHelper,
			"urgency":          "normal",
		})
	}

	// If preferred helper specified, send direct message
	if preferredHelper != "" && preferredHelper != "*" {
		message := fmt.Sprintf("Help requested: %s", helpDescription)
		return t.SendMessageToAgent(ctx, agentID, preferredHelper, message)
	}

	// Otherwise broadcast to all agents
	message := fmt.Sprintf("Help requested: %s", helpDescription)
	return t.BroadcastToAllAgents(ctx, agentID, message)
}

// ProposeCollaboration allows agents to propose working together
func (t *Team) ProposeCollaboration(ctx context.Context, proposerID, targetAgentID, proposal string) error {
	fmt.Fprintf(os.Stderr, "🤝 COLLABORATION PROPOSAL: %s → %s\n", proposerID, targetAgentID)
	fmt.Fprintf(os.Stderr, "📝 Proposal: %s\n", proposal)

	// Store proposal in shared memory
	proposalKey := fmt.Sprintf("proposal_%s_to_%s_%d", proposerID, targetAgentID, time.Now().Unix())
	proposalData := map[string]interface{}{
		"from":      proposerID,
		"to":        targetAgentID,
		"proposal":  proposal,
		"status":    "pending",
		"timestamp": time.Now(),
	}

	t.SetSharedData(proposalKey, proposalData)

	// Skip workspace event publishing in TUI mode to prevent console interference
	if os.Getenv("AGENTRY_TUI_MODE") != "1" {
		t.PublishWorkspaceEvent(proposerID, "collaboration_proposal", fmt.Sprintf("Proposed collaboration with %s", targetAgentID), map[string]interface{}{
			"target_agent": targetAgentID,
			"proposal":     proposal,
		})
	}

	// Send direct message
	message := fmt.Sprintf("Collaboration proposal: %s. Please respond with your thoughts.", proposal)
	return t.SendMessageToAgent(ctx, proposerID, targetAgentID, message)
}
