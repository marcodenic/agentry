package team

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// CollaborativeWorkflow represents a multi-agent collaborative workflow
type CollaborativeWorkflow struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Steps          []CollaborativeStep    `json:"steps"`
	Dependencies   map[string][]string    `json:"dependencies"` // step_id -> prerequisites
	Status         string                 `json:"status"`       // created, running, completed, failed
	CreatedAt      time.Time              `json:"created_at"`
	StartedAt      *time.Time             `json:"started_at"`
	CompletedAt    *time.Time             `json:"completed_at"`
	Results        map[string]interface{} `json:"results"`
	ActiveSteps    map[string]string      `json:"active_steps"` // step_id -> agent_id
	CompletedSteps []string               `json:"completed_steps"`
	FailedSteps    []string               `json:"failed_steps"`
}

// CollaborativeStep represents a step in a collaborative workflow
type CollaborativeStep struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	AgentID     string                 `json:"agent_id"`
	Task        string                 `json:"task"`
	Parameters  map[string]interface{} `json:"parameters"`
	Timeout     time.Duration          `json:"timeout"`
	RetryCount  int                    `json:"retry_count"`
	OnSuccess   []string               `json:"on_success"` // next steps
	OnFailure   []string               `json:"on_failure"` // failure handling steps
	Status      string                 `json:"status"`     // pending, running, completed, failed
	StartedAt   *time.Time             `json:"started_at"`
	CompletedAt *time.Time             `json:"completed_at"`
	Result      interface{}            `json:"result"`
}

// AgentCommunication handles real-time agent-to-agent communication
type AgentCommunication struct {
	channels map[string]chan AgentMessage // agent_id -> message channel
	mutex    sync.RWMutex
}

// AgentMessage represents a message between agents
type AgentMessage struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // direct, broadcast, collaboration_request, status_update
	From        string                 `json:"from"`
	To          string                 `json:"to"` // empty for broadcast
	Subject     string                 `json:"subject"`
	Content     string                 `json:"content"`
	Data        map[string]interface{} `json:"data"`
	Priority    string                 `json:"priority"` // high, normal, low
	Timestamp   time.Time              `json:"timestamp"`
	ReplyTo     string                 `json:"reply_to,omitempty"`
	RequiresAck bool                   `json:"requires_ack"`
}

// StatusTracker provides real-time status tracking for all agents
type StatusTracker struct {
	agentStatus   map[string]*DetailedAgentStatus
	globalMetrics *GlobalMetrics
	statusHistory []StatusChange
	mutex         sync.RWMutex
}

// DetailedAgentStatus provides comprehensive agent status information
type DetailedAgentStatus struct {
	AgentID       string                 `json:"agent_id"`
	Status        string                 `json:"status"` // idle, working, waiting, blocked, error, collaborating
	CurrentTask   string                 `json:"current_task"`
	TaskProgress  float64                `json:"task_progress"` // 0.0 to 1.0
	EstimatedTime *time.Duration         `json:"estimated_time"`
	Dependencies  []string               `json:"dependencies"` // waiting for these agents
	Dependents    []string               `json:"dependents"`   // these agents are waiting for this agent
	LastActivity  time.Time              `json:"last_activity"`
	Performance   map[string]interface{} `json:"performance"`     // metrics like tasks_completed, avg_time, etc.
	Capabilities  []string               `json:"capabilities"`    // what this agent can do
	CurrentLoad   float64                `json:"current_load"`    // 0.0 to 1.0
	Messages      []AgentMessage         `json:"recent_messages"` // recent communication
}

// GlobalMetrics provides system-wide collaboration metrics
type GlobalMetrics struct {
	TotalAgents       int                    `json:"total_agents"`
	ActiveAgents      int                    `json:"active_agents"`
	IdleAgents        int                    `json:"idle_agents"`
	BlockedAgents     int                    `json:"blocked_agents"`
	CompletedTasks    int                    `json:"completed_tasks"`
	PendingTasks      int                    `json:"pending_tasks"`
	FailedTasks       int                    `json:"failed_tasks"`
	AverageTaskTime   time.Duration          `json:"average_task_time"`
	SystemEfficiency  float64                `json:"system_efficiency"`  // 0.0 to 1.0
	CommunicationRate float64                `json:"communication_rate"` // messages per minute
	Bottlenecks       []string               `json:"bottlenecks"`        // agent_ids causing delays
	Collaborations    int                    `json:"active_collaborations"`
	LastUpdate        time.Time              `json:"last_update"`
	CustomMetrics     map[string]interface{} `json:"custom_metrics"`
}

