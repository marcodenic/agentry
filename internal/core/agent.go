package core

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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
	promptpkg "github.com/marcodenic/agentry/internal/prompt"
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
	// Optional iteration cap for debugging (0 = unlimited)
	MaxIter int
	// Error handling configuration
	ErrorHandling ErrorHandlingConfig
	// JSON validation for tool args, responses, and outputs
	JSONValidator *JSONValidator
	// Role for display
	Role string

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

// buildMessages creates the message chain for the agent (replaces context package)
func (a *Agent) buildMessages(prompt, input string, history []memory.Step) []model.ChatMessage {
	debug.Printf("=== buildMessages START ===")
	debug.Printf("History length: %d steps", len(history))
	for i, step := range history {
		debug.Printf("  HISTORY[%d]: Output=%.100s..., ToolCalls=%d, ToolResults=%d",
			i, step.Output, len(step.ToolCalls), len(step.ToolResults))
	}
	debug.Printf("=== buildMessages processing ===")

	// Always wrap the system prompt into simple sections using sectionizer
	// Build extras sections for the prompt envelope
	extras := map[string]string{}
	if a.Vars != nil {
		if s, ok := a.Vars["AGENTS_SECTION"]; ok {
			extras["agents"] = s
		}
	}

	// Add OS/platform guidance as part of tools section
	// Compose allowedBuiltins from registry and a standard set of command examples.
	{
		names := make([]string, 0, len(a.Tools))
		for n := range a.Tools {
			names = append(names, n)
		}
		sort.Strings(names)
		allowedCommands := []string{"list", "view", "write", "run", "search", "find", "cwd", "env"}
		guidance := GetPlatformContext(allowedCommands, names)
		if strings.TrimSpace(guidance) != "" {
			extras["tool_guidance"] = guidance
		}
	}
	// Placeholders for optional sections users might want to see; keep minimal by default
	// extras["output-format"] = "" // left empty unless explicitly provided by templates/config

	// Use default prompt if none provided
	if strings.TrimSpace(prompt) == "" {
		prompt = defaultPrompt()
	}

	prompt = promptpkg.Sectionize(prompt, a.Tools, extras)

	msgs := []model.ChatMessage{{Role: "system", Content: prompt}}

	// Include only the most recent history step to maintain tool call context
	// while preventing exponential growth. Most agents only need the immediate context.
	if len(history) > 0 {
		lastStep := history[len(history)-1]
		debug.Printf("Including LAST history step in messages:")
		debug.Printf("  Output: %.200s...", lastStep.Output)
		debug.Printf("  ToolCalls: %d", len(lastStep.ToolCalls))
		debug.Printf("  ToolResults: %d", len(lastStep.ToolResults))

		msgs = append(msgs, model.ChatMessage{Role: "assistant", Content: lastStep.Output, ToolCalls: lastStep.ToolCalls})
		for id, res := range lastStep.ToolResults {
			// Truncate large tool results to prevent context bloat
			truncatedRes := res
			if len(res) > 2048 {
				truncatedRes = res[:2048] + "...\n[TRUNCATED: originally " + fmt.Sprintf("%d", len(res)) + " bytes]"
				debug.Printf("  TRUNCATED tool result %s: %d -> %d chars", id, len(res), len(truncatedRes))
			}
			msgs = append(msgs, model.ChatMessage{Role: "tool", ToolCallID: id, Content: truncatedRes})
		}
	} else {
		debug.Printf("No history to include in messages")
	}

	msgs = append(msgs, model.ChatMessage{Role: "user", Content: input})
	return msgs
}

