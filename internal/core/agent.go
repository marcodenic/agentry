package core

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/marcodenic/agentry/internal/cost"
	"github.com/marcodenic/agentry/internal/debug"
	"github.com/marcodenic/agentry/internal/env"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/tokens"
	"github.com/marcodenic/agentry/internal/tool"
	"github.com/marcodenic/agentry/internal/trace"
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
	toolNamesMu     sync.RWMutex
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
	sort.Strings(names)
	return names
}

// toolNames returns cached tool names for this agent (compute once)
func (a *Agent) toolNames() []string {
	a.toolNamesMu.RLock()
	cached := a.cachedToolNames
	a.toolNamesMu.RUnlock()
	if cached != nil {
		return cached
	}
	names := getToolNames(a.Tools)
	a.toolNamesMu.Lock()
	a.cachedToolNames = names
	a.toolNamesMu.Unlock()
	return names
}

// InvalidateToolCache clears the cached tool name slice (call after mutating Tools)
func (a *Agent) InvalidateToolCache() { //lint:ignore U1000 exported for internal package use
	a.toolNamesMu.Lock()
	a.cachedToolNames = nil
	a.toolNamesMu.Unlock()
}

var redactPatterns = []*regexp.Regexp{
	// OpenAI style keys
	regexp.MustCompile(`sk-[A-Za-z0-9]{16,}`),
	// AWS access keys (approx)
	regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
	// Generic bearer tokens (very naive shortening)
	regexp.MustCompile(`(?i)bearer [A-Za-z0-9\-_.]{20,}`),
}