// StatusChange represents a change in agent status
type StatusChange struct {
	AgentID   string                 `json:"agent_id"`
	OldStatus string                 `json:"old_status"`
	NewStatus string                 `json:"new_status"`
	Reason    string                 `json:"reason"`
	Impact    []string               `json:"impact"` // affected agents
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time              `json:"timestamp"`
	Duration  time.Duration          `json:"duration"` // time in previous status
}

// Add collaborative features to Team struct
func (t *Team) InitializeCollaboration() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// Initialize collaboration components if not already done
	if t.workflows == nil {
		t.workflows = make(map[string]*CollaborativeWorkflow)
	}
	if t.communication == nil {
		t.communication = &AgentCommunication{
			channels: make(map[string]chan AgentMessage),
		}
	}
	if t.statusTracker == nil {
		t.statusTracker = &StatusTracker{
			agentStatus:   make(map[string]*DetailedAgentStatus),
			globalMetrics: &GlobalMetrics{},
			statusHistory: make([]StatusChange, 0),
		}
	}
}

// CreateCollaborativeWorkflow creates a new multi-agent workflow
func (t *Team) CreateCollaborativeWorkflow(id, name string, steps []CollaborativeStep, dependencies map[string][]string) *CollaborativeWorkflow {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	workflow := &CollaborativeWorkflow{
		ID:             id,
		Name:           name,
		Steps:          steps,
		Dependencies:   dependencies,
		Status:         "created",
		CreatedAt:      time.Now(),
		Results:        make(map[string]interface{}),
		ActiveSteps:    make(map[string]string),
		CompletedSteps: make([]string, 0),
		FailedSteps:    make([]string, 0),
	}

	if t.workflows == nil {
		t.workflows = make(map[string]*CollaborativeWorkflow)
	}
	t.workflows[id] = workflow

	// Notify agents about new workflow
	t.BroadcastMessage("system", "workflow_created", fmt.Sprintf("New workflow '%s' created", name), map[string]interface{}{
		"workflow_id": id,
		"name":        name,
		"steps":       len(steps),
	})

	return workflow
}

// StartCollaborativeWorkflow begins executing a collaborative workflow
func (t *Team) StartCollaborativeWorkflow(workflowID string) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	workflow, exists := t.workflows[workflowID]
	if !exists {
		return fmt.Errorf("workflow %s not found", workflowID)
	}

	workflow.Status = "running"
	now := time.Now()
	workflow.StartedAt = &now

	// Start steps that have no dependencies
	for _, step := range workflow.Steps {
		if len(workflow.Dependencies[step.ID]) == 0 {
			t.startWorkflowStep(workflow, &step)
		}
	}

	// Notify agents
	t.BroadcastMessage("system", "workflow_started", fmt.Sprintf("Workflow '%s' started", workflow.Name), map[string]interface{}{
		"workflow_id": workflowID,
	})

	return nil
}

// startWorkflowStep starts a specific workflow step
func (t *Team) startWorkflowStep(workflow *CollaborativeWorkflow, step *CollaborativeStep) {
	step.Status = "running"
	now := time.Now()
	step.StartedAt = &now
	workflow.ActiveSteps[step.ID] = step.AgentID

	// Send task to the assigned agent
	t.SendDirectMessage("system", step.AgentID, "workflow_task", step.Task, map[string]interface{}{
		"workflow_id": workflow.ID,
		"step_id":     step.ID,
		"parameters":  step.Parameters,
		"timeout":     step.Timeout,
	})

	// Update agent status
	t.UpdateDetailedAgentStatus(step.AgentID, "working", step.Task, 0.0, nil, []string{})
}

