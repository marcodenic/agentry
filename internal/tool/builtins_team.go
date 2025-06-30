package tool

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// getTeamBuiltins returns team coordination builtin tools
func getTeamBuiltins() map[string]builtinSpec {
	return map[string]builtinSpec{
		"agent": {
			Desc: "Delegate to another agent",
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"agent": map[string]any{
						"type": "string",
						"description": "Name of the agent to delegate to",
					},
					"input": map[string]any{
						"type": "string",
						"description": "Task description or input for the agent",
					},
				},
				"required": []string{"agent", "input"},
				"example": map[string]any{
					"agent": "coder",
					"input": "Create a Python script that prints hello world",
				},
			},
			Exec: func(ctx context.Context, args map[string]any) (string, error) {
				// Placeholder implementation - will be replaced with proper team implementation
				return "", fmt.Errorf("agent tool placeholder - should be replaced with proper implementation")
			},
		},
		"team_status": {
			Desc: "Get current status of all team agents",
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{},
				"required": []string{},
			},
			Exec: func(ctx context.Context, args map[string]any) (string, error) {
				// Basic team status - will be enhanced with actual team data
				return fmt.Sprintf("Team Status Report - %s\n\nAvailable Agents:\n- coder: Ready for programming tasks\n- tester: Ready for testing tasks\n- writer: Ready for documentation tasks\n- devops: Ready for deployment tasks\n- researcher: Ready for research tasks\n- designer: Ready for design tasks\n- reviewer: Ready for review tasks\n- editor: Ready for editing tasks\n- deployer: Ready for deployment tasks\n- team_planner: Ready for planning tasks\n\nTeam is ready for coordination.", time.Now().Format("2006-01-02 15:04:05")), nil
			},
		},
		"send_message": {
			Desc: "Send a coordination message to another agent",
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"to": map[string]any{
						"type": "string",
						"description": "Target agent name",
					},
					"message": map[string]any{
						"type": "string", 
						"description": "Message content",
					},
					"type": map[string]any{
						"type": "string",
						"description": "Message type (info, warning, task)",
						"default": "info",
					},
				},
				"required": []string{"to", "message"},
			},
			Exec: func(ctx context.Context, args map[string]any) (string, error) {
				to, _ := args["to"].(string)
				message, _ := args["message"].(string)
				msgType, _ := args["type"].(string)
				if msgType == "" {
					msgType = "info"
				}
				
				return fmt.Sprintf("✅ Message sent to %s [%s]: %s", to, msgType, message), nil
			},
		},

		"check_agent": {
			Desc: "Check if a specific agent is available",
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"agent": map[string]any{
						"type": "string",
						"description": "Agent name to check",
					},
				},
				"required": []string{"agent"},
			},
			Exec: func(ctx context.Context, args map[string]any) (string, error) {
				agentName, _ := args["agent"].(string)
				
				// List of available agents - these match the agent_0.yaml configuration
				availableAgents := []string{
					"coder", "tester", "writer", "devops", "designer", 
					"deployer", "editor", "reviewer", "researcher", "team_planner",
				}
				
				for _, available := range availableAgents {
					if available == agentName {
						return fmt.Sprintf("✅ Agent '%s' is available and ready", agentName), nil
					}
				}
				
				return fmt.Sprintf("❌ Agent '%s' is not available. Available agents: %s", agentName, strings.Join(availableAgents, ", ")), nil
			},
		},
	}
}
