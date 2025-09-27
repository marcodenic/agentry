package model

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
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

func (o *OpenAI) buildRequest(ctx context.Context, msgs []ChatMessage, tools []ToolSpec, stream bool) (*http.Request, error) {
	builder := newOARequestBuilder(o, msgs, tools)
	return builder.Build(ctx, stream)
}

func (o *OpenAI) Stream(ctx context.Context, msgs []ChatMessage, tools []ToolSpec) (<-chan StreamChunk, error) {
	startTime := time.Now()
	debug.Printf("OpenAI.Stream: START msgs=%d tools=%d", len(msgs), len(tools))

	// Log detailed model interaction start
	debug.LogModelInteraction("openai", o.model, len(msgs), map[string]int{"input_estimated": estimateTokens(msgs)}, 0)

	// Always use the create endpoint; when continuing, buildRequest will include previous_response_id
	var req *http.Request
	var err error
	req, err = o.buildRequest(ctx, msgs, tools, true)
	debug.Printf("OpenAI.Stream: buildRequest err=%v elapsed=%v", err, time.Since(startTime))
	if err != nil {
		debug.Printf("OpenAI.Stream: buildRequest failed: %v", err)
		debug.LogEvent("MODEL", "request_build_failed", map[string]interface{}{
			"provider": "openai",
			"model":    o.model,
			"error":    err.Error(),
		})
		return nil, err
	}

	// Check context timeout settings
	if deadline, ok := ctx.Deadline(); ok {
		timeoutDur := time.Until(deadline)
		debug.Printf("OpenAI.Stream: context timeout set to %v", timeoutDur)
	} else {
		debug.Printf("OpenAI.Stream: no context timeout set")
	}

	debug.Printf("OpenAI.Stream: starting HTTP request goroutine, elapsed=%v", time.Since(startTime))
	out := make(chan StreamChunk, 32)
	go func() {
		defer close(out)
		reqStartTime := time.Now()
		debug.Printf("OpenAI.Stream: HTTP request starting... elapsed=%v", time.Since(startTime))

		resp, err := o.client.Do(req)
		reqEndTime := time.Now()
		debug.Printf("OpenAI.Stream: HTTP response received err=%v request_duration=%v total_elapsed=%v", err, reqEndTime.Sub(reqStartTime), reqEndTime.Sub(startTime))

		if err != nil {
			debug.Printf("OpenAI.Stream: HTTP request failed: %v", err)
			debug.LogEvent("MODEL", "http_request_failed", map[string]interface{}{
				"provider": "openai",
				"model":    o.model,
				"error":    err.Error(),
				"duration": reqEndTime.Sub(reqStartTime).String(),
			})
			out <- StreamChunk{Err: err}
			return
		}
		defer resp.Body.Close()
		debug.Printf("OpenAI.Stream: status=%d headers=%v", resp.StatusCode, resp.Header)
		if resp.StatusCode >= 300 {
			body, _ := io.ReadAll(resp.Body)
			debug.LogEvent("MODEL", "api_error", map[string]interface{}{
				"provider": "openai",
				"model":    o.model,
				"status":   resp.StatusCode,
				"response": string(body),
			})
			out <- StreamChunk{Err: errors.New(string(body))}
			return
		}
		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		partials := map[int]*partial{}
		responseCalls := map[string]*partial{} // Track Responses API function calls by item_id
		var inTok, outTok int
		var responseID string
		scanStartTime := time.Now()
		lineCount := 0
		debug.Printf("OpenAI.Stream: starting stream scan, elapsed=%v", time.Since(startTime))

		for scanner.Scan() {
			lineCount++
			if lineCount%100 == 0 {
				debug.Printf("OpenAI.Stream: processed %d lines, elapsed=%v", lineCount, time.Since(startTime))
			}

			if ctx.Err() != nil {
				debug.Printf("OpenAI.Stream: context error during scan: %v, elapsed=%v", ctx.Err(), time.Since(startTime))
				debug.LogEvent("MODEL", "context_error", map[string]interface{}{
					"provider":        "openai",
					"model":           o.model,
					"error":           ctx.Err().Error(),
					"lines_processed": lineCount,
				})
				out <- StreamChunk{Err: ctx.Err()}
				return
			}
			line := scanner.Text()
			if !strings.HasPrefix(line, "data:") {
				continue
			}
			payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			// debug.Printf("OpenAI.Stream: payload=%q", payload) // Disabled: too verbose
			if payload == "[DONE]" {
				debug.Printf("OpenAI.Stream: [DONE], finalize (partials=%d responseCalls=%d) scan_duration=%v total_elapsed=%v lines=%d", len(partials), len(responseCalls), time.Since(scanStartTime), time.Since(startTime), lineCount)
				if responseID != "" {
					o.previousResponseID = responseID
					debug.Printf("OpenAI.Stream: Persisting response ID for next request: %s", responseID)
				}
				// Log successful completion
				debug.LogModelInteraction("openai", o.model, len(msgs), map[string]int{
					"input_tokens":  inTok,
					"output_tokens": outTok,
				}, time.Since(startTime))
				finalizeOpenAI(partials, out, inTok, outTok, o.model, responseID)
				return
			}
			if payload == "" {
				continue
			}
			var env map[string]any
			if err := json.Unmarshal([]byte(payload), &env); err != nil {
				continue
			}
			// Try to capture response identifier from any event that carries it
			if v, ok := env["response_id"].(string); ok && v != "" {
				responseID = v
				debug.Printf("OpenAI.Stream: Captured response_id: %s", responseID)
			}
			if respObj, ok := env["response"].(map[string]any); ok {
				if v, ok := respObj["id"].(string); ok && v != "" {
					responseID = v
					debug.Printf("OpenAI.Stream: Captured response.id: %s", responseID)
				}
			}
			t, _ := env["type"].(string)
			switch {
			case strings.HasSuffix(t, ".delta") && strings.Contains(t, "output_text"):
				if d, ok := env["delta"].(string); ok && d != "" {
					out <- StreamChunk{ContentDelta: d}
				}
			case strings.HasSuffix(t, ".delta") && strings.Contains(t, "tool_calls"):
				if arr, ok := env["tool_calls"].([]any); ok {
					for _, v := range arr {
						if m, ok := v.(map[string]any); ok {
							idx := 0
							if iv, ok := m["index"].(float64); ok {
								idx = int(iv)
							}
							p := partials[idx]
							if p == nil {
								p = &partial{index: idx}
								partials[idx] = p
							}
							if id, _ := m["id"].(string); id != "" {
								p.ID = id
							}
							// flattened fields
							if name, _ := m["name"].(string); name != "" {
								p.Name = name
							}
							if args, _ := m["arguments"].(string); args != "" {
								p.Arguments = append(p.Arguments, []byte(args)...)
							}
							// nested legacy fallback
							if fn, ok := m["function"].(map[string]any); ok {
								if name, _ := fn["name"].(string); name != "" {
									p.Name = name
								}
								if args, _ := fn["arguments"].(string); args != "" {
									p.Arguments = append(p.Arguments, []byte(args)...)
								}
							}
						}
					}
				}
				// Responses API: Handle function call start
			case t == "response.output_item.added":
				if item, ok := env["item"].(map[string]any); ok {
					if itemType, _ := item["type"].(string); itemType == "function_call" {
						if itemID, _ := item["id"].(string); itemID != "" {
							if name, _ := item["name"].(string); name != "" {
								callID, _ := item["call_id"].(string)
								p := &partial{ToolCall: ToolCall{ID: callID, Name: name}, index: 0}
								responseCalls[itemID] = p
								debug.Printf("OpenAI.Stream: Added function call %s/%s", itemID, name)
								debug.LogEvent("MODEL", "tool_call_start", map[string]interface{}{
									"provider":  "openai",
									"model":     o.model,
									"item_id":   itemID,
									"call_id":   callID,
									"tool_name": name,
								})
							}
						}
					}
				}
			// Responses API: Handle function call argument deltas
			case t == "response.function_call_arguments.delta":
				if itemID, _ := env["item_id"].(string); itemID != "" {
					if p, exists := responseCalls[itemID]; exists {
						if delta, _ := env["delta"].(string); delta != "" {
							// Always include deltas - whitespace may be important for JSON formatting
							p.Arguments = append(p.Arguments, []byte(delta)...)
						}
					}
				}
			// Responses API: Handle final arguments
			case t == "response.function_call_arguments.done":
				if itemID, _ := env["item_id"].(string); itemID != "" {
					if p, exists := responseCalls[itemID]; exists {
						if args, _ := env["arguments"].(string); args != "" {
							p.Arguments = []byte(args)
						}
						debug.Printf("OpenAI.Stream: Function call args done for %s, args length: %d", itemID, len(p.Arguments))
						debug.LogEvent("MODEL", "tool_call_complete", map[string]interface{}{
							"provider":    "openai",
							"model":       o.model,
							"item_id":     itemID,
							"tool_name":   p.Name,
							"args_length": len(p.Arguments),
						})
					}
				}
			case t == "response.completed":
				debug.Printf("OpenAI.Stream: response completed, elapsed=%v", time.Since(startTime))
				if u, ok := env["usage"].(map[string]any); ok {
					if iv, ok := u["input_tokens"].(float64); ok {
						inTok = int(iv)
					}
					if ov, ok := u["output_tokens"].(float64); ok {
						outTok = int(ov)
					}
				}
				// Persist response ID for subsequent requests if present
				if responseID != "" {
					o.previousResponseID = responseID
					debug.Printf("OpenAI.Stream: Persisting response ID for next request: %s", responseID)
				}
				// Log successful completion
				debug.LogModelInteraction("openai", o.model, len(msgs), map[string]int{
					"input_tokens":  inTok,
					"output_tokens": outTok,
				}, time.Since(startTime))
				finalizeWithResponses(partials, responseCalls, out, inTok, outTok, o.model, responseID)
				return
			default: /* ignore */
			}
		}
		scanEndTime := time.Now()
		if err := scanner.Err(); err != nil {
			debug.Printf("OpenAI.Stream: scanner error: %v scan_duration=%v total_elapsed=%v lines=%d", err, scanEndTime.Sub(scanStartTime), scanEndTime.Sub(startTime), lineCount)
			debug.LogEvent("MODEL", "scanner_error", map[string]interface{}{
				"provider":        "openai",
				"model":           o.model,
				"error":           err.Error(),
				"lines_processed": lineCount,
			})
			out <- StreamChunk{Err: err}
		} else {
			debug.Printf("OpenAI.Stream: scanner ended normally, finalize (partials=%d responseCalls=%d) scan_duration=%v total_elapsed=%v lines=%d", len(partials), len(responseCalls), scanEndTime.Sub(scanStartTime), scanEndTime.Sub(startTime), lineCount)
			if responseID != "" {
				o.previousResponseID = responseID
				debug.Printf("OpenAI.Stream: Persisting response ID for next request: %s", responseID)
			}
			finalizeWithResponses(partials, responseCalls, out, inTok, outTok, o.model, responseID)
		}
	}()
	return out, nil
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
