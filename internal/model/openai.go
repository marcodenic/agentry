package model

import (
	"context"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/marcodenic/agentry/internal/debug"
)

// defaultHTTPTimeout is the global HTTP client timeout in seconds for model clients.
// It can be overridden by CLI flags via SetHTTPTimeout.
var defaultHTTPTimeout = 300

// SetHTTPTimeout overrides the default HTTP client timeout (in seconds).
func SetHTTPTimeout(seconds int) {
	if seconds > 0 {
		defaultHTTPTimeout = seconds
	}
}

// OpenAI client implemented against /v1/responses (legacy chat completions removed).
type OpenAI struct {
	key         string
	model       string
	Temperature *float64
	client      *http.Client
	// previousResponseID holds the last response ID from Responses API
	// If set, it will be sent as previous_response_id to link conversation state.
	previousResponseID string
}

func NewOpenAI(key, model string) *OpenAI {
	// Create HTTP client with reasonable timeout
	client := &http.Client{
		Timeout: time.Duration(defaultHTTPTimeout) * time.Second,
	}
	return &OpenAI{key: key, model: model, client: client}
}

// Wire types for Responses API
type oaInputItem struct {
	Role    string          `json:"role"`
	Content []oaContentPart `json:"content"`
}
type oaContentPart struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	// For Responses API tool results
	ToolCallID string `json:"tool_call_id,omitempty"`
	Output     string `json:"output,omitempty"`
}

// openAITool matches Responses API tool definition (flattened function schema)
type openAITool struct {
	Type        string         `json:"type"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters"`
}

// partial represents an in-progress tool call assembly during streaming.
type partial struct {
	ToolCall
	index int
}

func buildOATools(tools []ToolSpec) []openAITool {
	oa := make([]openAITool, len(tools))
	for i, t := range tools {
		oa[i].Type = "function"
		oa[i].Name = t.Name
		oa[i].Description = t.Description
		if len(t.Parameters) > 0 {
			oa[i].Parameters = t.Parameters
		} else {
			oa[i].Parameters = map[string]any{"type": "object", "properties": map[string]any{}}
		}
	}
	return oa
}
func buildOAInput(msgs []ChatMessage) []oaInputItem {
	out := make([]oaInputItem, 0, len(msgs))
	for _, m := range msgs {
		// Map role to supported values
		role := m.Role
		// Build content parts by role. Note: Responses API does NOT accept
		// tool results in the input stream (no "tool_result" type). Tool
		// results must be sent via top-level "tool_outputs" with a
		// previous_response_id. Therefore, we ignore any tool-role messages
		// here; they will be handled separately during request construction.
		switch m.Role {
		case "tool":
			// Skip: handled via tool_outputs on continuation
			continue
		case "assistant":
			// Skip empty assistant outputs to avoid invalid output_text without text
			if strings.TrimSpace(m.Content) == "" {
				continue
			}
			out = append(out, oaInputItem{
				Role:    role,
				Content: []oaContentPart{{Type: "output_text", Text: m.Content}},
			})
		case "user", "system":
			if strings.TrimSpace(m.Content) == "" {
				continue
			}
			out = append(out, oaInputItem{Role: role, Content: []oaContentPart{{Type: "input_text", Text: m.Content}}})
		default:
			if strings.TrimSpace(m.Content) == "" {
				continue
			}
			out = append(out, oaInputItem{Role: role, Content: []oaContentPart{{Type: "input_text", Text: m.Content}}})
		}
	}
	return out
}

func (o *OpenAI) Stream(ctx context.Context, msgs []ChatMessage, tools []ToolSpec) (<-chan StreamChunk, error) {
	conv := newOpenAIConversation(o, msgs, tools)
	return conv.Stream(ctx)
}

// estimateTokens provides a rough token count estimate for logging
func estimateTokens(msgs []ChatMessage) int {
	total := 0
	for _, msg := range msgs {
		// Rough estimate: ~4 characters per token
		total += len(msg.Content) / 4
	}
	return total
}

