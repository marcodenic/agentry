package tool

import (
	"context"
	"encoding/json"
	"fmt"
)

// CollaborationTool provides collaborative capabilities for agents
type CollaborationTool struct {
	team interface{} // Will be set to team.Team, but we avoid import cycle
}

// NewCollaborationTool creates a new collaboration tool
func NewCollaborationTool(team interface{}) *CollaborationTool {
	return &CollaborationTool{team: team}
}

// Call implements the Tool interface
func (ct *CollaborationTool) Call(ctx context.Context, input json.RawMessage) (any, error) {
	var call struct {
		Action     string                 `json:"action"`
		Parameters map[string]interface{} `json:"parameters"`
	}

	if err := json.Unmarshal(input, &call); err != nil {
		return nil, fmt.Errorf("invalid collaboration tool call: %w", err)
	}

	// Use reflection or type assertion to call team methods
	// For now, return a placeholder response
	switch call.Action {
	case "send_message":
		from := getString(call.Parameters, "from")
		to := getString(call.Parameters, "to")
		messageType := getString(call.Parameters, "message_type")
		content := getString(call.Parameters, "content")

		// TODO: Call team.SendDirectMessage when integration is complete
		return map[string]interface{}{
			"status":  "message_sent",
			"from":    from,
			"to":      to,
			"type":    messageType,
			"content": content,
		}, nil

	case "request_collaboration":
		from := getString(call.Parameters, "from")
		to := getString(call.Parameters, "to")
		task := getString(call.Parameters, "task")

		return map[string]interface{}{
			"status": "collaboration_requested",
			"from":   from,
			"to":     to,
			"task":   task,
		}, nil

	case "update_status":
		agentID := getString(call.Parameters, "agent_id")
		status := getString(call.Parameters, "status")
		currentTask := getString(call.Parameters, "current_task")

		return map[string]interface{}{
			"status":       "status_updated",
			"agent_id":     agentID,
			"new_status":   status,
			"current_task": currentTask,
		}, nil

	case "get_team_status":
		return map[string]interface{}{
			"status": "team_status_retrieved",
			"agents": []map[string]interface{}{
				{"id": "coder", "status": "working", "task": "Writing code"},
				{"id": "tester", "status": "idle", "task": "Waiting for code"},
				{"id": "writer", "status": "working", "task": "Writing documentation"},
			},
		}, nil

	case "create_workflow":
		workflowID := getString(call.Parameters, "workflow_id")
		name := getString(call.Parameters, "name")

		return map[string]interface{}{
			"status":      "workflow_created",
			"workflow_id": workflowID,
			"name":        name,
		}, nil

	default:
		return nil, fmt.Errorf("unknown collaboration action: %s", call.Action)
	}
}

// Description provides tool description
func (ct *CollaborationTool) Description() string {
	return `Collaboration tool for multi-agent communication and coordination.

Available actions:
- send_message: Send direct message to another agent
  Parameters: {"from": "agent_id", "to": "agent_id", "message_type": "direct/request/status", "content": "message text"}

- request_collaboration: Request collaboration from another agent  
  Parameters: {"from": "agent_id", "to": "agent_id", "task": "description of what you need help with"}

- update_status: Update your current status and progress
  Parameters: {"agent_id": "your_id", "status": "working/idle/waiting/blocked", "current_task": "what you're doing", "progress": 0.5}

- get_team_status: Get status of all agents in the team
  Parameters: {}

- create_workflow: Create a multi-agent collaborative workflow
  Parameters: {"workflow_id": "unique_id", "name": "workflow name", "steps": [...]}

Example usage:
{"action": "send_message", "parameters": {"from": "coder", "to": "tester", "message_type": "request", "content": "I've finished the calculator. Please test the add() and subtract() functions."}}

{"action": "request_collaboration", "parameters": {"from": "coder", "to": "writer", "task": "Please write documentation for the calculator module I just created"}}

{"action": "update_status", "parameters": {"agent_id": "coder", "status": "working", "current_task": "Implementing calculator multiply function", "progress": 0.7}}

{"action": "get_team_status", "parameters": {}}
`
}

// Helper function to safely get string from parameters
func getString(params map[string]interface{}, key string) string {
	if val, ok := params[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}
