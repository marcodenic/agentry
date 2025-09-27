package tool

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/marcodenic/agentry/internal/contracts"
)

func coordinationBuiltins() map[string]builtinSpec {
	return map[string]builtinSpec{
		"coordination_status": coordinationStatusSpec(),
		"workspace_events":    workspaceEventsSpec(),
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
		return "", fmt.Errorf("invalid detail level. Use: summary, full, or recent")
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