func sanitizeForLog(s string) string {
	if s == "" {
		return s
	}
	out := s
	for _, re := range redactPatterns {
		out = re.ReplaceAllString(out, "***redacted***")
	}
	if len(out) > 400 {
		out = out[:400] + "…(truncated)"
	}
	return out
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
	safeInputPreview := sanitizeForLog(input[:min(300, len(input))])
	debug.Printf("Agent.Run: Input: %s", safeInputPreview)

	a.Trace(ctx, trace.EventModelStart, a.ModelName)
	msgs := BuildMessages(a.Prompt, a.Vars, a.Mem.History(), input)
	specs := tool.BuildSpecs(a.Tools)

	// ---- Context Size Instrumentation & Trimming ----
	// Calculate token usage of current messages (pre tool specs) and trim if needed.
	startMeasure := time.Now()
	maxContextTokens := env.Int("AGENTRY_CONTEXT_MAX_TOKENS", 0) // hard cap if set
	if maxContextTokens == 0 {
		// Derive from pricing table context window with safety margin
		pt := cost.NewPricingTable()
		limit := pt.GetContextLimit(a.ModelName)
		if limit <= 0 { // fallback
			limit = 120000
		}
		// Leave 15% headroom and min 2k for completion
		headroom := int(float64(limit) * 0.85)
		if headroom < 4000 {
			headroom = limit - 2000
		}
		if headroom < 2000 {
			headroom = limit - 1000
		}
		maxContextTokens = headroom
	}
	reserveForOutput := env.Int("AGENTRY_CONTEXT_RESERVE_OUTPUT", 1024)
	if reserveForOutput < 256 {
		reserveForOutput = 256
	}
	targetBudget := maxContextTokens - reserveForOutput
	if targetBudget < 1000 { // sanity guard
		targetBudget = maxContextTokens - 500
	}

	// Count tokens per message (approx) and trim oldest assistant/tool sections if over budget
	totalTokens := 0
	perMsgTokens := make([]int, len(msgs))
	for i, m := range msgs {
		tk := tokens.Count(m.Content, a.ModelName)
		perMsgTokens[i] = tk
		totalTokens += tk
	}
	if totalTokens > targetBudget {
		debug.Printf("Context trimming: initial=%d budget=%d reserve=%d model=%s", totalTokens, targetBudget, reserveForOutput, a.ModelName)
		// Keep system (index 0) & final user message (last). Remove/compact middle from oldest forward.
		// We'll drop assistant/tool pairs oldest-first until within budget.
		// We rebuild msgs slice rather than mutate in place for clarity.
		systemMsg := msgs[0]
		userMsg := msgs[len(msgs)-1]
		// Collect conversation pairs excluding system/user final
		mid := msgs[1 : len(msgs)-1]
		// We will remove from the front.
		idx := 0
		for totalTokens > targetBudget && idx < len(mid) {
			removed := tokens.Count(mid[idx].Content, a.ModelName)
			totalTokens -= removed
			mid[idx].Content = "" // mark
			idx++
		}
		// Reassemble without emptied messages
		newMid := make([]model.ChatMessage, 0, len(mid))
		for _, m := range mid {
			if strings.TrimSpace(m.Content) == "" && m.Role != "system" { // skip removed; system won't appear here
				continue
			}
			newMid = append(newMid, m)
		}
		msgs = append([]model.ChatMessage{systemMsg}, append(newMid, userMsg)...)
		debug.Printf("Context trimmed: finalTokens≈%d removedMessages=%d", totalTokens, idx)
	}
	if env.Bool("AGENTRY_DEBUG_CONTEXT", false) {
		// Detailed breakdown (optional heavy)
		var sb strings.Builder
		sb.WriteString("[CONTEXT BREAKDOWN]\n")
		for i, m := range msgs {
			role := m.Role
			if role == "system" && i == 0 {
				role = "system(root)"
			}
			sb.WriteString(fmt.Sprintf("%02d %-8s tokens=%d len=%d\n", i, role, tokens.Count(m.Content, a.ModelName), len(m.Content)))
		}
		sb.WriteString(fmt.Sprintf("Total≈%d (budget=%d reserve=%d) buildTime=%s\n", func() int {
			t := 0
			for _, m := range msgs {
				t += tokens.Count(m.Content, a.ModelName)
			}
			return t
		}(), targetBudget, reserveForOutput, time.Since(startMeasure)))
		debug.Printf(sb.String())
	}
	// ---- End Context Size Instrumentation & Trimming ----

	debug.Printf("Agent.Run: Built %d messages (post-trim), %d tool specs", len(msgs), len(specs))
	debug.Printf("Agent.Run: About to call model client with model %s", a.ModelName)

	// DEBUG: Print the full system prompt that will be sent to the API
	if len(msgs) > 0 && msgs[0].Role == "system" {
		debug.Printf("=== FULL SYSTEM PROMPT (sanitized) ===")
		debug.Printf("%s", sanitizeForLog(msgs[0].Content))
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
	if maxIter == 0 { // provide a safety default
		maxIter = 12
	}
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
		// If client supports streaming, prefer that path
		if sc, ok := a.Client.(model.StreamingClient); ok {
			streamCh, sErr := sc.Stream(ctx, msgs, specs)
			if sErr == nil && streamCh != nil {
				var assembled string
				var finalToolCalls []model.ToolCall
				firstTokenRecorded := false
				var sb strings.Builder
				for chunk := range streamCh {
					if chunk.Err != nil {
						return "", chunk.Err
					}
					if chunk.ContentDelta != "" {
						sb.WriteString(chunk.ContentDelta)
						if !firstTokenRecorded {
							firstTokenRecorded = true
						}
						// Emit raw delta for TUI-side smoothing
						a.Trace(ctx, trace.EventToken, chunk.ContentDelta)
					}
					if chunk.Done {
						finalToolCalls = chunk.ToolCalls
					}
				}
				assembled = sb.String()
				// After streaming, treat result as a single completion
				res := model.Completion{Content: assembled, ToolCalls: finalToolCalls}
				debug.Printf("Agent.Run: Streaming completed with %d tool calls", len(res.ToolCalls))
				a.Trace(ctx, trace.EventStepStart, res)
				// Approximate output tokens & update cost
				outTok := tokens.Count(res.Content, a.ModelName)
				if a.Cost != nil {
					a.Cost.AddModelUsage(a.ModelName, 0, outTok)
					if a.Cost.OverBudget() && env.Bool("AGENTRY_STOP_ON_BUDGET", false) {
						return "", fmt.Errorf("cost or token budget exceeded (tokens=%d cost=$%.4f)", a.Cost.TotalTokens(), a.Cost.TotalCost())
					}
				}
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
				if execErr != nil {
					return "", execErr
				}
				msgs = append(msgs, toolMsgs...)
				a.Mem.AddStep(step)
				_ = a.Checkpoint(ctx)
				if hadErrors {
					consecutiveErrors++
				} else {
					consecutiveErrors = 0
				}
				if consecutiveErrors > a.ErrorHandling.MaxErrorRetries {
					return "", fmt.Errorf("too many consecutive errors (%d), stopping execution", consecutiveErrors)
				}
				// Continue outer for-loop for next iteration
				continue
			}
		}
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
		r, err := t.Execute(ctx, args)
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
