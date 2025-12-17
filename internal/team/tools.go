package team

import (
	"context"
	"errors"
	"strings"

	"github.com/marcodenic/agentry/internal/debug"
	"github.com/marcodenic/agentry/internal/tool"
)

// RegisterAgentTool registers the "agent" tool with the given tool registry.
// This is a simplified version that uses read-only sub-agents for search tasks.
func (t *Team) RegisterAgentTool(registry tool.Registry) {
	// Create the sub-agent tool using the parent agent's client
	subAgent := tool.NewSubAgentTool(t.parent.Client)
	registry["agent"] = subAgent
}

// GetAgentToolSpec returns the tool specification for the agent tool.
// This requires a Team to be present in the context at execution time.
func GetAgentToolSpec() tool.Tool {
	return tool.NewWithSchema(
		"agent",
		subAgentDescription(),
		subAgentSchema(),
		func(ctx context.Context, args map[string]any) (string, error) {
			teamInstance := TeamFromContext(ctx)
			if teamInstance == nil {
				return "", errors.New("no team found in context")
			}

			prompt := resolveStringArg(args, "prompt", "input", "task", "query", "instructions")
			if prompt == "" {
				return "", errors.New("prompt is required")
			}

			debug.Printf("SubAgent: Starting search task: %s", truncateStr(prompt, 100))

			subAgent := tool.NewSubAgentTool(teamInstance.parent.Client)
			return subAgent.Execute(ctx, map[string]any{"prompt": prompt})
		},
	)
}

// subAgentDescription provides the tool description for the sub-agent.
func subAgentDescription() string {
	return `Launch a read-only sub-agent to search the codebase. Use for parallel research tasks.

CAPABILITIES:
- Search files with glob patterns  
- Search file contents with grep
- Read file contents with view
- List directories

LIMITATIONS:
- Cannot modify files
- Cannot run shell commands  
- Cannot spawn further agents
- Stateless - each invocation starts fresh

WHEN TO USE:
- "Find all usages of function X" → spawn sub-agent
- "Which files handle authentication?" → spawn sub-agent
- Searching multiple patterns → spawn multiple sub-agents in parallel

WHEN NOT TO USE:
- Reading a specific known file → use view directly
- Modifying files → do it yourself
- Running commands → do it yourself

PARALLEL EXECUTION:
Launch multiple sub-agents concurrently by making multiple tool calls in one response.
Each sub-agent returns a text summary of its findings.`
}

// subAgentSchema returns the schema for the sub-agent tool.
func subAgentSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"prompt": map[string]any{
				"type":        "string",
				"description": "The search/research task for the sub-agent to perform",
			},
			// Keep legacy aliases for backward compatibility
			"input":        map[string]any{"type": "string", "description": "Alias for prompt"},
			"task":         map[string]any{"type": "string", "description": "Alias for prompt"},
			"query":        map[string]any{"type": "string", "description": "Alias for prompt"},
			"instructions": map[string]any{"type": "string", "description": "Alias for prompt"},
		},
		"required": []string{},
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

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
