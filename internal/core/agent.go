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
	"github.com/marcodenic/agentry/internal/env"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/tool"
	"github.com/marcodenic/agentry/internal/trace"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Agent represents a conversational agent with LLM capabilities
type Agent struct {
	ID        uuid.UUID
	Client    model.Client
	ModelName string
	Tools     tool.Registry
	Mem       memory.Store
	Vector    memory.VectorStore
	Vars      map[string]string // cloned on spawn to avoid shared mutation
	Tracer    trace.Writer
	Cost      *cost.Manager
	Prompt    string
	// Error handling configuration
	ErrorHandling ErrorHandlingConfig

	// cached tool names to reduce repeated map iteration/log noise
	cachedToolNames []string
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
	modelLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "agentry_model_latency_seconds",
		Help:    "Latency of model completion calls in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"agent", "model"})
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

// toolNames returns cached tool names for this agent (compute once)
func (a *Agent) toolNames() []string {
	if a.cachedToolNames != nil {
		return a.cachedToolNames
	}
	a.cachedToolNames = getToolNames(a.Tools)
	return a.cachedToolNames
}

func New(client model.Client, modelName string, reg tool.Registry, mem memory.Store, vec memory.VectorStore, tr trace.Writer) *Agent {
	budgetTokens := env.Int("AGENTRY_BUDGET_TOKENS", 0)
	budgetDollars := env.Float("AGENTRY_BUDGET_DOLLARS", 0)
	return &Agent{
		ID:            uuid.New(),
		Tools:         reg,
		Mem:           mem,
		Vector:        vec,
		Client:        client,
		ModelName:     modelName,
		Tracer:        tr,
		Cost:          cost.New(budgetTokens, budgetDollars),
		ErrorHandling: DefaultErrorHandling(),
	}
}

func (a *Agent) Spawn() *Agent {
	// clone vars map to prevent parent/child accidental shared mutation
	var clonedVars map[string]string
	if a.Vars != nil {
		clonedVars = make(map[string]string, len(a.Vars))
		for k, v := range a.Vars {
			clonedVars[k] = v
		}
	}
	spawned := &Agent{
		ID:              uuid.New(),
		Prompt:          a.Prompt, // inherit prompt
		Vars:            clonedVars,
		Tools:           a.Tools,
		Mem:             memory.NewInMemory(),
		Vector:          a.Vector, // share vector store intentionally (semantic memory)
		Client:          a.Client,
		ModelName:       a.ModelName,
		Tracer:          a.Tracer,
		Cost:            cost.New(a.Cost.BudgetTokens, a.Cost.BudgetDollars),
		ErrorHandling:   DefaultErrorHandling(),
		cachedToolNames: a.cachedToolNames, // reuse already computed list if present
	}
	debug.Printf("Agent.Spawn: Parent ID=%s, Spawned ID=%s", a.ID.String()[:8], spawned.ID.String()[:8])
	debug.Printf("Agent.Spawn: Inherited prompt length=%d chars", len(spawned.Prompt))
	if l := len(spawned.Prompt); l > 0 {
		debug.Printf("Agent.Spawn: Inherited prompt preview: %s", spawned.Prompt[:min(100, l)])
	}
	return spawned
}

func (a *Agent) Run(ctx context.Context, input string) (string, error) {
	debug.Printf("Agent.Run: Agent ID=%s, Prompt length=%d chars", a.ID.String()[:8], len(a.Prompt))
	debug.Printf("Agent.Run: Available tools: %v", a.toolNames())
	debug.Printf("Agent.Run: Input: %s", input[:min(100, len(input))])

	a.Trace(ctx, trace.EventModelStart, a.ModelName)
	msgs := BuildMessages(a.Prompt, a.Vars, a.Mem.History(), input)
	specs := tool.BuildSpecs(a.Tools)

	debug.Printf("Agent.Run: Built %d messages, %d tool specs", len(msgs), len(specs))
	debug.Printf("Agent.Run: About to call model client with model %s", a.ModelName)

	// DEBUG: Print the full system prompt that will be sent to the API
	if len(msgs) > 0 && msgs[0].Role == "system" {
		debug.Printf("=== FULL SYSTEM PROMPT BEING SENT TO API ===")
		debug.Printf("%s", msgs[0].Content)
		debug.Printf("=== END SYSTEM PROMPT ===")
	}

	// DEBUG: Print available tool specs
	debug.Printf("=== AVAILABLE TOOLS ===")
	for _, spec := range specs {
		debug.Printf("Tool: %s", spec.Name)
	}
	debug.Printf("=== END TOOLS ===")

	// DEBUG: Print available tool names from registry
	debug.Printf("=== AVAILABLE TOOL NAMES ===")
	for _, name := range a.toolNames() {
		debug.Printf("Tool: %s", name)
	}
	debug.Printf("=== END TOOL NAMES ===")

	// Do not estimate tokens here; rely on actual counts from responses

	// Track consecutive errors for resilience
	consecutiveErrors := 0

	// optional iteration cap via env var (safety) "AGENTRY_MAX_ITER"
	maxIter := env.Int("AGENTRY_MAX_ITER", 0)
	for i := 0; ; i++ {
		if maxIter > 0 && i >= maxIter {
			return "", fmt.Errorf("iteration cap reached (%d)", maxIter)
		}
		// cancellation check early in loop
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}
		// Note: No iteration cap; agent runs until it produces a final answer.
		debug.Printf("Agent.Run: Starting iteration %d", i)
		startModel := time.Now()
		// If client supports streaming, prefer that path for first token latency
		if sc, ok := a.Client.(model.StreamingClient); ok {
			streamCh, sErr := sc.Stream(ctx, msgs, specs)
			if sErr == nil && streamCh != nil {
				var assembled string
				var finalToolCalls []model.ToolCall
				for chunk := range streamCh {
					if chunk.Err != nil {
						return "", chunk.Err
					}
					if chunk.ContentDelta != "" {
						assembled += chunk.ContentDelta
						a.Trace(ctx, trace.EventToken, chunk.ContentDelta)
					}
					if chunk.Done {
						finalToolCalls = chunk.ToolCalls
					}
				}
				modelLatency.WithLabelValues(a.ID.String(), a.ModelName).Observe(time.Since(startModel).Seconds())
				// After streaming, treat result as a single completion
				res := model.Completion{Content: assembled, ToolCalls: finalToolCalls}
				debug.Printf("Agent.Run: Streaming completed with %d tool calls", len(res.ToolCalls))
				a.Trace(ctx, trace.EventStepStart, res)
				// Append assistant message
				msgs = append(msgs, model.ChatMessage{Role: "assistant", Content: res.Content, ToolCalls: res.ToolCalls})
				step := memory.Step{Output: res.Content, ToolCalls: res.ToolCalls, ToolResults: map[string]string{}}
				if len(res.ToolCalls) == 0 {
					a.Mem.AddStep(step)
					_ = a.Checkpoint(ctx)
					a.Trace(ctx, trace.EventFinal, res.Content)
					return res.Content, nil
				}
				// Execute tools then continue loop with new messages
				toolMsgs, hadErrors, execErr := a.executeToolCalls(ctx, res.ToolCalls, step)
				if execErr != nil { return "", execErr }
				msgs = append(msgs, toolMsgs...)
				a.Mem.AddStep(step)
				_ = a.Checkpoint(ctx)
				if hadErrors { consecutiveErrors++ } else { consecutiveErrors = 0 }
				if consecutiveErrors > a.ErrorHandling.MaxErrorRetries {
					return "", fmt.Errorf("too many consecutive errors (%d), stopping execution", consecutiveErrors)
				}
				// Continue outer for-loop for next iteration
				continue
			}
		}
		res, err := a.Client.Complete(ctx, msgs, specs)
		modelLatency.WithLabelValues(a.ID.String(), a.ModelName).Observe(time.Since(startModel).Seconds())
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
					if env.Bool("AGENTRY_STOP_ON_BUDGET", false) {
						return "", fmt.Errorf("cost or token budget exceeded (tokens=%d cost=$%.4f)", a.Cost.TotalTokens(), a.Cost.TotalCost())
					}
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

		// Execute tool calls (extracted helper)
		toolMsgs, hadErrors, execErr := a.executeToolCalls(ctx, res.ToolCalls, step)
		stepHadErrors = hadErrors
		if execErr != nil {
			return "", execErr
		}
		msgs = append(msgs, toolMsgs...)

		// Handle error recovery logic
		if stepHadErrors {
			consecutiveErrors++
			if consecutiveErrors > a.ErrorHandling.MaxErrorRetries {
				return "", fmt.Errorf("too many consecutive errors (%d), stopping execution", consecutiveErrors)
			}
			debug.Printf("Agent '%s' had errors in step, continuing with error feedback (consecutive: %d/%d)",
				a.ID, consecutiveErrors, a.ErrorHandling.MaxErrorRetries)
		} else {
			consecutiveErrors = 0
		}
		a.Mem.AddStep(step)
		_ = a.Checkpoint(ctx)
	}
}