// SendDirectMessage sends a message directly to a specific agent
func (t *Team) SendDirectMessage(from, to, messageType, content string, data map[string]interface{}) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.communication == nil {
		t.InitializeCollaboration()
	}

	message := AgentMessage{
		ID:        fmt.Sprintf("%s_%d", from, time.Now().UnixNano()),
		Type:      messageType,
		From:      from,
		To:        to,
		Content:   content,
		Data:      data,
		Priority:  "normal",
		Timestamp: time.Now(),
	}

	// Add to communication channel if agent has one
	if channel, exists := t.communication.channels[to]; exists {
		select {
		case channel <- message:
			// Message sent successfully
		default:
			// Channel is full, log the issue
			if os.Getenv("AGENTRY_TUI_MODE") != "1" {
				fmt.Printf("Warning: Message channel for agent %s is full\n", to)
			}
		}
	}

	// Log the communication
	t.LogCoordinationEvent("direct_message", from, to, content, data)
}

// BroadcastMessage sends a message to all agents
func (t *Team) BroadcastMessage(from, messageType, content string, data map[string]interface{}) {
	t.mutex.RLock()
	agentIDs := make([]string, 0, len(t.agents))
	for agentID := range t.agents {
		agentIDs = append(agentIDs, agentID)
	}
	t.mutex.RUnlock()

	for _, agentID := range agentIDs {
		t.SendDirectMessage(from, agentID, messageType, content, data)
	}
}

// SubscribeToMessages allows an agent to receive messages
func (t *Team) SubscribeToMessages(agentID string) chan AgentMessage {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.communication == nil {
		t.InitializeCollaboration()
	}

	channel := make(chan AgentMessage, 100) // Buffered channel
	t.communication.channels[agentID] = channel
	return channel
}

// UpdateDetailedAgentStatus updates comprehensive agent status
func (t *Team) UpdateDetailedAgentStatus(agentID, status, currentTask string, progress float64, dependencies, dependents []string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.statusTracker == nil {
		t.InitializeCollaboration()
	}

	now := time.Now()
	oldStatus := "unknown"

	if existing, exists := t.statusTracker.agentStatus[agentID]; exists {
		oldStatus = existing.Status
	}

	newStatus := &DetailedAgentStatus{
		AgentID:      agentID,
		Status:       status,
		CurrentTask:  currentTask,
		TaskProgress: progress,
		Dependencies: dependencies,
		Dependents:   dependents,
		LastActivity: now,
		Performance:  make(map[string]interface{}),
		Capabilities: make([]string, 0),
		CurrentLoad:  progress,
		Messages:     make([]AgentMessage, 0),
	}

	t.statusTracker.agentStatus[agentID] = newStatus

	// Record status change
	if oldStatus != status {
		change := StatusChange{
			AgentID:   agentID,
			OldStatus: oldStatus,
			NewStatus: status,
			Timestamp: now,
		}
		t.statusTracker.statusHistory = append(t.statusTracker.statusHistory, change)
	}

	// Update global metrics
	t.updateGlobalMetrics()

	// Broadcast status update
	t.BroadcastMessage(agentID, "status_update", fmt.Sprintf("Status changed to %s", status), map[string]interface{}{
		"status":       status,
		"current_task": currentTask,
		"progress":     progress,
	})
}

// updateGlobalMetrics recalculates system-wide metrics
func (t *Team) updateGlobalMetrics() {
	if t.statusTracker == nil {
		return
	}

	metrics := t.statusTracker.globalMetrics
	metrics.TotalAgents = len(t.statusTracker.agentStatus)
	metrics.ActiveAgents = 0
	metrics.IdleAgents = 0
	metrics.BlockedAgents = 0

	var bottlenecks []string

	for agentID, status := range t.statusTracker.agentStatus {
		switch status.Status {
		case "working", "collaborating":
			metrics.ActiveAgents++
		case "idle":
			metrics.IdleAgents++
		case "blocked", "waiting":
			metrics.BlockedAgents++
		}

		// Identify bottlenecks (agents with many dependents)
		if len(status.Dependents) > 1 {
			bottlenecks = append(bottlenecks, agentID)
		}
	}

	metrics.Bottlenecks = bottlenecks
	metrics.LastUpdate = time.Now()

	// Calculate efficiency (active agents / total agents)
	if metrics.TotalAgents > 0 {
		metrics.SystemEfficiency = float64(metrics.ActiveAgents) / float64(metrics.TotalAgents)
	}
}

