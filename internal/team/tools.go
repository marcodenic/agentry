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
			"agent": map[string]any{
				"type": "string",
				"description": "Name of the agent to delegate to",
			},
			"input": map[string]any{
				"type": "string", 
				"description": "Task description or input for the agent",
			},
		},
		"required": []string{"agent", "input"},
	}
	
	registry["agent"] = tool.NewWithSchema("agent", "Delegate work to another agent", schema, func(ctx context.Context, args map[string]any) (string, error) {
		name, ok := args["agent"].(string)
		if !ok {
			return "", errors.New("agent name is required")
		}
		input, ok := args["input"].(string)
		if !ok {
			return "", errors.New("input is required")
		}
		
		// Use the team from the context if available, or use this team instance
		var teamInstance *Team
		if contextTeam := TeamFromContext(ctx); contextTeam != nil {
			teamInstance = contextTeam
		} else {
			teamInstance = t
		}
		
		return teamInstance.Call(ctx, name, input)
	})
}

// GetAgentToolSpec returns the tool specification for the agent tool
// This can be used to register the tool without creating a team instance
func GetAgentToolSpec() tool.Tool {
	return tool.New("agent", "Delegate work to another agent", func(ctx context.Context, args map[string]any) (string, error) {
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
	})
}
