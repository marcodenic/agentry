package tool

import (
	"context"
	"fmt"
)

// getTeamBuiltins returns team coordination builtin tools (placeholder implementations)
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
				// Placeholder implementation - will be replaced in buildAgent
				return "", fmt.Errorf("agent tool placeholder - should be replaced with proper implementation")
			},
		},
	}
}
