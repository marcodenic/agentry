package core

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"sync"

	"github.com/google/uuid"
	"github.com/marcodenic/agentry/internal/cost"
	"github.com/marcodenic/agentry/internal/debug"
	"github.com/marcodenic/agentry/internal/env"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
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
	return newPromptEnvelope(a).Build(prompt, input, history)
}

// applyBudget trims messages to fit within the model's context window budget
func (a *Agent) applyBudget(msgs []model.ChatMessage, specs []model.ToolSpec) []model.ChatMessage {
	return newContextBudgetManager(a.ModelName, specs).Trim(msgs)
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
