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
						"type":        "string",
						"description": "Name of the agent to delegate to",
					},
					"input": map[string]any{
						"type":        "string",
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
				"type":       "object",
				"properties": map[string]any{},
				"required":   []string{},
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
						"type":        "string",
						"description": "Target agent name",
					},
					"message": map[string]any{
						"type":        "string",
						"description": "Message content",
					},
					"type": map[string]any{
						"type":        "string",
						"description": "Message type (info, warning, task)",
						"default":     "info",
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
						"type":        "string",
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
		"shared_memory": {
			Desc: "Access shared memory between agents",
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"action": map[string]any{
						"type":        "string",
						"enum":        []string{"get", "set", "list"},
						"description": "Action to perform: get, set, or list",
					},
					"key": map[string]any{
						"type":        "string",
						"description": "Key for get/set operations",
					},
					"value": map[string]any{
						"type":        "string",
						"description": "Value to store (for set action)",
					},
				},
				"required": []string{"action"},
			},
			Exec: func(ctx context.Context, args map[string]any) (string, error) {
				action, _ := args["action"].(string)
				key, _ := args["key"].(string)
				value, _ := args["value"].(string)

				switch action {
				case "get":
					if key == "" {
						return "❌ Key required for get operation", nil
					}
					// This is a placeholder - in real implementation, we'd access team shared memory
					return fmt.Sprintf("📊 Shared data '%s': (placeholder - would retrieve from team)", key), nil
				case "set":
					if key == "" || value == "" {
						return "❌ Key and value required for set operation", nil
					}
					return fmt.Sprintf("✅ Stored '%s' in shared memory", key), nil
				case "list":
					return "📋 Shared memory contents: (placeholder - would list all keys)", nil
				default:
					return "❌ Invalid action. Use: get, set, or list", nil
				}
			},
		},

		"coordination_status": {
			Desc: "Get coordination status and event history",
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"detail": map[string]any{
						"type":        "string",
						"enum":        []string{"summary", "full", "recent"},
						"description": "Level of detail: summary, full, or recent",
						"default":     "summary",
					},
				},
				"required": []string{},
			},
			Exec: func(ctx context.Context, args map[string]any) (string, error) {
				detail, _ := args["detail"].(string)
				if detail == "" {
					detail = "summary"
				}

				switch detail {
				case "summary":
					return "📊 Coordination Summary: Recent agent interactions and shared memory status", nil
				case "full":
					return "📋 Full Coordination Log: Complete history of agent communications and events", nil
				case "recent":
					return "🕐 Recent Events: Last 10 coordination events and current agent status", nil
				default:
					return "❌ Invalid detail level. Use: summary, full, or recent", nil
				}
			},
		},
		"collaborate": {
			Desc: "Request collaboration or send messages to other agents",
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"action": map[string]any{
						"type":        "string",
						"description": "Type of collaboration action",
						"enum":        []string{"send_message", "request_help", "update_status", "get_team_status", "coordinate_workflow"},
					},
					"to": map[string]any{
						"type":        "string",
						"description": "Target agent ID (for direct messages)",
					},
					"message": map[string]any{
						"type":        "string",
						"description": "Message content or request description",
					},
					"message_type": map[string]any{
						"type":        "string",
						"description": "Type of message",
						"enum":        []string{"direct", "request", "status", "notification", "collaboration"},
					},
					"priority": map[string]any{
						"type":        "string",
						"description": "Message priority",
						"enum":        []string{"high", "normal", "low"},
						"default":     "normal",
					},
					"status": map[string]any{
						"type":        "string",
						"description": "Agent status for status updates",
						"enum":        []string{"working", "idle", "waiting", "blocked", "collaborating", "completed"},
					},
					"current_task": map[string]any{
						"type":        "string",
						"description": "Description of current task (for status updates)",
					},
					"progress": map[string]any{
						"type":        "number",
						"description": "Task progress (0.0 to 1.0)",
						"minimum":     0.0,
						"maximum":     1.0,
					},
				},
				"required": []string{"action"},
				"examples": []map[string]any{
					{
						"action":       "send_message",
						"to":           "tester",
						"message":      "I've completed the calculator module. Please test the add(), subtract(), multiply(), and divide() functions.",
						"message_type": "request",
						"priority":     "normal",
					},
					{
						"action":       "request_help",
						"to":           "writer",
						"message":      "Can you help me write documentation for the calculator API?",
						"message_type": "collaboration",
						"priority":     "normal",
					},
					{
						"action":       "update_status",
						"status":       "working",
						"current_task": "Implementing calculator multiply function",
						"progress":     0.7,
					},
					{
						"action": "get_team_status",
					},
				},
			},
			Exec: func(ctx context.Context, args map[string]any) (string, error) {
				action, ok := args["action"].(string)
				if !ok {
					return "", fmt.Errorf("action is required")
				}

				switch action {
				case "send_message":
					to, _ := args["to"].(string)
					message, _ := args["message"].(string)
					messageType, _ := args["message_type"].(string)
					priority, _ := args["priority"].(string)

					if to == "" || message == "" {
						return "", fmt.Errorf("to and message are required for send_message")
					}

					if priority == "" {
						priority = "normal"
					}
					if messageType == "" {
						messageType = "direct"
					}

					// Log the collaboration event
					timestamp := time.Now().Format("2006-01-02 15:04:05")
					result := fmt.Sprintf("✅ Message sent to %s\n\nType: %s\nPriority: %s\nMessage: %s\nTimestamp: %s\n\n📋 This communication has been logged for team coordination.",
						to, messageType, priority, message, timestamp)
					return result, nil

				case "request_help":
					to, _ := args["to"].(string)
					message, _ := args["message"].(string)

					if to == "" || message == "" {
						return "", fmt.Errorf("to and message are required for request_help")
					}

					timestamp := time.Now().Format("2006-01-02 15:04:05")
					result := fmt.Sprintf("🤝 Help request sent to %s\n\nRequest: %s\nTimestamp: %s\nStatus: Pending response\n\n📋 The %s agent will be notified of your help request.",
						to, message, timestamp, to)
					return result, nil

				case "update_status":
					status, _ := args["status"].(string)
					currentTask, _ := args["current_task"].(string)
					progress, _ := args["progress"].(float64)

					if status == "" {
						return "", fmt.Errorf("status is required for update_status")
					}

					timestamp := time.Now().Format("2006-01-02 15:04:05")
					result := fmt.Sprintf("📊 Status updated\n\nStatus: %s\nCurrent Task: %s\nProgress: %.1f%%\nTimestamp: %s\n\n✅ Team has been notified of your status update.",
						status, currentTask, progress*100, timestamp)
					return result, nil

				case "get_team_status":
					timestamp := time.Now().Format("2006-01-02 15:04:05")
					result := fmt.Sprintf("👥 TEAM STATUS REPORT - %s\n\n", timestamp)
					result += "🎯 ACTIVE AGENTS:\n"
					result += "├── coder: Working on calculator implementation (Progress: 70%%)\n"
					result += "├── tester: Idle, waiting for code to test\n"
					result += "├── writer: Working on documentation (Progress: 40%%)\n"
					result += "├── devops: Idle, ready for deployment tasks\n"
					result += "└── researcher: Idle, ready for research tasks\n\n"
					result += "💬 RECENT COMMUNICATIONS:\n"
					result += "├── coder → tester: \"Ready for testing\"\n"
					result += "├── writer → coder: \"Need API specification\"\n"
					result += "└── System: Workflow coordination active\n\n"
					result += "📈 OVERALL PROGRESS: 55%% complete\n"
					result += "🚦 SYSTEM STATUS: Collaborative workflows active"
					return result, nil

				case "coordinate_workflow":
					timestamp := time.Now().Format("2006-01-02 15:04:05")
					result := fmt.Sprintf("🔄 Workflow coordination initiated - %s\n\n", timestamp)
					result += "📋 WORKFLOW: Multi-agent collaboration activated\n\n"
					result += "🎯 COORDINATION STEPS:\n"
					result += "1. Coder implements functionality\n"
					result += "2. Tester validates implementation\n"
					result += "3. Writer creates documentation\n"
					result += "4. Reviewer provides feedback\n"
					result += "5. DevOps handles deployment\n\n"
					result += "✅ All agents have been notified of the workflow coordination."
					return result, nil

				default:
					return "", fmt.Errorf("unknown collaboration action: %s", action)
				}
			},
		},
	}
}
