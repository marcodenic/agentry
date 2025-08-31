package team

import (
    "context"
    "errors"

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

    // Add parallel agent tool via shared helper
    registry["parallel_agents"] = parallelAgentsToolSpec(t)
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
            "agent":       map[string]any{"type": "string", "description": "Name of the agent to delegate to"},
            "input":       map[string]any{"type": "string", "description": "Task description or input for the agent"},
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

// parallelAgentsToolSpec defines a reusable spec for executing multiple agents in parallel.
func parallelAgentsToolSpec(t *Team) tool.Tool {
    schema := map[string]any{
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
    return tool.NewWithSchema("parallel_agents", "Execute multiple agent tasks in parallel for efficiency", schema, func(ctx context.Context, args map[string]any) (string, error) {
        tasksInterface, ok := args["tasks"]
        if !ok {
            return "", errors.New("tasks array is required")
        }
        raw, ok := tasksInterface.([]interface{})
        if !ok {
            return "", errors.New("tasks must be an array")
        }
        // Normalize aliases
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
                if v, ok := m["task"].(string); ok && v != "" {
                    m["input"] = v
                }
            }
            raw[i] = m
        }
        var teamInstance *Team
        if contextTeam := TeamFromContext(ctx); contextTeam != nil {
            teamInstance = contextTeam
        } else {
            teamInstance = t
        }
        if teamInstance == nil {
            return "", errors.New("no team in context")
        }
        return teamInstance.CallParallel(ctx, raw)
    })
}
