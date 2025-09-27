package tool

import (
	"context"
	"fmt"
)

// Removed deprecated teamAPI alias; use contracts.TeamService directly

// getTeamBuiltins returns team coordination builtin tools
func getTeamBuiltins() map[string]builtinSpec {
	specs := teamTools()
	specs["agent"] = builtinSpec{
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
		},
		Exec: func(context.Context, map[string]any) (string, error) {
			return "", fmt.Errorf("agent tool placeholder - should be replaced by team registration")
		},
	}
	return specs
}

func registerTeamBuiltins(reg *builtinRegistry) {
	reg.addAll(getTeamBuiltins())
}
