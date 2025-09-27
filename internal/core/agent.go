package core

import (
	"context"
	"fmt"
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

	// Include only the most recent history step to maintain context
	if len(history) > 0 {
		lastStep := history[len(history)-1]
		debug.Printf("Including LAST history step in messages:")
		debug.Printf("  Input: %.200s...", lastStep.Input)
		debug.Printf("  Output: %.200s...", lastStep.Output)
		debug.Printf("  ToolCalls: %d", len(lastStep.ToolCalls))
		debug.Printf("  ToolResults: %d", len(lastStep.ToolResults))

		if strings.TrimSpace(lastStep.Input) != "" {
			msgs = append(msgs, model.ChatMessage{Role: "user", Content: lastStep.Input})
		}
		if strings.TrimSpace(lastStep.Output) != "" || len(lastStep.ToolCalls) > 0 {
			msgs = append(msgs, model.ChatMessage{Role: "assistant", Content: lastStep.Output, ToolCalls: lastStep.ToolCalls})
		}
		for id, res := range lastStep.ToolResults {
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
	a.toolNamesMu.RLock()
	cached := a.cachedToolNames
	a.toolNamesMu.RUnlock()
	if cached != nil {
		if debug.IsTraceEnabled() {
			debug.Printf("toolNames: cache hit with %d names", len(cached))
		}
		return cached
	}
	names := getToolNames(a.Tools)
	a.toolNamesMu.Lock()
	a.cachedToolNames = names
	a.toolNamesMu.Unlock()
	if debug.IsTraceEnabled() {
		debug.Printf("toolNames: cache populated with %d names", len(names))
	}
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
	return newConversationSession(a, ctx, input).Run()
}

// executeToolCalls runs model-requested tool calls with cancellation & error handling.
func (a *Agent) executeToolCalls(ctx context.Context, calls []model.ToolCall, step memory.Step) ([]model.ChatMessage, bool, error) {
	return newToolExecutor(a).execute(ctx, calls, step)
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