func finalizeOpenAI(partials map[int]*partial, out chan<- StreamChunk, inTok, outTok int, model string, responseID string) {
	if len(partials) == 0 {
		// No tool calls: emit final chunk with usage
		// Include provider/model for accurate pricing
		out <- StreamChunk{Done: true, InputTokens: inTok, OutputTokens: outTok, ModelName: "openai/" + model, ResponseID: responseID}
		return
	}
	idxs := make([]int, 0, len(partials))
	for i := range partials {
		idxs = append(idxs, i)
	}
	sort.Ints(idxs)
	final := make([]ToolCall, 0, len(partials))
	for _, i := range idxs {
		p := partials[i]
		final = append(final, ToolCall{ID: p.ID, Name: p.Name, Arguments: p.Arguments})
	}
	out <- StreamChunk{Done: true, ToolCalls: final, InputTokens: inTok, OutputTokens: outTok, ModelName: "openai/" + model, ResponseID: responseID}
}

func finalizeWithResponses(partials map[int]*partial, responseCalls map[string]*partial, out chan<- StreamChunk, inTok, outTok int, model string, responseID string) {
	// Prefer Responses API events when present; otherwise, fall back to legacy
	// deltas. This avoids double-emitting the same function call when servers
	// provide both representations.

	debug.Printf("finalizeWithResponses: partials=%d responseCalls=%d", len(partials), len(responseCalls))
	for id, p := range responseCalls {
		debug.Printf("  responseCalls[%s]: ID=%s Name=%s", id, p.ID, p.Name)
	}

	if len(responseCalls) > 0 {
		final := make([]ToolCall, 0, len(responseCalls))
		for _, p := range responseCalls {
			debug.Printf("finalizeWithResponses: Using response call ID=%s Name=%s", p.ID, p.Name)
			final = append(final, ToolCall{ID: p.ID, Name: p.Name, Arguments: p.Arguments})
		}
		out <- StreamChunk{Done: true, ToolCalls: final, InputTokens: inTok, OutputTokens: outTok, ModelName: "openai/" + model, ResponseID: responseID}
		return
	}

	// Fallback: legacy partials path
	if len(partials) == 0 {
		out <- StreamChunk{Done: true, InputTokens: inTok, OutputTokens: outTok, ModelName: "openai/" + model, ResponseID: responseID}
		return
	}
	idxs := make([]int, 0, len(partials))
	for i := range partials {
		idxs = append(idxs, i)
	}
	sort.Ints(idxs)
	final := make([]ToolCall, 0, len(partials))
	for _, i := range idxs {
		p := partials[i]
		final = append(final, ToolCall{ID: p.ID, Name: p.Name, Arguments: p.Arguments})
	}
	out <- StreamChunk{Done: true, ToolCalls: final, InputTokens: inTok, OutputTokens: outTok, ModelName: "openai/" + model, ResponseID: responseID}
}

// Clone returns a fresh OpenAI client that shares credentials but not
// request state with the original.
func (o *OpenAI) Clone() Client {
	if o == nil {
		return nil
	}
	clone := &OpenAI{
		key:   o.key,
		model: o.model,
		client: &http.Client{
			Timeout: time.Duration(defaultHTTPTimeout) * time.Second,
		},
	}
	if o.Temperature != nil {
		v := *o.Temperature
		clone.Temperature = &v
	}
	return clone
}

// ResetConversation clears any stored response linkage so the next request starts fresh.
func (o *OpenAI) ResetConversation() {
	o.previousResponseID = ""
}

func (o *OpenAI) ModelName() string { return o.model }
func supportsTemperature(model string) bool {
	m := strings.ToLower(model)
	if strings.HasPrefix(m, "gpt-5") || strings.HasPrefix(m, "o1") || strings.HasPrefix(m, "o3") || strings.HasPrefix(m, "o4") {
		return false
	}
	return true
}
