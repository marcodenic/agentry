package core

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/marcodenic/agentry/internal/cost"
	"github.com/marcodenic/agentry/internal/debug"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/tool"
	"github.com/marcodenic/agentry/internal/trace"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Agent represents a conversational agent with LLM capabilities
type Agent struct {
	ID            uuid.UUID
	Client        model.Client
	ModelName     string
	Tools         tool.Registry
	Mem           memory.Store
	Vector        memory.VectorStore
	Vars          map[string]string
	Tracer        trace.Writer
	Cost          *cost.Manager
	Prompt        string
	// Error handling configuration
	ErrorHandling ErrorHandlingConfig
}

// ErrorHandlingConfig defines how the agent handles errors
type ErrorHandlingConfig struct {
	// TreatErrorsAsResults makes tool errors visible to the agent instead of terminating
	TreatErrorsAsResults bool
	// MaxErrorRetries limits how many consecutive errors an agent can handle
	MaxErrorRetries int
	// IncludeErrorContext adds detailed error context to help with recovery
	IncludeErrorContext bool
}

// DefaultErrorHandling returns sensible defaults for error handling
func DefaultErrorHandling() ErrorHandlingConfig {
	return ErrorHandlingConfig{
		TreatErrorsAsResults: true,
		MaxErrorRetries:      3,
		IncludeErrorContext:  true,
	}
}

var (
	tokenCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "agentry_tokens_total",
		Help: "Total tokens processed by an agent",
	}, []string{"agent"})
	toolLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "agentry_tool_latency_seconds",
		Help:    "Latency of tool execution in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"agent", "tool"})
)

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// getToolNames extracts tool names from a registry for debugging
func getToolNames(reg tool.Registry) []string {
	var names []string
	for name := range reg {
		names = append(names, name)
	}
	return names
}

func New(client model.Client, modelName string, reg tool.Registry, mem memory.Store, vec memory.VectorStore, tr trace.Writer) *Agent {
	return &Agent{
		ID:            uuid.New(),
		Tools:         reg,
		Mem:           mem,
		Vector:        vec,
		Client:        client,
		ModelName:     modelName,
		Tracer:        tr,
		Cost:          cost.New(0, 0.0), // Initialize cost manager immediately
		ErrorHandling: DefaultErrorHandling(),
	}
}

func (a *Agent) Spawn() *Agent {
	spawned := &Agent{
		ID:            uuid.New(),
		Prompt:        a.Prompt, // Ensure prompt is inherited by sub-agents
		Vars:          a.Vars,
		Tools:         a.Tools,
		Mem:           memory.NewInMemory(),
		Vector:        a.Vector,
		Client:        a.Client,
		ModelName:     a.ModelName,
		Tracer:        a.Tracer,
		Cost:          cost.New(0, 0.0), // Each spawned agent gets its own cost manager
		ErrorHandling: DefaultErrorHandling(),
	}

	debug.Printf("Agent.Spawn: Parent ID=%s, Spawned ID=%s", a.ID.String()[:8], spawned.ID.String()[:8])
	debug.Printf("Agent.Spawn: Inherited prompt length=%d chars", len(spawned.Prompt))
	debug.Printf("Agent.Spawn: Inherited prompt preview: %s", spawned.Prompt[:min(100, len(spawned.Prompt))])

	return spawned
}

