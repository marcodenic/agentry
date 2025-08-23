package team

import (
	"context"
	"errors"

	"github.com/marcodenic/agentry/internal/tool"
)

// RegisterAgentTool registers the "agent" tool with the given tool registry.
// This must be called after creating the team to avoid import cycles.
func (t *Team) RegisterAgentTool(registry tool.Registry) {
	schema := map[string]any{
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
		"required": []string{}, // keep schema permissive; we validate at runtime with aliases
	}

	// Wrap delegation tool as terminal: its result is typically the final answer.
	registry["agent"] = tool.MarkTerminal(tool.NewWithSchema("agent", "Delegate work to another agent", schema, func(ctx context.Context, args map[string]any) (string, error) {
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

		// Use the team from the context if available, or use this team instance
		var teamInstance *Team
		if contextTeam := TeamFromContext(ctx); contextTeam != nil {
			teamInstance = contextTeam
		} else {
			teamInstance = t
		}

		return teamInstance.Call(ctx, name, input)
	}))

	// Add parallel agent tool for executing multiple agents simultaneously
	parallelSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"tasks": map[string]any{
				"type":        "array",
				"description": "Array of agent tasks to execute in parallel",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"agent": map[string]any{"type": "string", "description": "Name of the agent to delegate to"},
						"input": map[string]any{"type": "string", "description": "Task description or input for the agent"},
						"role":  map[string]any{"type": "string", "description": "Alias for agent"},
						"task":  map[string]any{"type": "string", "description": "Alias for input"},
					},
					"required": []string{},
				},
			},
		},
		"required": []string{"tasks"},
	}

	registry["parallel_agents"] = tool.NewWithSchema("parallel_agents", "Execute multiple agent tasks in parallel for efficiency", parallelSchema, func(ctx context.Context, args map[string]any) (string, error) {
		tasksInterface, ok := args["tasks"]
		if !ok {
			return "", errors.New("tasks array is required")
		}

		raw, ok := tasksInterface.([]interface{})
		if !ok {
			return "", errors.New("tasks must be an array")
		}

		// Convert alias keys for each task item
		for i, item := range raw {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			if _, has := m["agent"]; !has {
				if v, ok := m["role"].(string); ok && v != "" {
					m["agent"] = v
				}
			}
			if _, has := m["input"]; !has {
				for _, k := range []string{"task"} {
					if v, ok := m[k].(string); ok && v != "" {
						m["input"] = v
						break
					}
				}
			}
			raw[i] = m
		}

		// Use team from context if available
		var teamInstance *Team
		if contextTeam := TeamFromContext(ctx); contextTeam != nil {
			teamInstance = contextTeam
		} else {
			teamInstance = t
		}

		return teamInstance.CallParallel(ctx, raw)
	})
}

// GetAgentToolSpec returns the tool specification for the agent tool
// This can be used to register the tool without creating a team instance
func GetAgentToolSpec() tool.Tool {
	return tool.MarkTerminal(tool.New("agent", "Delegate work to another agent", func(ctx context.Context, args map[string]any) (string, error) {
		name, ok := args["agent"].(string)
		if !ok {
			return "", errors.New("agent name is required")
		}
		input, ok := args["input"].(string)
		if !ok {
			return "", errors.New("input is required")
		}

		// Get the team from context
		teamInstance := TeamFromContext(ctx)
		if teamInstance == nil {
			return "", errors.New("no team found in context")
		}

		return teamInstance.Call(ctx, name, input)
	}))
}
