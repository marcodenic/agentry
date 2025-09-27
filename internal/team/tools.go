package team

import (
	"context"
	"errors"
	"strings"

	"github.com/marcodenic/agentry/internal/tool"
)

// RegisterAgentTool registers the "agent" tool with the given tool registry.
// This must be called after creating the team to avoid import cycles.
func (t *Team) RegisterAgentTool(registry tool.Registry) {
	// Wrap delegation tool as terminal: its result is typically the final answer.
	registry["agent"] = tool.MarkTerminal(tool.NewWithSchema(
		"agent",
		"Delegate work to another agent",
		agentToolSchema(),
		agentDelegationExec(t),
	))

}

// GetAgentToolSpec returns the tool specification for the agent tool
// This can be used to register the tool without creating a team instance
func GetAgentToolSpec() tool.Tool {
	// Provide the same permissive schema and alias handling used by RegisterAgentTool.
	// Requires a Team in context at execution time.
	return tool.MarkTerminal(tool.NewWithSchema(
		"agent",
		"Delegate work to another agent",
		agentToolSchema(),
		agentDelegationExec(nil),
	))
}

// agentToolSchema returns a permissive schema accepting common alias keys.
func agentToolSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"agent": map[string]any{"type": "string", "description": "Name of the agent to delegate to"},
			"input": map[string]any{"type": "string", "description": "Task description or input for the agent"},
			// Accept common aliases the model may produce
			"role":         map[string]any{"type": "string", "description": "Alias for agent"},
			"task":         map[string]any{"type": "string", "description": "Alias for input"},
			"message":      map[string]any{"type": "string", "description": "Alias for input"},
			"query":        map[string]any{"type": "string", "description": "Alias for input"},
			"instructions": map[string]any{"type": "string", "description": "Alias for input"},
		},
		// Keep schema permissive; runtime resolves aliases
		"required": []string{},
	}
}

// agentDelegationExec returns the Exec function used by the agent delegation tool.
// If t is nil, the function requires a Team present in the context.
func agentDelegationExec(t *Team) func(ctx context.Context, args map[string]any) (string, error) {
	return func(ctx context.Context, args map[string]any) (string, error) {
		name := resolveStringArg(args, "agent", "role")
		if name == "" {
			return "", errors.New("agent name is required (use 'agent' or 'role')")
		}

		input := resolveStringArg(args, "input", "task", "message", "query", "instructions")
		if input == "" {
			return "", errors.New("input is required (use 'input', 'task', 'message', 'query', or 'instructions')")
		}

		// Use the team from the context if available, or use provided team instance
		var teamInstance *Team
		if contextTeam := TeamFromContext(ctx); contextTeam != nil {
			teamInstance = contextTeam
		} else {
			teamInstance = t
		}
		if teamInstance == nil {
			return "", errors.New("no team found in context")
		}
		return teamInstance.Call(ctx, name, input)
	}
}

// resolveStringArg returns the first non-empty string value for the provided keys.
func resolveStringArg(args map[string]any, primary string, aliases ...string) string {
	keys := append([]string{primary}, aliases...)
	for _, key := range keys {
		if v, ok := args[key]; ok {
			if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
				return s
			}
		}
	}
	return ""
}

// parallelAgentsToolSpec defines a reusable spec for executing multiple agents in parallel.