// GetCollaborativeWorkflow returns a workflow by ID
func (t *Team) GetCollaborativeWorkflow(workflowID string) (*CollaborativeWorkflow, error) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	if t.workflows == nil {
		return nil, fmt.Errorf("no workflows initialized")
	}

	workflow, exists := t.workflows[workflowID]
	if !exists {
		return nil, fmt.Errorf("workflow %s not found", workflowID)
	}

	return workflow, nil
}

// GetGlobalMetrics returns current system metrics
func (t *Team) GetGlobalMetrics() *GlobalMetrics {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	if t.statusTracker == nil {
		t.InitializeCollaboration()
	}

	return t.statusTracker.globalMetrics
}

// GetDetailedAgentStatus returns detailed status for an agent
func (t *Team) GetDetailedAgentStatus(agentID string) (*DetailedAgentStatus, error) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	if t.statusTracker == nil {
		return nil, fmt.Errorf("status tracker not initialized")
	}

	status, exists := t.statusTracker.agentStatus[agentID]
	if !exists {
		return nil, fmt.Errorf("agent %s not found", agentID)
	}

	return status, nil
}

// GetAllAgentStatuses returns detailed status for all agents
func (t *Team) GetAllAgentStatuses() map[string]*DetailedAgentStatus {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	if t.statusTracker == nil {
		t.InitializeCollaboration()
	}

	// Create a copy to avoid race conditions
	result := make(map[string]*DetailedAgentStatus)
	for agentID, status := range t.statusTracker.agentStatus {
		result[agentID] = status
	}
	return result
}

// CompleteWorkflowStep marks a workflow step as completed
func (t *Team) CompleteWorkflowStep(workflowID, stepID string, result interface{}) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	workflow, exists := t.workflows[workflowID]
	if !exists {
		return fmt.Errorf("workflow %s not found", workflowID)
	}

	// Find and update the step
	for i, step := range workflow.Steps {
		if step.ID == stepID {
			workflow.Steps[i].Status = "completed"
			now := time.Now()
			workflow.Steps[i].CompletedAt = &now
			workflow.Steps[i].Result = result

			// Remove from active steps
			delete(workflow.ActiveSteps, stepID)
			workflow.CompletedSteps = append(workflow.CompletedSteps, stepID)

			// Check if we can start dependent steps
			t.checkAndStartDependentSteps(workflow, stepID)

			// Check if workflow is complete
			if len(workflow.CompletedSteps) == len(workflow.Steps) {
				workflow.Status = "completed"
				workflow.CompletedAt = &now
			}

			return nil
		}
	}

	return fmt.Errorf("step %s not found in workflow %s", stepID, workflowID)
}

// checkAndStartDependentSteps starts steps that were waiting for the completed step
func (t *Team) checkAndStartDependentSteps(workflow *CollaborativeWorkflow, completedStepID string) {
	for _, step := range workflow.Steps {
		if step.Status == "pending" {
			// Check if all dependencies are completed
			canStart := true
			for _, dep := range workflow.Dependencies[step.ID] {
				if !t.isStepCompleted(workflow, dep) {
					canStart = false
					break
				}
			}

			if canStart {
				t.startWorkflowStep(workflow, &step)
			}
		}
	}
}

// isStepCompleted checks if a step is completed
func (t *Team) isStepCompleted(workflow *CollaborativeWorkflow, stepID string) bool {
	for _, completedID := range workflow.CompletedSteps {
		if completedID == stepID {
			return true
		}
	}
	return false
}

// Add workflow tracking to Team struct (add these fields to the Team struct)
func (t *Team) addWorkflowFields() {
	// This is a placeholder to document the fields that need to be added to Team struct:
	// workflows      map[string]*CollaborativeWorkflow
	// communication  *AgentCommunication
	// statusTracker  *StatusTracker
}
