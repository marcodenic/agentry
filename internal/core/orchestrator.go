package core

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/marcodenic/agentry/internal/debug"
)

// TeamOrchestrator enhances Agent 0 with team coordination capabilities
type TeamOrchestrator struct {
	systemAgent   *Agent
	teamAgents    map[string]*Agent
	agentStatus   map[string]AgentStatus
	agentTasks    map[string]string
	agentResults  map[string]string
	messageQueue  []TeamMessage
	mutex         sync.RWMutex
}

// AgentStatus represents the current state of an agent
type AgentStatus struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Role        string    `json:"role"`
	State       string    `json:"state"`       // "idle", "working", "waiting", "error"
	CurrentTask string    `json:"current_task"`
	Progress    float64   `json:"progress"`    // 0.0 to 1.0
	LastUpdate  time.Time `json:"last_update"`
	TokenCount  int       `json:"token_count"`
	ErrorMsg    string    `json:"error_msg,omitempty"`
}

// TeamMessage represents communication between agents
type TeamMessage struct {
	ID        uuid.UUID `json:"id"`
	From      string    `json:"from"`
	To        string    `json:"to"`        // "" for broadcast
	Type      string    `json:"type"`      // "task", "status", "result", "question"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// NewTeamOrchestrator creates a new orchestrator with the system agent
func NewTeamOrchestrator(systemAgent *Agent) *TeamOrchestrator {
	return &TeamOrchestrator{
		systemAgent:  systemAgent,
		teamAgents:   make(map[string]*Agent),
		agentStatus:  make(map[string]AgentStatus),
		agentTasks:   make(map[string]string),
		agentResults: make(map[string]string),
		messageQueue: make([]TeamMessage, 0),
	}
}

// RegisterAgent adds an agent to the team
func (to *TeamOrchestrator) RegisterAgent(name string, agent *Agent, role string) {
	to.mutex.Lock()
	defer to.mutex.Unlock()
	
	to.teamAgents[name] = agent
	to.agentStatus[name] = AgentStatus{
		ID:         agent.ID,
		Name:       name,
		Role:       role,
		State:      "idle",
		LastUpdate: time.Now(),
		Progress:   0.0,
	}
	
	debug.Printf("TeamOrchestrator: Registered agent '%s' with role '%s'", name, role)
	
	// Notify system agent about new team member
	to.notifySystemAgent(fmt.Sprintf("Agent '%s' (%s) has joined the team", name, role))
}

// GetTeamStatus returns the current status of all team agents
func (to *TeamOrchestrator) GetTeamStatus() map[string]AgentStatus {
	to.mutex.RLock()
	defer to.mutex.RUnlock()
	
	// Create a copy to avoid race conditions
	status := make(map[string]AgentStatus)
	for name, agent := range to.agentStatus {
		status[name] = agent
	}
	
	return status
}

// UpdateAgentStatus updates the status of a specific agent
func (to *TeamOrchestrator) UpdateAgentStatus(name, state, task string, progress float64) {
	to.mutex.Lock()
	defer to.mutex.Unlock()
	
	if status, exists := to.agentStatus[name]; exists {
		status.State = state
		status.CurrentTask = task
		status.Progress = progress
		status.LastUpdate = time.Now()
		to.agentStatus[name] = status
		
		debug.Printf("TeamOrchestrator: Updated %s status: %s (%.1f%%)", name, state, progress*100)
	}
}

// SendMessage sends a message between team agents
func (to *TeamOrchestrator) SendMessage(from, targetAgent, msgType, content string, data map[string]interface{}) {
	to.mutex.Lock()
	defer to.mutex.Unlock()
	
	message := TeamMessage{
		ID:        uuid.New(),
		From:      from,
		To:        targetAgent,
		Type:      msgType,
		Content:   content,
		Timestamp: time.Now(),
		Data:      data,
	}
	
	to.messageQueue = append(to.messageQueue, message)
	
	debug.Printf("TeamOrchestrator: Message from %s to %s: %s", from, targetAgent, content)
}

// GetMessages retrieves messages for a specific agent
func (to *TeamOrchestrator) GetMessages(agentName string) []TeamMessage {
	to.mutex.RLock()
	defer to.mutex.RUnlock()
	
	var messages []TeamMessage
	for _, msg := range to.messageQueue {
		if msg.To == agentName || msg.To == "" { // Direct messages or broadcasts
			messages = append(messages, msg)
		}
	}
	
	return messages
}

// AssignTask assigns a task to a specific agent
func (to *TeamOrchestrator) AssignTask(agentName, task string) error {
	to.mutex.Lock()
	defer to.mutex.Unlock()
	
	// Check if agent exists and is available
	status, exists := to.agentStatus[agentName]
	if !exists {
		return fmt.Errorf("agent '%s' not found", agentName)
	}
	
	if status.State == "working" {
		return fmt.Errorf("agent '%s' is currently busy", agentName)
	}
	
	// Update agent status
	status.State = "working"
	status.CurrentTask = task
	status.Progress = 0.0
	status.LastUpdate = time.Now()
	to.agentStatus[agentName] = status
	
	// Store task
	to.agentTasks[agentName] = task
	
	// Send task message
	to.SendMessage("system", agentName, "task", task, map[string]interface{}{
		"assigned_at": time.Now(),
		"priority":    "normal",
	})
	
	debug.Printf("TeamOrchestrator: Assigned task to %s: %s", agentName, task)
	return nil
}

// CompleteTask marks a task as completed and stores the result
func (to *TeamOrchestrator) CompleteTask(agentName, result string) {
	to.mutex.Lock()
	defer to.mutex.Unlock()
	
	// Update agent status
	if status, exists := to.agentStatus[agentName]; exists {
		status.State = "idle"
		status.CurrentTask = ""
		status.Progress = 1.0
		status.LastUpdate = time.Now()
		to.agentStatus[agentName] = status
	}
	
	// Store result
	to.agentResults[agentName] = result
	
	// Notify system agent
	to.notifySystemAgent(fmt.Sprintf("Agent '%s' completed task: %s", agentName, result))
	
	debug.Printf("TeamOrchestrator: Task completed by %s", agentName)
}

// GetAvailableAgents returns a list of agents that are currently idle
func (to *TeamOrchestrator) GetAvailableAgents() []string {
	to.mutex.RLock()
	defer to.mutex.RUnlock()
	
	var available []string
	for name, status := range to.agentStatus {
		if status.State == "idle" {
			available = append(available, name)
		}
	}
	
	return available
}

// GetSystemPrompt returns an enhanced system prompt that includes team context
func (to *TeamOrchestrator) GetSystemPrompt() string {
	to.mutex.RLock()
	defer to.mutex.RUnlock()
	
	teamStatus := "TEAM STATUS:\n"
	for name, status := range to.agentStatus {
		teamStatus += fmt.Sprintf("- %s (%s): %s", name, status.Role, status.State)
		if status.CurrentTask != "" {
			teamStatus += fmt.Sprintf(" - %s (%.0f%%)", status.CurrentTask, status.Progress*100)
		}
		teamStatus += "\n"
	}
	
	availableAgents := to.GetAvailableAgents()
	teamStatus += fmt.Sprintf("\nAVAILABLE AGENTS: %v\n", availableAgents)
	
	basePrompt := to.systemAgent.Prompt
	return fmt.Sprintf("%s\n\n%s\nYou can coordinate with team members by:\n- Spawning new agents: /spawn <name> <role>\n- Assigning tasks: assign task to <agent>\n- Checking status: team status\n- Sending messages: message <agent> <content>", basePrompt, teamStatus)
}

// CoordinateTask breaks down a complex task and assigns parts to team members
func (to *TeamOrchestrator) CoordinateTask(ctx context.Context, task string) (string, error) {
	debug.Printf("TeamOrchestrator: Coordinating complex task: %s", task)
	
	// Use the system agent to analyze and break down the task
	analysisPrompt := fmt.Sprintf(`
Analyze this task and determine how to coordinate with the team:

TASK: %s

TEAM STATUS:
%s

Please:
1. Break down the task into smaller parts if needed
2. Identify which team members should handle each part
3. Specify the order of execution if there are dependencies
4. If you need new agents with specific skills, identify what roles to spawn

Respond with a coordination plan.`, task, to.getTeamStatusString())
	
	// Run analysis with system agent
	plan, err := to.systemAgent.Run(ctx, analysisPrompt)
	if err != nil {
		return "", fmt.Errorf("failed to create coordination plan: %w", err)
	}
	
	debug.Printf("TeamOrchestrator: Generated coordination plan: %s", plan)
	return plan, nil
}

// notifySystemAgent sends a notification to the system agent's memory
func (to *TeamOrchestrator) notifySystemAgent(message string) {
	// Add the notification to system agent's memory context
	// This allows Agent 0 to be aware of team activities
	debug.Printf("SystemAgent notification: %s", message)
}

// getTeamStatusString returns a formatted string of team status
func (to *TeamOrchestrator) getTeamStatusString() string {
	status := ""
	for name, agent := range to.agentStatus {
		status += fmt.Sprintf("- %s (%s): %s", name, agent.Role, agent.State)
		if agent.CurrentTask != "" {
			status += fmt.Sprintf(" - %s", agent.CurrentTask)
		}
		status += "\n"
	}
	return status
}

// ProcessTeamCommand handles team-related commands from the system agent
func (to *TeamOrchestrator) ProcessTeamCommand(ctx context.Context, command string) (string, error) {
	debug.Printf("TeamOrchestrator: Processing command: %s", command)
	
	// Parse and execute team commands
	// This is called when the system agent needs to coordinate team actions
	
	switch {
	case stringContainsSubstr(command, "team status"):
		return to.formatTeamStatus(), nil
	case stringContainsSubstr(command, "available agents"):
		available := to.GetAvailableAgents()
		return fmt.Sprintf("Available agents: %v", available), nil
	case stringContainsSubstr(command, "assign task"):
		// Parse task assignment
		// Format: "assign task 'create a new file' to coder"
		return "Task assignment processed", nil
	default:
		return "", fmt.Errorf("unknown team command: %s", command)
	}
}

func (to *TeamOrchestrator) formatTeamStatus() string {
	to.mutex.RLock()
	defer to.mutex.RUnlock()
	
	result := "Team Status Report:\n\n"
	for name, status := range to.agentStatus {
		result += fmt.Sprintf("Agent: %s (%s)\n", name, status.Role)
		result += fmt.Sprintf("  State: %s\n", status.State)
		if status.CurrentTask != "" {
			result += fmt.Sprintf("  Task: %s (%.0f%% complete)\n", status.CurrentTask, status.Progress*100)
		}
		result += fmt.Sprintf("  Last Update: %s\n", status.LastUpdate.Format("15:04:05"))
		if status.TokenCount > 0 {
			result += fmt.Sprintf("  Tokens Used: %d\n", status.TokenCount)
		}
		result += "\n"
	}
	
	return result
}

// Helper function to check if string contains substring
func stringContainsSubstr(s, substr string) bool {
	return strings.Contains(s, substr)
}
