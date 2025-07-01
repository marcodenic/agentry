package tool

import (
	"context"
	"encoding/json"

	"github.com/marcodenic/agentry/internal/team"
)

// CollaborativeAgentTool provides collaborative capabilities to agents
type CollaborativeAgentTool struct {
	collabTool *team.CollaborativeTool
}

// NewCollaborativeAgentTool creates a new collaborative agent tool
func NewCollaborativeAgentTool(collabTool *team.CollaborativeTool) *CollaborativeAgentTool {
	return &CollaborativeAgentTool{
		collabTool: collabTool,
	}
}

// Call implements the Tool interface
func (cat *CollaborativeAgentTool) Call(ctx context.Context, input json.RawMessage) (any, error) {
	return cat.collabTool.Call(ctx, input)
}

// Description provides tool description
func (cat *CollaborativeAgentTool) Description() string {
	return `Collaborative tool for multi-agent communication and coordination.

Available actions:
- send_message: Send direct message to another agent
  Parameters: {"from": "agent_id", "to": "agent_id", "message": "text", "metadata": {}}

- request_collaboration: Request collaboration from another agent  
  Parameters: {"from": "agent_id", "to": "agent_id", "task": "description", "priority": "high/normal/low"}

- acquire_file_lock: Acquire a lock on a file to prevent conflicts
  Parameters: {"agent_id": "id", "file_path": "path", "lock_type": "read/write/exclusive", "duration_minutes": 30}

- release_file_lock: Release a file lock
  Parameters: {"agent_id": "id", "file_path": "path"}

- update_status: Update agent status and progress
  Parameters: {"agent_id": "id", "status": "working/idle/waiting/blocked", "current_task": "description", "progress": 0.5, "dependencies": ["agent1", "agent2"]}

- get_global_status: Get overall system status
  Parameters: {}

- get_team_status: Get status of all agents
  Parameters: {}

- request_help: Request help from the team
  Parameters: {"from": "agent_id", "help_type": "technical/review/guidance", "description": "what you need help with"}

Example usage:
{"action": "send_message", "parameters": {"from": "coder", "to": "tester", "message": "I've finished implementing the calculator. Please test it.", "metadata": {"file_path": "/tmp/calculator.py"}}}

{"action": "request_collaboration", "parameters": {"from": "coder", "to": "writer", "task": "Please write documentation for the calculator module", "priority": "normal"}}

{"action": "update_status", "parameters": {"agent_id": "coder", "status": "working", "current_task": "Implementing calculator functions", "progress": 0.7, "dependencies": []}}
`
}
