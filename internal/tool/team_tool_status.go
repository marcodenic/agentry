package tool

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/marcodenic/agentry/internal/contracts"
)

func teamStatusBuiltins() map[string]builtinSpec {
	return map[string]builtinSpec{
		"team_status":     teamStatusSpec(),
		"check_agent":     checkAgentSpec(),
		"available_roles": availableRolesSpec(),
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
			if strings.TrimSpace(agentName) == "" {
				return "", fmt.Errorf("agent name is required")
			}
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

func containsString(list []string, target string) bool {
	for _, item := range list {
		if item == target {
			return true
		}
	}
	return false
}
