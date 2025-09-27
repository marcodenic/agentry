package tool

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/marcodenic/agentry/internal/contracts"
)

func teamTools() map[string]builtinSpec {
	return map[string]builtinSpec{
		"team":                teamPlanSpec(),
		"team_status":         teamStatusSpec(),
		"check_agent":         checkAgentSpec(),
		"available_roles":     availableRolesSpec(),
		"shared_memory":       sharedMemorySpec(),
		"coordination_status": coordinationStatusSpec(),
		"workspace_events":    workspaceEventsSpec(),
	}
}

func teamPlanSpec() builtinSpec {
	return builtinSpec{
		Desc: "Coordinate a simple ad-hoc team for a topic (demo)",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"n":     map[string]any{"type": "integer", "description": "Number of helpers"},
				"topic": map[string]any{"type": "string", "description": "Topic to coordinate on"},
			},
			"required": []string{},
		},
		Exec: func(ctx context.Context, args map[string]any) (string, error) {
			n, _ := getIntArg(args, "n", 1)
			topic, _ := args["topic"].(string)
			if n < 1 {
				n = 1
			}
			var b strings.Builder
			b.WriteString("Team plan:\n")
			for i := 1; i <= n; i++ {
				b.WriteString(fmt.Sprintf("- agent_%d handles %s\n", i, topic))
			}
			return b.String(), nil
		},
	}
}

func teamStatusSpec() builtinSpec {
	return builtinSpec{
		Desc: "Get current status of all team agents",
		Schema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
			"required":   []string{},
		},
		Exec: func(ctx context.Context, args map[string]any) (string, error) {
			tv := ctx.Value(contracts.TeamContextKey)
			t, _ := tv.(contracts.TeamService)
			if t == nil {
				return "", fmt.Errorf("no team in context")
			}
			spawned := t.SpawnedAgentNames()
			var b strings.Builder
			b.WriteString(fmt.Sprintf("Team Status - %s\n", time.Now().Format("2006-01-02 15:04:05")))
			if len(spawned) == 0 {
				b.WriteString("No agents currently running.\n")
			} else {
				b.WriteString("Running Agents:\n")
				for _, n := range spawned {
					b.WriteString("- ")
					b.WriteString(n)
					b.WriteString("\n")
				}
			}
			b.WriteString("\nRecent Coordination:\n")
			lines := t.GetCoordinationHistory(5)
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
	}
}

func checkAgentSpec() builtinSpec {
	return builtinSpec{
		Desc: "Check if a specific agent is available",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"agent": map[string]any{"type": "string", "description": "Agent name to check"},
			},
			"required": []string{"agent"},
		},
		Exec: func(ctx context.Context, args map[string]any) (string, error) {
			agentName := stringArg(args, "agent")
			tv := ctx.Value(contracts.TeamContextKey)
			t, _ := tv.(contracts.TeamService)
			if t == nil {
				return "", fmt.Errorf("no team in context")
			}
			spawned := t.SpawnedAgentNames()
			for _, n := range spawned {
				if n == agentName {
					return fmt.Sprintf("✅ Agent '%s' is available", agentName), nil
				}
			}
			return fmt.Sprintf("❌ Agent '%s' is not available. Available agents: %s", agentName, strings.Join(spawned, ", ")), nil
		},
	}
}

func availableRolesSpec() builtinSpec {
	return builtinSpec{
		Desc: "List all available agent roles from configuration",
		Schema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
			"required":   []string{},
		},
		Exec: func(ctx context.Context, args map[string]any) (string, error) {
			tv := ctx.Value(contracts.TeamContextKey)
			t, _ := tv.(contracts.TeamService)
			if t == nil {
				return "", fmt.Errorf("no team in context")
			}
			available := t.AvailableRoleNames()
			spawned := t.SpawnedAgentNames()
			var b strings.Builder
			b.WriteString("🎭 Available Agent Roles:\n\n")
			if len(available) == 0 {
				b.WriteString("❌ No roles configured. Check your .agentry.yaml include paths.\n")
			} else {
				b.WriteString("📋 Configured Roles:\n")
				for _, role := range available {
					b.WriteString("- ")
					b.WriteString(role)
					if containsString(spawned, role) {
						b.WriteString(" (currently running)")
					}
					b.WriteString("\n")
				}
			}
			b.WriteString("\n💡 Use the 'agent' tool to delegate tasks to any of these roles.\n")
			b.WriteString("Example: {\"agent\": \"coder\", \"input\": \"create a hello world program\"}\n")
			return b.String(), nil
		},
	}
}

func sharedMemorySpec() builtinSpec {
	return builtinSpec{
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
					"description": "Value to set (for action 'set')",
				},
			},
			"required": []string{"action"},
		},
		Exec: sharedMemoryExec,
	}
}

func coordinationStatusSpec() builtinSpec {
	return builtinSpec{
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
		Exec: coordinationStatusExec,
	}
}

func coordinationStatusExec(ctx context.Context, args map[string]any) (string, error) {
	detail := stringArg(args, "detail")
	if detail == "" {
		detail = "summary"
	}
	tv := ctx.Value(contracts.TeamContextKey)
	t, _ := tv.(contracts.TeamService)
	if t == nil {
		return "", fmt.Errorf("no team in context")
	}
	switch detail {
	case "summary":
		return t.GetCoordinationSummary(), nil
	case "full":
		lines := t.GetCoordinationHistory(0)
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
		lines := t.GetCoordinationHistory(10)
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
}

func workspaceEventsSpec() builtinSpec {
	return builtinSpec{
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
		Exec: workspaceEventsExec,
	}
}

func workspaceEventsExec(ctx context.Context, args map[string]any) (string, error) {
	tv := ctx.Value(contracts.TeamContextKey)
	t, _ := tv.(contracts.TeamService)
	if t == nil {
		return "", fmt.Errorf("no team in context")
	}
	limit, _ := getIntArg(args, "limit", 10)
	raw, _ := t.GetSharedData("workspace_events")
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
	}
	if len(events) == 0 {
		return "📭 No workspace events", nil
	}
	if limit > 0 && len(events) > limit {
		events = events[len(events)-limit:]
	}
	var b strings.Builder
	b.WriteString("📡 Workspace Events:\n")
	for _, ev := range events {
		b.WriteString("- ")
		if ts, ok := ev["timestamp"].(time.Time); ok {
			b.WriteString(ts.Format("15:04:05"))
			b.WriteString(" ")
		}
		if agent, ok := ev["agent_id"].(string); ok && agent != "" {
			b.WriteString(agent)
			b.WriteString(" | ")
		}
		if desc, ok := ev["description"].(string); ok {
			b.WriteString(desc)
		}
		b.WriteString("\n")
	}
	return b.String(), nil
}

func containsString(list []string, target string) bool {
	for _, item := range list {
		if item == target {
			return true
		}
	}
	return false
}

func sharedMemoryExec(ctx context.Context, args map[string]any) (string, error) {
	action := stringArg(args, "action")
	key := stringArg(args, "key")
	value := stringArg(args, "value")
	tv := ctx.Value(contracts.TeamContextKey)
	t, _ := tv.(contracts.TeamService)
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
}
