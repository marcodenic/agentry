package team

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// CollaborativeTool provides collaborative capabilities to agents
type CollaborativeTool struct {
	engine *CollaborationEngine
}

// NewCollaborativeTool creates a new collaborative tool
func NewCollaborativeTool(engine *CollaborationEngine) *CollaborativeTool {
	return &CollaborativeTool{engine: engine}
}

// SendDirectMessage sends a message directly to another agent
func (ct *CollaborativeTool) SendDirectMessage(fromAgent, toAgent, message string, metadata map[string]interface{}) error {
	event := AgentEvent{
		Type: "direct_message",
		From: fromAgent,
		To:   toAgent,
		Data: map[string]interface{}{
			"message":  message,
			"metadata": metadata,
		},
		Priority: "normal",
	}

	ct.engine.eventBus.Publish(event)

	// Log the communication
	ct.engine.team.LogCoordinationEvent("direct_message", fromAgent, toAgent, message, metadata)

	return nil
}

// RequestCollaboration requests collaboration from another agent
func (ct *CollaborativeTool) RequestCollaboration(fromAgent, toAgent, taskDescription string, priority string) error {
	event := AgentEvent{
		Type: "collaboration_request",
		From: fromAgent,
		To:   toAgent,
		Data: map[string]interface{}{
			"task_description": taskDescription,
			"priority":         priority,
		},
		Priority: priority,
	}

	ct.engine.eventBus.Publish(event)

	// Update status - requesting agent is now waiting
	status := &AgentStatus{
		AgentID:      fromAgent,
		Status:       "waiting",
		CurrentTask:  fmt.Sprintf("Waiting for collaboration from %s", toAgent),
		Dependencies: []string{toAgent},
		LastUpdate:   time.Now(),
	}
	ct.engine.statusBoard.UpdateAgentStatus(fromAgent, status)

	return nil
}

// RespondToCollaboration responds to a collaboration request
func (ct *CollaborativeTool) RespondToCollaboration(fromAgent, toAgent, response string, accepted bool) error {
	event := AgentEvent{
		Type: "collaboration_response",
		From: fromAgent,
		To:   toAgent,
		Data: map[string]interface{}{
			"response": response,
			"accepted": accepted,
		},
		Priority: "high",
	}

	ct.engine.eventBus.Publish(event)

	return nil
}

// AcquireFileLock requests a lock on a file
func (ct *CollaborativeTool) AcquireFileLock(agentID, filePath, lockType string, durationMinutes int) (*FileLock, error) {
	duration := time.Duration(durationMinutes) * time.Minute
	lock, err := ct.engine.fileManager.AcquireLock(filePath, agentID, lockType, duration)

	if err == nil {
		// Notify other agents about the lock
		event := AgentEvent{
			Type: "file_locked",
			From: agentID,
			Data: map[string]interface{}{
				"file_path": filePath,
				"lock_type": lockType,
				"duration":  durationMinutes,
			},
			Priority: "normal",
		}
		ct.engine.eventBus.Publish(event)
	}

	return lock, err
}

// ReleaseFileLock releases a file lock
func (ct *CollaborativeTool) ReleaseFileLock(agentID, filePath string) error {
	err := ct.engine.fileManager.ReleaseLock(filePath, agentID)

	if err == nil {
		// Notify other agents about the release
		event := AgentEvent{
			Type: "file_unlocked",
			From: agentID,
			Data: map[string]interface{}{
				"file_path": filePath,
			},
			Priority: "normal",
		}
		ct.engine.eventBus.Publish(event)
	}

	return err
}

// NotifyFileChange notifies about file changes
func (ct *CollaborativeTool) NotifyFileChange(agentID, filePath, changeType string, metadata map[string]interface{}) {
	ct.engine.fileManager.NotifyFileChange(filePath, agentID, changeType, metadata)

	// Also send as event
	event := AgentEvent{
		Type: "file_changed",
		From: agentID,
		Data: map[string]interface{}{
			"file_path":   filePath,
			"change_type": changeType,
			"metadata":    metadata,
		},
		Priority: "normal",
	}
	ct.engine.eventBus.Publish(event)
}

// UpdateStatus updates an agent's status
func (ct *CollaborativeTool) UpdateStatus(agentID, status, currentTask string, progress float64, dependencies []string) {
	agentStatus := &AgentStatus{
		AgentID:      agentID,
		Status:       status,
		CurrentTask:  currentTask,
		Progress:     progress,
		Dependencies: dependencies,
		LastUpdate:   time.Now(),
	}

	ct.engine.statusBoard.UpdateAgentStatus(agentID, agentStatus)

	// Broadcast status update
	event := AgentEvent{
		Type: "status_update",
		From: agentID,
		Data: map[string]interface{}{
			"status":       status,
			"current_task": currentTask,
			"progress":     progress,
			"dependencies": dependencies,
		},
		Priority: "low",
	}
	ct.engine.eventBus.Publish(event)
}

// SubscribeToEvents allows an agent to subscribe to specific event types
func (ct *CollaborativeTool) SubscribeToEvents(agentID string, eventTypes []string) chan AgentEvent {
	return ct.engine.eventBus.Subscribe(agentID, eventTypes)
}

