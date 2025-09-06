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

// buildMessages creates the message chain for the agent (replaces context package)
func (a *Agent) buildMessages(prompt, input string, history []memory.Step) []model.ChatMessage {
	msgs := []model.ChatMessage{{Role: "system", Content: prompt}}

	// Include only the most recent history step to maintain tool call context
	// while preventing exponential growth. Most agents only need the immediate context.
	if len(history) > 0 {
		lastStep := history[len(history)-1]
		msgs = append(msgs, model.ChatMessage{Role: "assistant", Content: lastStep.Output, ToolCalls: lastStep.ToolCalls})
		for id, res := range lastStep.ToolResults {
			// Truncate large tool results to prevent context bloat
			truncatedRes := res
			if len(res) > 2048 {
				truncatedRes = res[:2048] + "...\n[TRUNCATED: originally " + fmt.Sprintf("%d", len(res)) + " bytes]"
			}
			msgs = append(msgs, model.ChatMessage{Role: "tool", ToolCallID: id, Content: truncatedRes})
		}
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
	}
}

func (a *Agent) Run(ctx context.Context, input string) (string, error) {
	debug.Printf("Agent.Run: Agent ID=%s, Prompt length=%d chars", a.ID.String()[:8], len(a.Prompt))
	debug.Printf("Agent.Run: Available tools: %v", a.toolNames())
	safeInputPreview := sanitizeForLog(input[:min(300, len(input))])
	debug.Printf("Agent.Run: Input: %s", safeInputPreview)

	a.Trace(ctx, trace.EventModelStart, a.ModelName)

	prompt := a.Prompt
	if prompt == "" {
		prompt = defaultPrompt()
	}
	// Inject platform/tool guidance only once per agent (persist into Agent.Prompt)
	if !strings.Contains(prompt, "<!-- PLATFORM_CONTEXT_START -->") {
		prompt = InjectPlatformContextFromRegistry(prompt, a.Tools)
		a.Prompt = prompt
	}
	prompt = applyVars(prompt, a.Vars)

	specs := tool.BuildSpecs(a.Tools)
	// Build initial messages and apply budgeting
	msgs := a.buildMessages(prompt, input, a.Mem.History())
	msgs = a.applyBudget(msgs, specs)

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

	// Track recent tool calls to detect infinite loops
	type toolCallSignature struct {
		Name string
		Args string
	}
	var recentToolCalls []toolCallSignature
	const maxRecentCalls = 6    // Track last 6 tool calls
	const maxIdenticalCalls = 3 // Break if we see the same call 3 times in recent history

	// Iteration cap: prevent infinite loops while allowing reasonable work
	maxIter := env.Int("AGENTRY_MAX_ITER", 50) // Default to 50 iterations max
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
		streamCh, sErr := a.Client.Stream(ctx, msgs, specs)
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
				if chunk.InputTokens > 0 {
					inputTokensUsed = chunk.InputTokens
				}
				if chunk.OutputTokens > 0 {
					outputTokensUsed = chunk.OutputTokens
				}
				if chunk.ModelName != "" {
					modelNameUsed = chunk.ModelName
				}
			}
		}
		assembled = sb.String()

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

		// Append assistant message
		msgs = append(msgs, model.ChatMessage{Role: "assistant", Content: res.Content, ToolCalls: res.ToolCalls})
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
			a.Mem.AddStep(step)
			_ = a.Checkpoint(ctx)
			a.Trace(ctx, trace.EventFinal, res.Content)
			return res.Content, nil
		}

		// Check for repeated identical tool calls to prevent infinite loops
		if len(res.ToolCalls) > 0 {
			// Create signature for this batch of tool calls
			for _, tc := range res.ToolCalls {
				// Serialize args to string for comparison
				argsBytes, _ := json.Marshal(tc.Arguments)
				signature := toolCallSignature{
					Name: tc.Name,
					Args: string(argsBytes),
				}

				// Add to recent calls (maintain sliding window)
				recentToolCalls = append(recentToolCalls, signature)
				if len(recentToolCalls) > maxRecentCalls {
					recentToolCalls = recentToolCalls[1:]
				}

				// Count identical calls in recent history
				identicalCount := 0
				for _, recent := range recentToolCalls {
					if recent.Name == signature.Name && recent.Args == signature.Args {
						identicalCount++
					}
				}

				if identicalCount >= maxIdenticalCalls {
					debug.Printf("Agent.Run: Detected repeated tool call (%s) %d times, breaking loop", tc.Name, identicalCount)
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
				a.Mem.AddStep(step)
				_ = a.Checkpoint(ctx)
				a.Trace(ctx, trace.EventFinal, out)
				return out, nil
			}
		}
		msgs = append(msgs, toolMsgs...)
		a.Mem.AddStep(step)
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

		// Show successful tool execution result to user
		if os.Getenv("AGENTRY_TUI_MODE") != "1" {
			fmt.Fprintf(os.Stderr, "✅ %s: %s completed\n", a.ID.String()[:8], tc.Name)
		}

		a.Trace(ctx, trace.EventToolEnd, map[string]any{"name": tc.Name, "result": r})
		step.ToolResults[tc.ID] = r
		debug.Printf("Agent '%s' adding tool result to messages, role=tool, callID=%s", a.ID, tc.ID)
		msgs = append(msgs, model.ChatMessage{Role: "tool", ToolCallID: tc.ID, Content: r})
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
