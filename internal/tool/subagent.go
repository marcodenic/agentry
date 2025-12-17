// Package tool provides the sub-agent tool for parallel read-only operations.
package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/marcodenic/agentry/internal/debug"
	"github.com/marcodenic/agentry/internal/model"
)

// SubAgentTool spawns ephemeral read-only agents for parallel search operations.
// Unlike the old delegation model, sub-agents:
// - Are stateless (no memory between invocations)
// - Have read-only tools only (glob, grep, view, ls)
// - Can run in parallel
// - Return simple text results
type SubAgentTool struct {
	// clientFactory creates fresh clients for each sub-agent invocation
	// to avoid sharing conversation state with the parent agent
	clientFactory func() model.Client
	tools         Registry
}

// ReadOnlyBuiltins lists the tools available to sub-agents.
// These are safe, read-only operations that cannot modify the filesystem.
var ReadOnlyBuiltins = []string{"glob", "grep", "view", "ls"}

// NewSubAgentTool creates a new sub-agent tool with read-only capabilities.
// It uses the provided client to determine model provider/name and creates
// fresh clients for each invocation to avoid state sharing.
func NewSubAgentTool(client model.Client) Tool {
	// Create a registry with only read-only tools from builtinMap
	readOnlyTools := Registry{}

	for _, name := range ReadOnlyBuiltins {
		if spec, ok := builtinMap[name]; ok {
			readOnlyTools[name] = NewWithSchema(name, spec.Desc, spec.Schema, spec.Exec)
		}
	}

	// Determine the client factory based on the client type
	// We create fresh clients to avoid sharing previousResponseID state
	var clientFactory func() model.Client
	switch c := client.(type) {
	case *model.OpenAI:
		// Get the key from environment since we can't access it directly
		key := os.Getenv("OPENAI_API_KEY")
		modelName := c.ModelName()
		clientFactory = func() model.Client {
			return model.NewOpenAI(key, modelName)
		}
	case *model.Anthropic:
		key := os.Getenv("ANTHROPIC_API_KEY")
		modelName := c.ModelName()
		clientFactory = func() model.Client {
			return model.NewAnthropic(key, modelName)
		}
	default:
		// Fallback: reuse the client (may cause issues with stateful APIs)
		clientFactory = func() model.Client { return client }
	}

	sat := &SubAgentTool{
		clientFactory: clientFactory,
		tools:         readOnlyTools,
	}

	return sat
}

func (sat *SubAgentTool) Name() string { return "agent" }

func (sat *SubAgentTool) Description() string {
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

func (sat *SubAgentTool) JSONSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"prompt": map[string]any{
				"type":        "string",
				"description": "The search/research task for the sub-agent to perform",
			},
		},
		"required": []string{"prompt"},
	}
}

func (sat *SubAgentTool) Execute(ctx context.Context, args map[string]any) (string, error) {
	prompt, _ := args["prompt"].(string)
	if strings.TrimSpace(prompt) == "" {
		// Try legacy aliases
		for _, key := range []string{"input", "task", "query", "instructions"} {
			if v, ok := args[key].(string); ok && strings.TrimSpace(v) != "" {
				prompt = v
				break
			}
		}
	}

	if strings.TrimSpace(prompt) == "" {
		return "", fmt.Errorf("prompt is required")
	}

	return sat.Run(ctx, prompt)
}

// Run executes a sub-agent with the given prompt.
func (sat *SubAgentTool) Run(ctx context.Context, prompt string) (string, error) {
	debug.Printf("SubAgent: Starting with prompt: %s", truncate(prompt, 100))

	// Create a fresh client for this sub-agent invocation
	// to avoid sharing conversation state with the parent agent
	client := sat.clientFactory()

	specs := BuildSpecs(sat.tools)

	systemPrompt := `You are a search agent. Your task is to find information in the codebase.

RULES:
1. Use the available tools to search for the requested information
2. Be concise and direct in your response
3. Return only the relevant findings
4. If you can't find what was requested, say so clearly

AVAILABLE TOOLS: glob (find files), grep (search contents), view (read files), ls (list dirs)

After finding the information, provide a clear summary of your findings.`

	msgs := []model.ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: prompt},
	}

	// Simple conversation loop with max iterations
	maxIter := 10
	for i := 0; i < maxIter; i++ {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		ch, err := client.Stream(ctx, msgs, specs)
		if err != nil {
			return "", fmt.Errorf("sub-agent stream error: %w", err)
		}

		completion := aggregateStreamChunks(ch)

		// Add assistant response to messages
		msgs = append(msgs, model.ChatMessage{
			Role:      "assistant",
			Content:   completion.Content,
			ToolCalls: completion.ToolCalls,
		})

		// If no tool calls, we're done
		if len(completion.ToolCalls) == 0 {
			debug.Printf("SubAgent: Completed with response length: %d", len(completion.Content))
			return completion.Content, nil
		}

		// Execute tool calls in parallel
		results := sat.executeToolCalls(ctx, completion.ToolCalls)

		// Add tool results to messages
		for _, result := range results {
			msgs = append(msgs, model.ChatMessage{
				Role:       "tool",
				Content:    result.content,
				ToolCallID: result.id,
				Name:       result.name,
			})
		}
	}

	return "", fmt.Errorf("sub-agent reached maximum iterations (%d)", maxIter)
}

// aggregateStreamChunks collects all chunks from a stream into a Completion.
func aggregateStreamChunks(ch <-chan model.StreamChunk) model.Completion {
	var content strings.Builder
	var toolCalls []model.ToolCall
	var inputTokens, outputTokens int
	var modelName string

	for chunk := range ch {
		if chunk.Err != nil {
			continue
		}
		content.WriteString(chunk.ContentDelta)
		// Tool calls come in the final Done chunk
		if chunk.Done {
			toolCalls = chunk.ToolCalls
			if chunk.InputTokens > 0 {
				inputTokens = chunk.InputTokens
			}
			if chunk.OutputTokens > 0 {
				outputTokens = chunk.OutputTokens
			}
			if chunk.ModelName != "" {
				modelName = chunk.ModelName
			}
		}
	}

	return model.Completion{
		Content:      content.String(),
		ToolCalls:    toolCalls,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		ModelName:    modelName,
	}
}

type toolResult struct {
	id      string
	name    string
	content string
}

func (sat *SubAgentTool) executeToolCalls(ctx context.Context, calls []model.ToolCall) []toolResult {
	results := make([]toolResult, len(calls))
	var wg sync.WaitGroup

	for i, tc := range calls {
		wg.Add(1)
		go func(idx int, call model.ToolCall) {
			defer wg.Done()

			result := toolResult{
				id:   call.ID,
				name: call.Name,
			}

			tool, ok := sat.tools[call.Name]
			if !ok {
				result.content = fmt.Sprintf("Error: unknown tool '%s'. Available: glob, grep, view, ls", call.Name)
				results[idx] = result
				return
			}

			// Parse the JSON arguments
			var args map[string]any
			if err := json.Unmarshal(call.Arguments, &args); err != nil {
				result.content = fmt.Sprintf("Error parsing arguments: %v", err)
				results[idx] = result
				return
			}

			output, err := tool.Execute(ctx, args)
			if err != nil {
				result.content = fmt.Sprintf("Error: %v", err)
			} else {
				// Truncate large outputs
				if len(output) > 10000 {
					output = output[:10000] + "\n... [output truncated]"
				}
				result.content = output
			}
			results[idx] = result
		}(i, tc)
	}

	wg.Wait()
	return results
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
