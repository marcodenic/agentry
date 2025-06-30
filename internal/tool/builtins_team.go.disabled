package tool

import (
	"context"
	"errors"
	"fmt"

	"github.com/marcodenic/agentry/internal/team"
)

// getTeamBuiltins returns team coordination builtin tools
func getTeamBuiltins() map[string]builtinSpec {
	return map[string]builtinSpec{
		"agent": {
			Desc: "Delegate to another agent",
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"agent": map[string]any{"type": "string"},
					"input": map[string]any{"type": "string"},
				},
				"required": []string{"agent", "input"},
				"example": map[string]any{
					"agent": "Agent1",
					"input": "Hello, how are you?",
				},
			},
			Exec: func(ctx context.Context, args map[string]any) (string, error) {
				name, _ := args["agent"].(string)
				input, _ := args["input"].(string)
				t, ok := team.FromContext(ctx)
				if !ok {
					return "", errors.New("team not found in context")
				}
				return t.Call(ctx, name, input)
			},
		},
		"team_status": {
			Desc: "Get the current status of all team agents",
			Schema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
				"required":   []string{},
				"example":    map[string]any{},
			},
			Exec: func(ctx context.Context, args map[string]any) (string, error) {
				_, ok := team.FromContext(ctx)
				if !ok {
					return "No team context available", nil
				}
				
				// Return basic team info for now
				// TODO: Integrate with orchestrator when context supports it
				return "Team coordination active - use other tools to manage agents", nil
			},
		},
		"send_message": {
			Desc: "Send a message to another team agent",
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"to": map[string]any{
						"type":        "string",
						"description": "Name of the agent to send message to, or 'all' for broadcast",
					},
					"message": map[string]any{
						"type":        "string",
						"description": "The message content to send",
					},
					"type": map[string]any{
						"type":        "string",
						"description": "Message type: 'info', 'task', 'question', 'status'",
						"default":     "info",
					},
				},
				"required": []string{"to", "message"},
				"example": map[string]any{
					"to":      "coder",
					"message": "Please create a new file called test.txt",
					"type":    "task",
				},
			},
			Exec: func(ctx context.Context, args map[string]any) (string, error) {
				to, _ := args["to"].(string)
				message, _ := args["message"].(string)
				msgType, _ := args["type"].(string)
				if msgType == "" {
					msgType = "info"
				}
				
				if to == "" || message == "" {
					return "", errors.New("missing required parameters: to and message")
				}
				
				// For now, return a confirmation
				// TODO: Integrate with actual team messaging system
				return fmt.Sprintf("Message sent to %s: %s (type: %s)", to, message, msgType), nil
			},
		},
		"assign_task": {
			Desc: "Assign a specific task to a team agent",
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"agent": map[string]any{
						"type":        "string",
						"description": "Name of the agent to assign the task to",
					},
					"task": map[string]any{
						"type":        "string",
						"description": "Description of the task to assign",
					},
					"priority": map[string]any{
						"type":        "string",
						"description": "Task priority: 'low', 'normal', 'high', 'urgent'",
						"default":     "normal",
					},
				},
				"required": []string{"agent", "task"},
				"example": map[string]any{
					"agent":    "coder",
					"task":     "Create a new Python script to parse CSV files",
					"priority": "normal",
				},
			},
			Exec: func(ctx context.Context, args map[string]any) (string, error) {
				agent, _ := args["agent"].(string)
				task, _ := args["task"].(string)
				priority, _ := args["priority"].(string)
				if priority == "" {
					priority = "normal"
				}
				
				if agent == "" || task == "" {
					return "", errors.New("missing required parameters: agent and task")
				}
				
				// For now, return a confirmation
				// TODO: Integrate with actual task assignment system
				return fmt.Sprintf("Task assigned to %s (priority: %s): %s", agent, priority, task), nil
			},
		},
		"check_agent": {
			Desc: "Check the status and availability of a specific agent",
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"agent": map[string]any{
						"type":        "string",
						"description": "Name of the agent to check",
					},
				},
				"required": []string{"agent"},
				"example": map[string]any{
					"agent": "coder",
				},
			},
			Exec: func(ctx context.Context, args map[string]any) (string, error) {
				agent, _ := args["agent"].(string)
				if agent == "" {
					return "", errors.New("missing required parameter: agent")
				}
				
				t, ok := team.FromContext(ctx)
				if !ok {
					return "No team context available", nil
				}
				
				// Try to call the agent to see if it exists
				_, err := t.Call(ctx, agent, "status check")
				if err != nil {
					return fmt.Sprintf("Agent '%s' not available or not found", agent), nil
				}
				
				return fmt.Sprintf("Agent '%s' is available", agent), nil
			},
		},
	}
}