// executeToolCalls runs model-requested tool calls with cancellation & error handling.
func (a *Agent) executeToolCalls(ctx context.Context, calls []model.ToolCall, step memory.Step) ([]model.ChatMessage, bool, error) {
	var msgs []model.ChatMessage
	hadErrors := false
	for _, tc := range calls {
		select { // cancellation between tools
		case <-ctx.Done():
			return msgs, hadErrors, ctx.Err()
		default:
		}
		t, ok := a.Tools.Use(tc.Name)
		if !ok {
			errorMsg := fmt.Sprintf("Error: Unknown tool '%s'. Available tools: %v", tc.Name, getToolNames(a.Tools))
			if a.ErrorHandling.TreatErrorsAsResults {
				step.ToolResults[tc.ID] = errorMsg
				msgs = append(msgs, model.ChatMessage{Role: "tool", ToolCallID: tc.ID, Content: errorMsg})
				hadErrors = true
				continue
			}
			return msgs, hadErrors, fmt.Errorf("unknown tool: %s", tc.Name)
		}
		var args map[string]any
		if err := json.Unmarshal(tc.Arguments, &args); err != nil {
			errorMsg := fmt.Sprintf("Error: Invalid tool arguments for '%s': %v", tc.Name, err)
			if a.ErrorHandling.TreatErrorsAsResults {
				step.ToolResults[tc.ID] = errorMsg
				msgs = append(msgs, model.ChatMessage{Role: "tool", ToolCallID: tc.ID, Content: errorMsg})
				hadErrors = true
				continue
			}
			return msgs, hadErrors, err
		}
		applyVarsMap(args, a.Vars)
		debug.Printf("Agent '%s' executing tool '%s' with args: %v", a.ID, tc.Name, args)
		a.Trace(ctx, trace.EventToolStart, map[string]any{"name": tc.Name, "args": args})
		start := time.Now()
		r, err := t.Execute(ctx, args)
		toolLatency.WithLabelValues(a.ID.String(), tc.Name).Observe(time.Since(start).Seconds())
		if err != nil {
			debug.Printf("Agent '%s' tool '%s' failed: %v", a.ID, tc.Name, err)
			var errorMsg string
			if a.ErrorHandling.IncludeErrorContext {
				errorMsg = fmt.Sprintf("Error executing tool '%s': %v\n\nContext:\n- Tool: %s\n- Arguments: %v\n- Suggestion: Please try a different approach or check the tool usage.", tc.Name, err, tc.Name, args)
			} else {
				errorMsg = fmt.Sprintf("Error executing tool '%s': %v", tc.Name, err)
			}
			if a.ErrorHandling.TreatErrorsAsResults {
				step.ToolResults[tc.ID] = errorMsg
				msgs = append(msgs, model.ChatMessage{Role: "tool", ToolCallID: tc.ID, Content: errorMsg})
				hadErrors = true
				continue
			}
			return msgs, hadErrors, err
		}
		debug.Printf("Agent '%s' tool '%s' succeeded, result length: %d", a.ID, tc.Name, len(r))
		a.Trace(ctx, trace.EventToolEnd, map[string]any{"name": tc.Name, "result": r})
		step.ToolResults[tc.ID] = r
		msgs = append(msgs, model.ChatMessage{Role: "tool", ToolCallID: tc.ID, Content: r})
	}
	return msgs, hadErrors, nil
}
