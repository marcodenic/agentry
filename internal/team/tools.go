package team

import (
	"context"
	"errors"

	"github.com/marcodenic/agentry/internal/tool"
)

// RegisterAgentTool registers the "agent" tool with the given tool registry.
// This must be called after creating the team to avoid import cycles.
func (t *Team) RegisterAgentTool(registry tool.Registry) {
    // Register the agent tool normally; completion remains agent-driven.
    registry["agent"] = tool.NewWithSchema(
        "agent",
        "Delegate work to another agent",
        agentToolSchema(),
        agentDelegationExec(t),
    )
}

// GetAgentToolSpec returns the tool specification for the agent tool
// This can be used to register the tool without creating a team instance
func GetAgentToolSpec() tool.Tool {
    // Provide the same permissive schema and alias handling used by RegisterAgentTool.
    // Requires a Team in context at execution time.
    return tool.NewWithSchema(
        "agent",
        "Delegate work to another agent",
        agentToolSchema(),
        agentDelegationExec(nil),
    )
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
		// Only require one agent name and one input field - let runtime resolve aliases
		"required": []string{"agent", "input"},
	}
}

// agentDelegationExec returns the Exec function used by the agent delegation tool.
// If t is nil, the function requires a Team present in the context.
func agentDelegationExec(t *Team) func(ctx context.Context, args map[string]any) (string, error) {
	return func(ctx context.Context, args map[string]any) (string, error) {
		// Resolve aliases for agent name
		var name string
		if v, ok := args["agent"].(string); ok && v != "" {
			name = v
		} else if v, ok := args["role"].(string); ok && v != "" {
			name = v
		}
		if name == "" {
			return "", errors.New("agent name is required (use 'agent' or 'role')")
		}

		// Resolve aliases for input text
		var input string
		for _, k := range []string{"input", "task", "message", "query", "instructions"} {
			if v, ok := args[k].(string); ok && v != "" {
				input = v
				break
			}
		}
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