func (a *Agent) Run(ctx context.Context, input string) (string, error) {
	debug.Printf("Agent.Run: Agent ID=%s, Prompt length=%d chars", a.ID.String()[:8], len(a.Prompt))
	debug.Printf("Agent.Run: Available tools: %v", getToolNames(a.Tools))
	debug.Printf("Agent.Run: Input: %s", input[:min(100, len(input))])

	a.Trace(ctx, trace.EventModelStart, a.ModelName)
	msgs := BuildMessages(a.Prompt, a.Vars, a.Mem.History(), input)
	specs := tool.BuildSpecs(a.Tools)

	debug.Printf("Agent.Run: Built %d messages, %d tool specs", len(msgs), len(specs))
	debug.Printf("Agent.Run: About to call model client with model %s", a.ModelName)

	// Do not estimate tokens here; rely on actual counts from responses

	// Track consecutive errors for resilience
	consecutiveErrors := 0

	for i := 0; ; i++ {
	// Note: No iteration cap; agent runs until it produces a final answer.
		debug.Printf("Agent.Run: Starting iteration %d", i)
		res, err := a.Client.Complete(ctx, msgs, specs)
		debug.Printf("Agent.Run: Model call completed for iteration %d, err=%v", i, err)
		if err != nil {
			debug.Printf("Agent.Run: Model call failed: %v", err)
			return "", err
		}

		debug.Printf("Agent.Run: Got response with %d tool calls", len(res.ToolCalls))

		a.Trace(ctx, trace.EventStepStart, res)

		// Use actual token counts from API response
		actualInputTokens := res.InputTokens
		actualOutputTokens := res.OutputTokens
		debug.Printf("Agent.Run: Iteration %d - Input tokens: %d, Output tokens: %d", i, actualInputTokens, actualOutputTokens)

		// Update metrics with actual tokens (count input+output per step)
		tokenCounter.WithLabelValues(a.ID.String()).Add(float64(actualInputTokens + actualOutputTokens))

		// Update cost manager with actual token usage
		if a.Cost != nil {
			if a.Cost.AddModelUsage(a.ModelName, actualInputTokens, actualOutputTokens) {
				debug.Printf("Agent.Run: Updated cost manager, total tokens now: %d", a.Cost.TotalTokens())
				if a.Cost.OverBudget() {
					debug.Printf("budget exceeded")
				}
			}
		}

		msgs = append(msgs, model.ChatMessage{Role: "assistant", Content: res.Content, ToolCalls: res.ToolCalls})
		step := memory.Step{Output: res.Content, ToolCalls: res.ToolCalls, ToolResults: map[string]string{}}
		if len(res.ToolCalls) == 0 {
			debug.Printf("Agent.Run: No tool calls, returning final result (length: %d)", len(res.Content))
			a.Mem.AddStep(step)
			_ = a.Checkpoint(ctx)

			// Emit token events for streaming effect with proper formatting preservation
			// Process character by character to preserve newlines and formatting
			for _, r := range res.Content {
				a.Trace(ctx, trace.EventToken, string(r))
				// No artificial delay - stream in real time as received
			}

			// Emit final message with the complete content for fallback
			a.Trace(ctx, trace.EventFinal, res.Content)
			debug.Printf("Agent.Run: Returning successfully with %d total tokens", func() int {
				if a.Cost != nil {
					return a.Cost.TotalTokens()
				}
				return 0
			}())
			return res.Content, nil
		}

		// Track if any tools in this step had errors
		stepHadErrors := false

		for _, tc := range res.ToolCalls {
			t, ok := a.Tools.Use(tc.Name)
			if !ok {
				errorMsg := fmt.Sprintf("Error: Unknown tool '%s'. Available tools: %v", tc.Name, getToolNames(a.Tools))

				if a.ErrorHandling.TreatErrorsAsResults {
					step.ToolResults[tc.ID] = errorMsg
					msgs = append(msgs, model.ChatMessage{Role: "tool", ToolCallID: tc.ID, Content: errorMsg})
					stepHadErrors = true
					continue
				} else {
					return "", fmt.Errorf("unknown tool: %s", tc.Name)
				}
			}
			var args map[string]any
			if err := json.Unmarshal(tc.Arguments, &args); err != nil {
				errorMsg := fmt.Sprintf("Error: Invalid tool arguments for '%s': %v", tc.Name, err)

				if a.ErrorHandling.TreatErrorsAsResults {
					step.ToolResults[tc.ID] = errorMsg
					msgs = append(msgs, model.ChatMessage{Role: "tool", ToolCallID: tc.ID, Content: errorMsg})
					stepHadErrors = true
					continue
				} else {
					return "", err
				}
			}
			applyVarsMap(args, a.Vars)

			// Debug: Log tool execution details for coder agent
			debug.Printf("Agent '%s' executing tool '%s' with args: %v", a.ID, tc.Name, args)

			a.Trace(ctx, trace.EventToolStart, map[string]any{"name": tc.Name, "args": args})
			start := time.Now()
			r, err := t.Execute(ctx, args)
			toolLatency.WithLabelValues(a.ID.String(), tc.Name).Observe(time.Since(start).Seconds())
			if err != nil {
				debug.Printf("Agent '%s' tool '%s' failed: %v", a.ID, tc.Name, err)

				// Format error message with context if enabled
				var errorMsg string
				if a.ErrorHandling.IncludeErrorContext {
					errorMsg = fmt.Sprintf("Error executing tool '%s': %v\n\nContext:\n- Tool: %s\n- Arguments: %v\n- Suggestion: Please try a different approach or check the tool usage.",
						tc.Name, err, tc.Name, args)
				} else {
					errorMsg = fmt.Sprintf("Error executing tool '%s': %v", tc.Name, err)
				}

				if a.ErrorHandling.TreatErrorsAsResults {
					step.ToolResults[tc.ID] = errorMsg
					msgs = append(msgs, model.ChatMessage{Role: "tool", ToolCallID: tc.ID, Content: errorMsg})
					stepHadErrors = true
					continue
				} else {
					return "", err
				}
			}

			debug.Printf("Agent '%s' tool '%s' succeeded, result length: %d", a.ID, tc.Name, len(r))

			// Tool output tokens are accounted for in subsequent model calls; avoid double-counting here
			a.Trace(ctx, trace.EventToolEnd, map[string]any{"name": tc.Name, "result": r})
			step.ToolResults[tc.ID] = r
			msgs = append(msgs, model.ChatMessage{Role: "tool", ToolCallID: tc.ID, Content: r})
		}

		// Handle error recovery logic
		if stepHadErrors {
			consecutiveErrors++
			if consecutiveErrors > a.ErrorHandling.MaxErrorRetries {
				return "", fmt.Errorf("too many consecutive errors (%d), stopping execution", consecutiveErrors)
			}
			debug.Printf("Agent '%s' had errors in step, continuing with error feedback (consecutive: %d/%d)",
				a.ID, consecutiveErrors, a.ErrorHandling.MaxErrorRetries)
		} else {
			consecutiveErrors = 0 // Reset counter on successful step
		}

		a.Mem.AddStep(step)
		_ = a.Checkpoint(ctx)
	}
}