// applyBudget trims messages to fit within the model's context window budget
func (a *Agent) applyBudget(msgs []model.ChatMessage, specs []model.ToolSpec) []model.ChatMessage {
	maxContextTokens := env.Int("AGENTRY_CONTEXT_MAX_TOKENS", 0)
	if maxContextTokens == 0 {
		pt := cost.NewPricingTable()
		limit := pt.GetContextLimit(a.ModelName)
		if limit <= 0 {
			limit = 120000
		}
		headroom := int(float64(limit) * 0.85)
		if headroom < 4000 {
			headroom = limit - 2000
		}
		if headroom < 2000 {
			headroom = limit - 1000
		}
		maxContextTokens = headroom
	}
	// Additional safeguard: for very large Anthropic windows, keep budget moderate
	if strings.Contains(strings.ToLower(a.ModelName), "claude") && maxContextTokens > 60000 {
		maxContextTokens = 60000
	}
	reserveForOutput := env.Int("AGENTRY_CONTEXT_RESERVE_OUTPUT", 1024)
	if reserveForOutput < 256 {
		reserveForOutput = 256
	}
	targetBudget := maxContextTokens - reserveForOutput
	if targetBudget < 1000 {
		targetBudget = maxContextTokens - 500
	}

	totalTokens := 0
	for _, m := range msgs {
		totalTokens += tokens.Count(m.Content, a.ModelName)
		for _, tc := range m.ToolCalls {
			totalTokens += tokens.Count(tc.Name, a.ModelName)
			totalTokens += tokens.Count(string(tc.Arguments), a.ModelName)
		}
	}
	// Include tool schema tokens (names, descriptions, parameter JSON) so trimming considers them
	toolSchemaTokens := 0
	if len(specs) > 0 {
		for _, s := range specs {
			toolSchemaTokens += tokens.Count(s.Name, a.ModelName)
			toolSchemaTokens += tokens.Count(s.Description, a.ModelName)
			// crude JSON size counting – convert map to string naïvely
			if len(s.Parameters) > 0 {
				// Quick approximation: join keys
				for k, v := range s.Parameters {
					toolSchemaTokens += tokens.Count(k, a.ModelName)
					// parameter value structure size approximation
					toolSchemaTokens += tokens.Count(fmt.Sprintf("%v", v), a.ModelName)
				}
			}
		}
	}
	totalWithTools := totalTokens + toolSchemaTokens
	if totalWithTools <= targetBudget {
		return msgs
	}
	// Guard: need at least system + user to do mid trimming safely
	if len(msgs) < 2 {
		debug.Printf("applyBudget: only %d messages available; skipping mid-trim", len(msgs))
		return msgs
	}

	debug.Printf("Context trimming: initial=%d (msgs=%d + tools≈%d) budget=%d reserve=%d model=%s", totalWithTools, totalTokens, toolSchemaTokens, targetBudget, reserveForOutput, a.ModelName)
	systemMsg := msgs[0]
	userMsg := msgs[len(msgs)-1]
	mid := msgs[1 : len(msgs)-1]
	idx := 0
	for (totalTokens+toolSchemaTokens) > targetBudget && idx < len(mid) {
		removed := tokens.Count(mid[idx].Content, a.ModelName)
		for _, tc := range mid[idx].ToolCalls {
			removed += tokens.Count(tc.Name, a.ModelName)
			removed += tokens.Count(string(tc.Arguments), a.ModelName)
		}
		totalTokens -= removed
		mid[idx].Content = ""
		mid[idx].ToolCalls = nil
		idx++
	}
	newMid := make([]model.ChatMessage, 0, len(mid))
	for _, m := range mid {
		if strings.TrimSpace(m.Content) == "" && m.Role != "system" {
			continue
		}
		newMid = append(newMid, m)
	}
	msgs = append([]model.ChatMessage{systemMsg}, append(newMid, userMsg)...)
	debug.Printf("Context trimmed: finalTokens≈%d (msgs) + tools≈%d removedMessages=%d", totalTokens, toolSchemaTokens, idx)

	return msgs
}

