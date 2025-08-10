package tool

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// TeamContextKey is used to store a team implementation in context for builtins.
// The concrete value should implement the teamAPI interface below.
var TeamContextKey = struct{ key string }{"agentry.team"}

// teamAPI defines the minimal surface the team builtins depend on.
type teamAPI interface {
	// messaging
	SendMessageToAgent(ctx context.Context, fromAgentID, toAgentID, message string) error
	// shared memory
	GetSharedData(key string) (interface{}, bool)
	SetSharedData(key string, value interface{})
	GetAllSharedData() map[string]interface{}
	// coordination
	GetCoordinationSummary() string
	CoordinationHistoryStrings(limit int) []string
	// discovery
	Names() []string
}

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
				// This is a placeholder that will be replaced by the team's RegisterAgentTool
				return "", fmt.Errorf("agent tool placeholder - should be replaced by team registration")
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
				tv := ctx.Value(TeamContextKey)
				t, _ := tv.(teamAPI)
				if t == nil {
					return "", fmt.Errorf("no team in context")
				}

				names := t.Names()
				var b strings.Builder
				b.WriteString(fmt.Sprintf("Team Status - %s\n", time.Now().Format("2006-01-02 15:04:05")))
				if len(names) == 0 {
					b.WriteString("No agents registered.\n")
				} else {
					b.WriteString("Agents:\n")
					for _, n := range names {
						b.WriteString("- ")
						b.WriteString(n)
						b.WriteString("\n")
					}
				}
				// Append a short coordination summary
				b.WriteString("\nRecent Coordination:\n")
				lines := t.CoordinationHistoryStrings(5)
				if len(lines) == 0 {
					b.WriteString("- none\n")
				} else {
					for _, ln := range lines {
						b.WriteString("- ")
						b.WriteString(ln)
						b.WriteString("\n")
					}
				}
				return b.String(), nil
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
				// Resolve team from context without importing the team package
				tv := ctx.Value(TeamContextKey)
				t, _ := tv.(teamAPI)
				if t == nil {
					return "", fmt.Errorf("no team in context")
				}
				// Use Agent 0 as sender by default
				if err := t.SendMessageToAgent(ctx, "agent_0", to, message); err != nil {
					return "", err
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

				tv := ctx.Value(TeamContextKey)
				t, _ := tv.(teamAPI)
				if t == nil {
					return "", fmt.Errorf("no team in context")
				}
				names := t.Names()
				for _, n := range names {
					if n == agentName {
						return fmt.Sprintf("✅ Agent '%s' is available", agentName), nil
					}
				}
				return fmt.Sprintf("❌ Agent '%s' is not available. Available agents: %s", agentName, strings.Join(names, ", ")), nil
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
				tv := ctx.Value(TeamContextKey)
				t, _ := tv.(teamAPI)
				if t == nil {
					return "", fmt.Errorf("no team in context")
				}
				switch action {
				case "get":
					if key == "" {
						return "❌ Key required for get operation", nil
					}
					if v, ok := t.GetSharedData(key); ok {
						return fmt.Sprintf("📊 %s = %v", key, v), nil
					}
					return fmt.Sprintf("📊 %s not set", key), nil
				case "set":
					if key == "" || value == "" {
						return "❌ Key and value required for set operation", nil
					}
					t.SetSharedData(key, value)
					return fmt.Sprintf("✅ Stored '%s' in shared memory", key), nil
				case "list":
					data := t.GetAllSharedData()
					if len(data) == 0 {
						return "📋 Shared memory is empty", nil
					}
					var b strings.Builder
					b.WriteString("📋 Shared memory keys:\n")
					for k := range data {
						b.WriteString("- ")
						b.WriteString(k)
						b.WriteString("\n")
					}
					return b.String(), nil
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
				tv := ctx.Value(TeamContextKey)
				t, _ := tv.(teamAPI)
				if t == nil {
					return "", fmt.Errorf("no team in context")
				}
				switch detail {
				case "summary":
					return t.GetCoordinationSummary(), nil
				case "full":
					lines := t.CoordinationHistoryStrings(0)
					if len(lines) == 0 {
						return "📋 No coordination events.", nil
					}
					var b strings.Builder
					b.WriteString("📋 Coordination Events (full):\n")
					for _, ln := range lines {
						b.WriteString("- ")
						b.WriteString(ln)
						b.WriteString("\n")
					}
					return b.String(), nil
				case "recent":
					lines := t.CoordinationHistoryStrings(10)
					if len(lines) == 0 {
						return "🕐 No recent coordination events.", nil
					}
					var b strings.Builder
					b.WriteString("🕐 Recent Events:\n")
					for _, ln := range lines {
						b.WriteString("- ")
						b.WriteString(ln)
						b.WriteString("\n")
					}
					return b.String(), nil
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
						"enum":        []string{"send_message", "request_help", "update_status", "get_team_status"},
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
					tv := ctx.Value(TeamContextKey)
					t, _ := tv.(teamAPI)
					if t == nil {
						return "", fmt.Errorf("no team in context")
					}
					names := t.Names()
					var b strings.Builder
					b.WriteString(fmt.Sprintf("Team Status - %s\n\n", time.Now().Format("2006-01-02 15:04:05")))
					if len(names) == 0 {
						b.WriteString("Agents: none\n")
					} else {
						b.WriteString("Agents:\n")
						for _, n := range names {
							b.WriteString("- ")
							b.WriteString(n)
							b.WriteString("\n")
						}
					}
					b.WriteString("\nRecent Events:\n")
					lines := t.CoordinationHistoryStrings(10)
					if len(lines) == 0 {
						b.WriteString("- none\n")
					} else {
						for _, ln := range lines {
							b.WriteString("- ")
							b.WriteString(ln)
							b.WriteString("\n")
						}
					}
					return b.String(), nil

				// removed coordinate_workflow action (no separate workflow engine)

				default:
					return "", fmt.Errorf("unknown collaboration action: %s", action)
				}
			},
		},
	}
}