// CreateWorkflow is not supported (no separate workflow engine).
func (ct *CollaborativeTool) CreateWorkflow(_ string, _ string, _ any, _ map[string][]string) error {
	return fmt.Errorf("workflow creation is not supported; use Agent 0 to orchestrate task sequences")
}

// StartWorkflow is not supported (no separate workflow engine).
func (ct *CollaborativeTool) StartWorkflow(_ string, _ string) error {
	return fmt.Errorf("workflow start is not supported; use Agent 0 to orchestrate task sequences")
}

// GetWorkflowStatus is not supported (no separate workflow engine).
func (ct *CollaborativeTool) GetWorkflowStatus(_ string) (any, error) {
	return nil, fmt.Errorf("workflow status is not supported; no workflow engine is present")
}

// GetGlobalStatus returns overall system status
func (ct *CollaborativeTool) GetGlobalStatus() *GlobalStatus {
	return ct.engine.statusBoard.GetGlobalStatus()
}

// GetTeamStatus returns status of all agents
func (ct *CollaborativeTool) GetTeamStatus() map[string]*AgentStatus {
	ct.engine.statusBoard.mutex.RLock()
	defer ct.engine.statusBoard.mutex.RUnlock()

	// Create a copy to avoid race conditions
	result := make(map[string]*AgentStatus)
	for agentID, status := range ct.engine.statusBoard.agentStatus {
		result[agentID] = status
	}
	return result
}

// SendHelp responds to a help request
func (ct *CollaborativeTool) SendHelp(fromAgent, toAgent, helpType, helpContent string) error {
	event := AgentEvent{
		Type: "help_response",
		From: fromAgent,
		To:   toAgent,
		Data: map[string]interface{}{
			"help_type":    helpType,
			"help_content": helpContent,
		},
		Priority: "high",
	}

	ct.engine.eventBus.Publish(event)

	return nil
}

// RequestHelp requests help from the team
func (ct *CollaborativeTool) RequestHelp(fromAgent, helpType, description string) error {
	event := AgentEvent{
		Type: "help_request",
		From: fromAgent,
		Data: map[string]interface{}{
			"help_type":   helpType,
			"description": description,
		},
		Priority: "high",
	}

	ct.engine.eventBus.Publish(event)

	// Update agent status to waiting for help
	status := &AgentStatus{
		AgentID:     fromAgent,
		Status:      "waiting",
		CurrentTask: fmt.Sprintf("Waiting for help: %s", description),
		LastUpdate:  time.Now(),
	}
	ct.engine.statusBoard.UpdateAgentStatus(fromAgent, status)

	return nil
}

// Tool interface implementation for collaborative tool

// CollaborativeToolCall represents a collaborative tool call
type CollaborativeToolCall struct {
	Action     string                 `json:"action"`
	Parameters map[string]interface{} `json:"parameters"`
}

// Call implements the tool interface
func (ct *CollaborativeTool) Call(ctx context.Context, input json.RawMessage) (any, error) {
	var call CollaborativeToolCall
	if err := json.Unmarshal(input, &call); err != nil {
		return nil, fmt.Errorf("invalid collaborative tool call: %w", err)
	}

	switch call.Action {
	case "send_message":
		from := call.Parameters["from"].(string)
		to := call.Parameters["to"].(string)
		message := call.Parameters["message"].(string)
		metadata := make(map[string]interface{})
		if m, ok := call.Parameters["metadata"]; ok {
			metadata = m.(map[string]interface{})
		}
		return nil, ct.SendDirectMessage(from, to, message, metadata)

	case "request_collaboration":
		from := call.Parameters["from"].(string)
		to := call.Parameters["to"].(string)
		task := call.Parameters["task"].(string)
		priority := "normal"
		if p, ok := call.Parameters["priority"]; ok {
			priority = p.(string)
		}
		return nil, ct.RequestCollaboration(from, to, task, priority)

	case "acquire_file_lock":
		agentID := call.Parameters["agent_id"].(string)
		filePath := call.Parameters["file_path"].(string)
		lockType := call.Parameters["lock_type"].(string)
		duration := int(call.Parameters["duration_minutes"].(float64))
		lock, err := ct.AcquireFileLock(agentID, filePath, lockType, duration)
		return lock, err

	case "release_file_lock":
		agentID := call.Parameters["agent_id"].(string)
		filePath := call.Parameters["file_path"].(string)
		return nil, ct.ReleaseFileLock(agentID, filePath)

	case "update_status":
		agentID := call.Parameters["agent_id"].(string)
		status := call.Parameters["status"].(string)
		currentTask := call.Parameters["current_task"].(string)
		progress := call.Parameters["progress"].(float64)
		dependencies := make([]string, 0)
		if d, ok := call.Parameters["dependencies"]; ok {
			for _, dep := range d.([]interface{}) {
				dependencies = append(dependencies, dep.(string))
			}
		}
		ct.UpdateStatus(agentID, status, currentTask, progress, dependencies)
		return nil, nil

	case "get_global_status":
		return ct.GetGlobalStatus(), nil

	case "get_team_status":
		return ct.GetTeamStatus(), nil

	case "request_help":
		fromAgent := call.Parameters["from"].(string)
		helpType := call.Parameters["help_type"].(string)
		description := call.Parameters["description"].(string)
		return nil, ct.RequestHelp(fromAgent, helpType, description)

	default:
		return nil, fmt.Errorf("unknown collaborative action: %s", call.Action)
	}
}