// toolNames returns cached tool names for this agent (compute once)
func (a *Agent) toolNames() []string {
	debug.Printf("toolNames: Attempting RLock")
	a.toolNamesMu.RLock()
	debug.Printf("toolNames: Acquired RLock")
	cached := a.cachedToolNames
	a.toolNamesMu.RUnlock()
	debug.Printf("toolNames: Released RLock, cached=%v", cached != nil)
	if cached != nil {
		return cached
	}
	debug.Printf("toolNames: Cache miss, calling getToolNames")
	names := getToolNames(a.Tools)
	debug.Printf("toolNames: getToolNames returned %d names, attempting Lock", len(names))
	a.toolNamesMu.Lock()
	debug.Printf("toolNames: Acquired Lock")
	a.cachedToolNames = names
	a.toolNamesMu.Unlock()
	debug.Printf("toolNames: Released Lock, returning %d names", len(names))
	return names
}

// InvalidateToolCache clears the cached tool name slice (call after mutating Tools)
func (a *Agent) InvalidateToolCache() { //lint:ignore U1000 exported for internal package use
	a.toolNamesMu.Lock()
	a.cachedToolNames = nil
	a.toolNamesMu.Unlock()
}

// allToolCallsTerminal returns true if every tool call corresponds to a tool
// implementing a Terminal() bool method (avoids direct import dep on tool.TerminalAware).
func (a *Agent) allToolCallsTerminal(calls []model.ToolCall) bool {
	if len(calls) == 0 {
		return false
	}
	for _, tc := range calls {
		t, ok := a.Tools.Use(tc.Name)
		if !ok {
			return false
		}
		if ta, ok := t.(interface{ Terminal() bool }); !ok || !ta.Terminal() {
			return false
		}
	}
	return true
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
		JSONValidator: NewJSONValidator(),
		Role:          "agent", // Default role
	}
}

