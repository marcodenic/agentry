package tool

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// TeamContextKey is used to store a team implementation in context for builtins.
// The concrete value should implement the teamAPI interface below.
var TeamContextKey = struct{ key string }{"agentry.team"}

// AgentNameContextKey provides the current agent's logical name (e.g., "agent_0" or role name)
// when running within a Team. Team sets this value before invoking the agent.
var AgentNameContextKey = struct{ key string }{"agentry.agent-name"}

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
	// inbox
	GetAgentInbox(agentID string) []map[string]interface{}
	MarkMessagesAsRead(agentID string)
	// help
	RequestHelp(ctx context.Context, agentID, helpDescription string, preferredHelper string) error
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
				// Use current agent as sender if available; fallback to Agent 0
				from := "agent_0"
				if v := ctx.Value(AgentNameContextKey); v != nil {
					if s, ok := v.(string); ok && s != "" {
						from = s
					}
				}
				if err := t.SendMessageToAgent(ctx, from, to, message); err != nil {
					return "", err
				}
				return fmt.Sprintf("✅ Message sent from %s to %s [%s]: %s", from, to, msgType, message), nil
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

		// inbox_read: fetch unread messages for self or a specified agent; optionally mark as read
		"inbox_read": {
			Desc: "Read inbox messages for an agent (defaults to self)",
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"agent": map[string]any{
						"type":        "string",
						"description": "Agent name (defaults to current agent)",
					},
					"mark_read": map[string]any{
						"type":        "boolean",
						"description": "Mark returned messages as read",
						"default":     true,
					},
					"limit": map[string]any{
						"type":        "integer",
						"description": "Optional max number of messages to return",
						"default":     0,
					},
				},
				"required": []string{},
			},
			Exec: func(ctx context.Context, args map[string]any) (string, error) {
				tv := ctx.Value(TeamContextKey)
				t, _ := tv.(teamAPI)
				if t == nil {
					return "", fmt.Errorf("no team in context")
				}
				agent, _ := args["agent"].(string)
				if agent == "" {
					if v := ctx.Value(AgentNameContextKey); v != nil {
						if s, ok := v.(string); ok && s != "" {
							agent = s
						}
					}
				}
				if agent == "" {
					agent = "agent_0"
				}
				limit := 0
				switch vv := args["limit"].(type) {
				case float64:
					limit = int(vv)
				case int:
					limit = vv
				case string:
					if n, err := strconv.Atoi(vv); err == nil {
						limit = n
					}
				}
				markRead := true
				if mr, ok := args["mark_read"].(bool); ok {
					markRead = mr
				}

				msgs := t.GetAgentInbox(agent)
				// Filter unread
				unread := make([]map[string]interface{}, 0, len(msgs))
				for _, m := range msgs {
					if read, ok := m["read"].(bool); !ok || !read {
						unread = append(unread, m)
					}
				}
				if limit > 0 && len(unread) > limit {
					unread = unread[len(unread)-limit:]
				}
				if len(unread) == 0 {
					return "📭 Inbox empty", nil
				}
				if markRead {
					t.MarkMessagesAsRead(agent)
				}
				var b strings.Builder
				b.WriteString(fmt.Sprintf("📬 Inbox for %s (%d unread):\n", agent, len(unread)))
				for _, m := range unread {
					from, _ := m["from"].(string)
					msg, _ := m["message"].(string)
					ts := ""
					if tv, ok := m["timestamp"].(time.Time); ok {
						ts = tv.Format("15:04:05")
					}
					b.WriteString("- ")
					if ts != "" {
						b.WriteString("[")
						b.WriteString(ts)
						b.WriteString("] ")
					}
					b.WriteString(from)
					b.WriteString(": ")
					b.WriteString(msg)
					b.WriteString("\n")
				}
				return b.String(), nil
			},
		},

		// inbox_clear: clear all messages from an agent's inbox
		"inbox_clear": {
			Desc: "Clear all inbox messages for an agent (defaults to self)",
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"agent": map[string]any{
						"type":        "string",
						"description": "Agent name (defaults to current agent)",
					},
				},
				"required": []string{},
			},
			Exec: func(ctx context.Context, args map[string]any) (string, error) {
				tv := ctx.Value(TeamContextKey)
				t, _ := tv.(teamAPI)
				if t == nil {
					return "", fmt.Errorf("no team in context")
				}
				agent, _ := args["agent"].(string)
				if agent == "" {
					if v := ctx.Value(AgentNameContextKey); v != nil {
						if s, ok := v.(string); ok && s != "" {
							agent = s
						}
					}
				}
				if agent == "" {
					agent = "agent_0"
				}
				// Clear inbox by setting an empty slice
				inboxKey := fmt.Sprintf("inbox_%s", agent)
				t.SetSharedData(inboxKey, []map[string]interface{}{})
				return fmt.Sprintf("🧹 Cleared inbox for %s", agent), nil
			},
		},

		// workspace_events: show recent workspace events
		"workspace_events": {
			Desc: "List recent workspace events",
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"limit": map[string]any{
						"type":        "integer",
						"description": "Max events to show (0 = all)",
						"default":     10,
					},
				},
				"required": []string{},
			},
			Exec: func(ctx context.Context, args map[string]any) (string, error) {
				tv := ctx.Value(TeamContextKey)
				t, _ := tv.(teamAPI)
				if t == nil {
					return "", fmt.Errorf("no team in context")
				}
				limit := 10
				switch vv := args["limit"].(type) {
				case float64:
					limit = int(vv)
				case int:
					limit = vv
				case string:
					if n, err := strconv.Atoi(vv); err == nil {
						limit = n
					}
				}
				// Read raw events slice from shared memory; tolerate type differences
				raw, _ := t.GetSharedData("workspace_events")
				// Attempt to normalize to []map[string]interface{}
				var events []map[string]interface{}
				switch ev := raw.(type) {
				case []map[string]interface{}:
					events = ev
				case []interface{}:
					for _, it := range ev {
						if m, ok := it.(map[string]interface{}); ok {
							events = append(events, m)
						}
					}
				default:
					events = nil
				}
				if len(events) == 0 {
					return "📭 No workspace events", nil
				}
				var b strings.Builder
				b.WriteString("📡 Workspace Events:\n")
				// If limit > 0, take last limit
				start := 0
				if limit > 0 && len(events) > limit {
					start = len(events) - limit
				}
				for _, ev := range events[start:] {
					ts := ""
					if tsv, ok := ev["timestamp"].(time.Time); ok {
						ts = tsv.Format("15:04:05")
					} else if s, ok := ev["timestamp"].(string); ok {
						ts = s
					}
					agentID, _ := ev["agent_id"].(string)
					typ, _ := ev["type"].(string)
					desc, _ := ev["description"].(string)
					b.WriteString("- ")
					if ts != "" { b.WriteString("["); b.WriteString(ts); b.WriteString("] ") }
					if agentID != "" { b.WriteString(agentID); b.WriteString(" | ") }
					b.WriteString(typ)
					if desc != "" { b.WriteString(": "); b.WriteString(desc) }
					b.WriteString("\n")
				}
				return b.String(), nil
			},
		},

		// request_help: ask other agents for assistance via Team.RequestHelp
		"request_help": {
			Desc: "Request help from another agent (or broadcast)",
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"description": map[string]any{
						"type":        "string",
						"description": "What do you need help with?",
					},
					"preferred_helper": map[string]any{
						"type":        "string",
						"description": "Specific agent to ask (optional)",
						"default":     "",
					},
				},
				"required": []string{"description"},
			},
			Exec: func(ctx context.Context, args map[string]any) (string, error) {
				tv := ctx.Value(TeamContextKey)
				t, _ := tv.(teamAPI)
				if t == nil {
					return "", fmt.Errorf("no team in context")
				}
				desc, _ := args["description"].(string)
				helper, _ := args["preferred_helper"].(string)
				if desc == "" {
					return "", fmt.Errorf("description is required")
				}
				agent := "agent_0"
				if v := ctx.Value(AgentNameContextKey); v != nil {
					if s, ok := v.(string); ok && s != "" {
						agent = s
					}
				}
				if err := t.RequestHelp(ctx, agent, desc, helper); err != nil {
					return "", err
				}
				target := helper
				if target == "" { target = "all agents" }
				return fmt.Sprintf("🆘 Help requested from %s: %s", target, desc), nil
			},
		},
	}
}