func (a *Agent) Run(ctx context.Context, input string) (string, error) {
	debug.Printf("Agent.Run: Agent ID=%s, Prompt length=%d chars", a.ID.String()[:8], len(a.Prompt))
	debug.Printf("Agent.Run: Available tools: %v", a.toolNames())
	safeInputPreview := sanitizeForLog(input[:min(300, len(input))])
	debug.Printf("Agent.Run: Input: %s", safeInputPreview)

	if resetter, ok := a.Client.(interface{ ResetConversation() }); ok {
		resetter.ResetConversation()
	}

	a.Trace(ctx, trace.EventModelStart, a.ModelName)

	prompt := a.Prompt
	if prompt == "" {
		prompt = defaultPrompt()
	}
	// Skip legacy platform/tool guidance injection to avoid duplication.
	prompt = applyVars(prompt, a.Vars)

	specs := tool.BuildSpecs(a.Tools)
	// Build initial messages without history to prevent prompt bloat
	msgs := a.buildMessages(prompt, input, []memory.Step{})
	msgs = a.applyBudget(msgs, specs)

	debug.Printf("Agent.Run: Built %d messages (post-trim), %d tool specs", len(msgs), len(specs))
	debug.Printf("Agent.Run: About to call model client with model %s", a.ModelName)
	debug.Printf("Agent.Run: ENTERING MAIN LOOP NOW - BEFORE FOR LOOP")

	// DEBUG: Print ALL messages that will be sent to the API
	debug.Printf("=== FULL MESSAGE PAYLOAD TO API ===")
	totalChars := 0
	for i, msg := range msgs {
		msgSize := len(msg.Content)
		totalChars += msgSize
		debug.Printf("[MSG %d] Role: %s, Size: %d chars, ToolCalls: %d", i, msg.Role, msgSize, len(msg.ToolCalls))
		if msg.Role == "system" {
			debug.Printf("  SYSTEM CONTENT (first 500 chars): %.500s...", msg.Content)
		} else if msg.Role == "user" {
			debug.Printf("  USER CONTENT: %s", msg.Content)
		} else if msg.Role == "assistant" {
			debug.Printf("  ASSISTANT CONTENT: %.200s...", msg.Content)
			for j, tc := range msg.ToolCalls {
				debug.Printf("    TOOL_CALL[%d]: %s(%s)", j, tc.Name, string(tc.Arguments))
			}
		} else if msg.Role == "tool" {
			debug.Printf("  TOOL RESULT (ID: %s): %.200s...", msg.ToolCallID, msg.Content)
		}
	}
	debug.Printf("=== TOTAL PAYLOAD: %d messages, %d total chars ===", len(msgs), totalChars)

	// DEBUG: Print available tool specs
	debug.Printf("=== AVAILABLE TOOLS ===")
	for _, spec := range specs {
		debug.Printf("Tool: %s", spec.Name)
	}
	debug.Printf("=== END TOOLS ===")

	// DEBUG: Print available tool names from registry
	debug.Printf("=== AVAILABLE TOOL NAMES ===")
	debug.Printf("About to call a.toolNames() - BEFORE MUTEX")
	toolNames := a.toolNames()
	debug.Printf("Returned from a.toolNames() - AFTER MUTEX, got %d names", len(toolNames))
	for _, name := range toolNames {
		debug.Printf("Tool: %s", name)
	}
	debug.Printf("=== END TOOL NAMES ===")

	// Do not estimate tokens here; rely on actual counts from responses

	// Track consecutive errors for resilience
	consecutiveErrors := 0

	// Track recent tool calls to detect infinite loops
	type toolCallSignature struct {
		Name string
		Args string
	}
	var recentToolCalls []toolCallSignature
	const maxRecentCalls = 6    // Track last 6 tool calls
	const maxIdenticalCalls = 3 // Break if we see the same call 3 times in recent history

	// Optional iteration cap (0 = unlimited), set via CLI flag
	maxIter := a.MaxIter
	for i := 0; ; i++ {
		debug.Printf("Agent.Run: *** ITERATION %d START ***", i)
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
		// Apply budgeting including tool schemas for accurate trimming
		msgs = a.applyBudget(msgs, specs)
		debug.Printf("Agent.Run: Current message count: %d", len(msgs))
		if i > 0 {
			// Log recent messages to see what's causing continued iterations
			debug.Printf("Agent.Run: Recent messages in iteration %d:", i)
			start := len(msgs) - 3
			if start < 0 {
				start = 0
			}
			for j := start; j < len(msgs); j++ {
				debug.Printf("  [%d] Role: %s, Content: %.100s...", j, msgs[j].Role, msgs[j].Content)
			}
		}
		// Use streaming API exclusively (responses API only supports streaming)
		debug.Printf("Agent.Run: About to call model with %d messages", len(msgs))
		debug.Printf("Agent.Run: WHAT TRIGGERS NEW CALL? Messages context:")
		for j, msg := range msgs {
			debug.Printf("  MSG[%d] Role:%s ToolCalls:%d Content:%.150s...", j, msg.Role, len(msg.ToolCalls), msg.Content)
		}
		debug.Printf("Agent.Run: CALLING MODEL CLIENT NOW - BEFORE STREAM")
		streamStartTime := time.Now()
		streamCh, sErr := a.Client.Stream(ctx, msgs, specs)
		streamCallDuration := time.Since(streamStartTime)
		debug.Printf("Agent.Run: MODEL CLIENT RETURNED - AFTER STREAM, err=%v, call_duration=%v", sErr, streamCallDuration)
		if sErr != nil {
			return "", sErr
		}
		if streamCh == nil {
			return "", fmt.Errorf("streaming client returned nil channel")
		}

		var assembled string
		var finalToolCalls []model.ToolCall
		var inputTokensUsed, outputTokensUsed int
		var modelNameUsed string
		var responseIDUsed string
		firstTokenRecorded := false
		var sb strings.Builder

		debug.Printf("Agent.Run: Starting to read from stream channel")
		chunkCount := 0
		streamReadStartTime := time.Now()
		for chunk := range streamCh {
			chunkCount++
			debug.Printf("Agent.Run: Received chunk %d", chunkCount)
			if chunk.Err != nil {
				debug.Printf("Agent.Run: Chunk error: %v", chunk.Err)
				return "", chunk.Err
			}
			if chunk.ContentDelta != "" {
				sb.WriteString(chunk.ContentDelta)
				if !firstTokenRecorded {
					firstTokenRecorded = true
					debug.Printf("Agent.Run: First token received")
				}
				// Emit raw delta for TUI-side smoothing
				a.Trace(ctx, trace.EventToken, chunk.ContentDelta)
			}
			if chunk.Done {
				debug.Printf("Agent.Run: Received final chunk (Done=true)")
				finalToolCalls = chunk.ToolCalls
				if chunk.InputTokens > 0 {
					inputTokensUsed = chunk.InputTokens
				}
				if chunk.OutputTokens > 0 {
					outputTokensUsed = chunk.OutputTokens
				}
				if chunk.ModelName != "" {
					modelNameUsed = chunk.ModelName
				}
				if chunk.ResponseID != "" {
					responseIDUsed = chunk.ResponseID
				}
			}
		}

		debug.Printf("Agent.Run: Stream reading completed with %d chunks, read_duration=%v", chunkCount, time.Since(streamReadStartTime))
		assembled = sb.String()
		debug.Printf("Agent.Run: Assembled response length: %d chars", len(assembled))

		// After streaming, treat result as a single completion
		res := model.Completion{Content: assembled, ToolCalls: finalToolCalls, InputTokens: inputTokensUsed, OutputTokens: outputTokensUsed, ModelName: func() string {
			if modelNameUsed != "" {
				return modelNameUsed
			}
			return a.ModelName
		}()}
		debug.Printf("Agent.Run: Streaming completed with %d tool calls", len(res.ToolCalls))
		debug.Printf("Agent.Run: Agent response content: '%.200s...'", res.Content)
		if len(res.ToolCalls) > 0 {
			debug.Printf("Agent.Run: AGENT IS MAKING TOOL CALLS - WHY?")
			for i, tc := range res.ToolCalls {
				debug.Printf("  TOOL_CALL[%d]: %s with args %s", i, tc.Name, string(tc.Arguments))
			}
		}
		a.Trace(ctx, trace.EventStepStart, res)
		if responseIDUsed != "" {
			// Emit a summary trace with response linkage identifier (for debugging/observability)
			a.Trace(ctx, trace.EventSummary, map[string]any{
				"response_id": responseIDUsed,
			})
		}

		// Approximate output tokens & update cost
		// Prefer API-provided counts else fallback to estimation
		inTok := res.InputTokens
		if inTok == 0 { // fallback estimate based on current messages (excluding assistant response just added later)
			// Roughly count tokens of system + user + last assistant/tool context
			for _, m := range msgs {
				inTok += tokens.Count(m.Content, a.ModelName)
				for _, tc := range m.ToolCalls {
					inTok += tokens.Count(tc.Name, a.ModelName)
					inTok += tokens.Count(string(tc.Arguments), a.ModelName)
				}
			}
		}
		outTok := res.OutputTokens
		if outTok == 0 {
			outTok = tokens.Count(res.Content, a.ModelName)
		}
		if a.Cost != nil {
			// Prefer the provider/model used by the backend when available
			modelForCost := res.ModelName
			if strings.TrimSpace(modelForCost) == "" {
				modelForCost = a.ModelName
			}
			a.Cost.AddModelUsage(modelForCost, inTok, outTok)
			if a.Cost.OverBudget() && env.Bool("AGENTRY_STOP_ON_BUDGET", false) {
				return "", fmt.Errorf("cost or token budget exceeded (tokens=%d cost=$%.4f)", a.Cost.TotalTokens(), a.Cost.TotalCost())
			}
		}

		// Only append assistant message to local context if NOT using conversation linking
		// When responseIDUsed is set, OpenAI maintains conversation state server-side
		if responseIDUsed == "" {
			debug.Printf("Agent.Run: Appending assistant message to local context (no conversation linking)")
			msgs = append(msgs, model.ChatMessage{Role: "assistant", Content: res.Content, ToolCalls: res.ToolCalls})
		} else {
			debug.Printf("Agent.Run: Conversation linking active (responseID: %s), not appending to local context", responseIDUsed)
		}
		step := memory.Step{Output: res.Content, ToolCalls: res.ToolCalls, ToolResults: map[string]string{}}

		if len(res.ToolCalls) == 0 {
			// Previously we injected a follow-up prompt when the model produced a planning-style response.
			// That caused unintended loops in TUI for simple greetings (because the role prompt mentions "plan").
			// Now we return immediately like the non-streaming path, unless explicitly enabled via env.
			debug.Printf("Agent.Run: No tool calls in response, checking for completion")
			debug.Printf("Agent.Run: Response content snippet: '%.200s...'", res.Content)
			if env.Bool("AGENTRY_PLAN_HEURISTIC", false) && len(specs) > 0 && (strings.Contains(strings.ToLower(res.Content), "plan") || strings.Contains(res.Content, "I'll ") || strings.Contains(strings.ToLower(res.Content), "i will")) {
				debug.Printf("Agent.Run: Plan heuristic enabled; injecting follow-up to trigger tools")
				follow := "You provided a plan. Now execute the necessary steps using the available tools. For each data collection action, call the appropriate tool (e.g., sysinfo). Then produce the consolidated final report. Respond only with tool calls until data is gathered."
				msgs = append(msgs, model.ChatMessage{Role: "system", Content: follow})
				continue
			}
			// Default behavior: finalize
			debug.Printf("Agent.Run: Finalizing - no tools needed, returning response")

			// Validate final agent output
			if err := a.JSONValidator.ValidateAgentOutput(res.Content); err != nil {
				debug.Printf("Agent.Run: Agent output validation failed: %v", err)
				return fmt.Sprintf("Agent completed task but output validation failed: %v", err), nil
			}

			// Memory disabled for now: do not persist steps
			_ = a.Checkpoint(ctx)
			a.Trace(ctx, trace.EventFinal, res.Content)
			return res.Content, nil
		}

		// Check for repeated identical tool calls to prevent infinite loops
		if len(res.ToolCalls) > 0 {
			debug.Printf("Agent.Run: Processing %d tool calls for duplicate detection", len(res.ToolCalls))

			// Create signature for this batch of tool calls
			for _, tc := range res.ToolCalls {
				// Serialize args to string for comparison (use raw json bytes)
				signature := toolCallSignature{
					Name: tc.Name,
					Args: string(tc.Arguments),
				}
				debug.Printf("Agent.Run: Tool call signature: %s(%s)", signature.Name, signature.Args)

				// Add to recent calls (maintain sliding window)
				recentToolCalls = append(recentToolCalls, signature)
				if len(recentToolCalls) > maxRecentCalls {
					recentToolCalls = recentToolCalls[1:]
				}

				// Count identical calls in recent history
				identicalCount := 0
				for j, recent := range recentToolCalls {
					if recent.Name == signature.Name && recent.Args == signature.Args {
						identicalCount++
						debug.Printf("Agent.Run: Found identical call at position %d: %s(%s)", j, recent.Name, recent.Args)
					}
				}
				debug.Printf("Agent.Run: Tool %s has %d identical calls in recent history (max=%d)", signature.Name, identicalCount, maxIdenticalCalls)

				if identicalCount >= maxIdenticalCalls {
					debug.Printf("Agent.Run: BREAKING LOOP - Detected repeated tool call (%s) %d times", tc.Name, identicalCount)
					return fmt.Sprintf("Task completed. Detected repeated tool execution (%s), stopping to prevent infinite loop.", tc.Name), nil
				}
			}
		}

		// Execute tools then continue loop with new messages
		toolMsgs, hadErrors, execErr := a.executeToolCalls(ctx, res.ToolCalls, step)
		if execErr != nil {
			return "", execErr
		}

		// Structural finalization: if all executed tools are terminal and no errors
		// occurred, finalize with their combined outputs.
		if !hadErrors && a.allToolCallsTerminal(res.ToolCalls) && len(toolMsgs) > 0 {
			var b strings.Builder
			for _, m := range toolMsgs {
				if m.Role == "tool" && strings.TrimSpace(m.Content) != "" {
					if b.Len() > 0 {
						b.WriteString("\n")
					}
					b.WriteString(m.Content)
				}
			}
			out := strings.TrimSpace(b.String())
			if out != "" {
				debug.Printf("Agent.Run: Finalizing after terminal tool calls (%d tools, %d chars output)", len(res.ToolCalls), len(out))
				_ = a.Checkpoint(ctx)
				a.Trace(ctx, trace.EventFinal, out)
				return out, nil
			}
		}

		// When using conversation linking, we still need to provide tool results
		// for the function calls OpenAI made, even though it maintains conversation state
		if responseIDUsed == "" {
			debug.Printf("Agent.Run: Appending %d tool results to local context (no conversation linking)", len(toolMsgs))
			msgs = append(msgs, toolMsgs...)
		} else {
			debug.Printf("Agent.Run: Conversation linking active (responseID: %s), appending %d tool results for function calls", responseIDUsed, len(toolMsgs))
			msgs = append(msgs, toolMsgs...)
		}
		// Memory disabled for now: do not persist steps
		_ = a.Checkpoint(ctx)

		// DEBUG: Log the messages Agent 0 will see in the next iteration
		debug.Printf("Agent.Run: Messages after tool execution (count=%d):", len(msgs))
		for j, msg := range msgs {
			debug.Printf("  [%d] Role: %s, Content: %.100s...", j, msg.Role, msg.Content)
		}

		if hadErrors {
			consecutiveErrors++
		} else {
			consecutiveErrors = 0
		}
		if consecutiveErrors > a.ErrorHandling.MaxErrorRetries {
			return "", fmt.Errorf("too many consecutive errors (%d), stopping execution", consecutiveErrors)
		}
		// Continue outer for-loop for next iteration
		debug.Printf("Agent.Run: Iteration %d complete, continuing to next iteration", i)
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

		// Validate tool arguments
		if err := a.JSONValidator.ValidateToolArgs(args); err != nil {
			errorMsg := fmt.Sprintf("Error: Invalid tool arguments for '%s': %v", tc.Name, err)
			if a.ErrorHandling.TreatErrorsAsResults {
				step.ToolResults[tc.ID] = errorMsg
				msgs = append(msgs, model.ChatMessage{Role: "tool", ToolCallID: tc.ID, Content: errorMsg})
				hadErrors = true
				continue
			}
			return msgs, hadErrors, err
		}

		// Sanitize tool args before logging to avoid leaking secrets
		if b, _ := json.Marshal(args); len(b) > 0 {
			debug.Printf("Agent '%s' executing tool '%s' with args: %s", a.ID, tc.Name, sanitizeForLog(string(b)))
		} else {
			debug.Printf("Agent '%s' executing tool '%s'", a.ID, tc.Name)
		}
		a.Trace(ctx, trace.EventToolStart, map[string]any{"name": tc.Name, "args": args})

		// Show tool execution to user (not just debug mode)
		if os.Getenv("AGENTRY_TUI_MODE") != "1" {
			// Show tool with key arguments for better visibility
			argSummary := getToolArgSummary(tc.Name, args)
			if argSummary != "" {
				fmt.Fprintf(os.Stderr, "🔧 %s: %s %s\n", a.ID.String()[:8], tc.Name, argSummary)
			} else {
				fmt.Fprintf(os.Stderr, "🔧 %s: %s\n", a.ID.String()[:8], tc.Name)
			}
		}

		r, err := t.Execute(ctx, args)
		debug.Printf("Agent '%s' tool '%s' execute completed, err=%v, result_length=%d", a.ID, tc.Name, err, len(r))
		if err != nil {
			debug.Printf("Agent '%s' tool '%s' failed: %v", a.ID, tc.Name, err)

			// Show tool failure to user
			if os.Getenv("AGENTRY_TUI_MODE") != "1" {
				fmt.Fprintf(os.Stderr, "❌ %s: %s failed: %v\n", a.ID.String()[:8], tc.Name, err)
			}

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

		// Validate tool response
		if err := a.JSONValidator.ValidateToolResponse(r); err != nil {
			errorMsg := fmt.Sprintf("Error: Tool '%s' produced invalid response: %v", tc.Name, err)
			if a.ErrorHandling.TreatErrorsAsResults {
				step.ToolResults[tc.ID] = errorMsg
				msgs = append(msgs, model.ChatMessage{Role: "tool", ToolCallID: tc.ID, Content: errorMsg})
				hadErrors = true
				continue
			}
			return msgs, hadErrors, err
		}

		// Show successful tool execution result to user
		if os.Getenv("AGENTRY_TUI_MODE") != "1" {
			fmt.Fprintf(os.Stderr, "✅ %s: %s completed\n", a.ID.String()[:8], tc.Name)
		}

		a.Trace(ctx, trace.EventToolEnd, map[string]any{"name": tc.Name, "result": r})
		step.ToolResults[tc.ID] = r
		debug.Printf("Agent '%s' adding tool result to messages, role=tool, callID=%s", a.ID, tc.ID)

		// Fix: Ensure empty tool results are interpreted as success by the model
		toolResult := r
		if strings.TrimSpace(r) == "" {
			// For tools that succeed with no output, provide clear success feedback
			switch tc.Name {
			case "bash", "sh":
				toolResult = "Command executed successfully."
			case "create":
				if path, ok := args["path"].(string); ok {
					toolResult = fmt.Sprintf("File '%s' created successfully.", path)
				} else {
					toolResult = "File created successfully."
				}
			case "edit_range", "search_replace":
				toolResult = "File edited successfully."
			default:
				toolResult = "Operation completed successfully."
			}
		}

		msgs = append(msgs, model.ChatMessage{Role: "tool", ToolCallID: tc.ID, Content: toolResult})
	}
	debug.Printf("Agent.Run: executeToolCalls completed, returning %d messages, hadErrors=%v", len(msgs), hadErrors)
	return msgs, hadErrors, nil
}

// getToolArgSummary returns a brief summary of key tool arguments for user-friendly logging
func getToolArgSummary(toolName string, args map[string]any) string {
	switch toolName {
	case "create":
		if path, ok := args["path"].(string); ok {
			return fmt.Sprintf("'%s'", path)
		}
	case "view", "read_lines":
		if path, ok := args["path"].(string); ok {
			return fmt.Sprintf("'%s'", path)
		}
	case "edit_range":
		if path, ok := args["path"].(string); ok {
			start, _ := args["start_line"].(float64)
			end, _ := args["end_line"].(float64)
			return fmt.Sprintf("'%s' (lines %g-%g)", path, start, end)
		}
	case "find":
		if pattern, ok := args["name"].(string); ok {
			if path, pathOk := args["path"].(string); pathOk && path != "." {
				return fmt.Sprintf("'%s' in '%s'", pattern, path)
			}
			return fmt.Sprintf("'%s'", pattern)
		}
	case "grep":
		if pattern, ok := args["pattern"].(string); ok {
			if filePattern, fpOk := args["file_pattern"].(string); fpOk {
				return fmt.Sprintf("'%s' in %s files", pattern, filePattern)
			}
			return fmt.Sprintf("'%s'", pattern)
		}
	case "ls":
		if path, ok := args["path"].(string); ok && path != "." {
			return fmt.Sprintf("'%s'", path)
		}
	case "sh", "bash":
		if cmd, ok := args["command"].(string); ok {
			// Truncate very long commands
			if len(cmd) > 50 {
				return fmt.Sprintf("'%s...'", cmd[:47])
			}
			return fmt.Sprintf("'%s'", cmd)
		}
	case "agent":
		if agent, ok := args["agent"].(string); ok {
			if input, inputOk := args["input"].(string); inputOk {
				// Truncate long inputs
				if len(input) > 40 {
					return fmt.Sprintf("-> %s ('%s...')", agent, input[:37])
				}
				return fmt.Sprintf("-> %s ('%s')", agent, input)
			}
			return fmt.Sprintf("-> %s", agent)
		}
	}
	return ""
}
