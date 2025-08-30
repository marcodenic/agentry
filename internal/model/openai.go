package model

import (
    "bufio"
    "bytes"
    "context"
    "encoding/json"
    "errors"
    "io"
    "net/http"
    "sort"
    "strings"

    "github.com/marcodenic/agentry/internal/debug"
)

// OpenAI client implemented against /v1/responses (legacy chat completions removed).
type OpenAI struct {
	key         string
	model       string
	Temperature *float64
	client      *http.Client
}

func NewOpenAI(key, model string) *OpenAI {
	return &OpenAI{key: key, model: model, client: http.DefaultClient}
}

// Wire types for Responses API
type oaInputItem struct {
	Role    string          `json:"role"`
	Content []oaContentPart `json:"content"`
}
type oaContentPart struct {
	Type string `json:"type"`
	Text string `json:"text"`
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
	out := make([]oaInputItem, len(msgs))
	for i, m := range msgs {
		// Map role to supported values
		role := m.Role
		switch m.Role {
		case "tool":
			role = "user" // Convert tool results to user messages
		}
		out[i].Role = role

		// Use appropriate content type based on message role
		var contentType string
		switch m.Role {
		case "assistant":
			contentType = "output_text"
		case "user", "system", "tool":
			contentType = "input_text"
		default:
			contentType = "input_text" // Default fallback
		}
		out[i].Content = []oaContentPart{{Type: contentType, Text: m.Content}}
	}
	return out
}

func (o *OpenAI) buildRequest(ctx context.Context, msgs []ChatMessage, tools []ToolSpec, stream bool) (*http.Request, error) {
	if o.key == "" {
		return nil, errors.New("missing api key")
	}

	body := map[string]any{"model": o.model, "input": buildOAInput(msgs)}
	if len(tools) > 0 {
		body["tools"] = buildOATools(tools)
	}
	if stream {
		body["stream"] = true
	}
	if o.Temperature != nil && supportsTemperature(o.model) {
		body["temperature"] = *o.Temperature
	}

	b, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/responses", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.key)
	return req, nil
}

func (o *OpenAI) Stream(ctx context.Context, msgs []ChatMessage, tools []ToolSpec) (<-chan StreamChunk, error) {
    debug.Printf("OpenAI.Stream: msgs=%d tools=%d", len(msgs), len(tools))

    req, err := o.buildRequest(ctx, msgs, tools, true)
    debug.Printf("OpenAI.Stream: buildRequest err=%v", err)
    if err != nil {
        debug.Printf("OpenAI.Stream: buildRequest failed: %v", err)
        return nil, err
    }
    debug.Printf("OpenAI.Stream: starting HTTP request goroutine")
    out := make(chan StreamChunk, 32)
    go func() {
        defer close(out)
        debug.Printf("OpenAI.Stream: HTTP request...")
        resp, err := o.client.Do(req)
        debug.Printf("OpenAI.Stream: HTTP response err=%v", err)
        if err != nil {
            debug.Printf("OpenAI.Stream: HTTP request failed: %v", err)
            out <- StreamChunk{Err: err}
            return
        }
        defer resp.Body.Close()
        debug.Printf("OpenAI.Stream: status=%d", resp.StatusCode)
        if resp.StatusCode >= 300 {
            body, _ := io.ReadAll(resp.Body)
            out <- StreamChunk{Err: errors.New(string(body))}
            return
        }
		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		partials := map[int]*partial{}
		responseCalls := map[string]*partial{} // Track Responses API function calls by item_id
		var inTok, outTok int
		for scanner.Scan() {
			if ctx.Err() != nil {
				out <- StreamChunk{Err: ctx.Err()}
				return
			}
			line := scanner.Text()
			if !strings.HasPrefix(line, "data:") {
				continue
			}
			payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
            debug.Printf("OpenAI.Stream: payload=%q", payload)
            if payload == "[DONE]" {
                debug.Printf("OpenAI.Stream: [DONE], finalize (partials=%d responseCalls=%d)", len(partials), len(responseCalls))
                finalizeOpenAI(partials, out, inTok, outTok)
                return
            }
			if payload == "" {
				continue
			}
			var env map[string]any
			if err := json.Unmarshal([]byte(payload), &env); err != nil {
				continue
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
                            }
                        }
                    }
                }
			// Responses API: Handle function call argument deltas
			case t == "response.function_call_arguments.delta":
				if itemID, _ := env["item_id"].(string); itemID != "" {
					if p, exists := responseCalls[itemID]; exists {
						if delta, _ := env["delta"].(string); delta != "" {
							// Filter out whitespace-only deltas
							trimmed := strings.TrimSpace(delta)
							if trimmed != "" {
								p.Arguments = append(p.Arguments, []byte(delta)...)
							}
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
                        debug.Printf("OpenAI.Stream: Function call args done for %s", itemID)
                    }
                }
            case t == "response.completed":
                debug.Printf("OpenAI.Stream: response completed")
                if u, ok := env["usage"].(map[string]any); ok {
                    if iv, ok := u["input_tokens"].(float64); ok {
                        inTok = int(iv)
                    }
                    if ov, ok := u["output_tokens"].(float64); ok {
                        outTok = int(ov)
                    }
                }
                // Combine both legacy and responses API calls
                finalizeWithResponses(partials, responseCalls, out, inTok, outTok)
                return
            default: /* ignore */
            }
        }
        if err := scanner.Err(); err != nil {
            debug.Printf("OpenAI.Stream: scanner error: %v", err)
            out <- StreamChunk{Err: err}
        } else {
            debug.Printf("OpenAI.Stream: scanner ended, finalize (partials=%d responseCalls=%d)", len(partials), len(responseCalls))
            finalizeWithResponses(partials, responseCalls, out, inTok, outTok)
        }
    }()
    return out, nil
}

func finalizeOpenAI(partials map[int]*partial, out chan<- StreamChunk, inTok, outTok int) {
    if len(partials) == 0 {
        // No tool calls: emit final chunk with usage
        out <- StreamChunk{Done: true, InputTokens: inTok, OutputTokens: outTok, ModelName: "openai"}
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
	out <- StreamChunk{Done: true, ToolCalls: final, InputTokens: inTok, OutputTokens: outTok, ModelName: "openai"}
}

func finalizeWithResponses(partials map[int]*partial, responseCalls map[string]*partial, out chan<- StreamChunk, inTok, outTok int) {
	final := make([]ToolCall, 0, len(partials)+len(responseCalls))
	// Add legacy format calls
	if len(partials) > 0 {
		idxs := make([]int, 0, len(partials))
		for i := range partials {
			idxs = append(idxs, i)
		}
		sort.Ints(idxs)
		for _, i := range idxs {
			p := partials[i]
			final = append(final, ToolCall{ID: p.ID, Name: p.Name, Arguments: p.Arguments})
		}
	}
	// Add Responses API calls
	for _, p := range responseCalls {
		final = append(final, ToolCall{ID: p.ID, Name: p.Name, Arguments: p.Arguments})
	}
	out <- StreamChunk{Done: true, ToolCalls: final, InputTokens: inTok, OutputTokens: outTok, ModelName: "openai"}
}

func (o *OpenAI) ModelName() string { return o.model }
func supportsTemperature(model string) bool {
	m := strings.ToLower(model)
	if strings.HasPrefix(m, "gpt-5") || strings.HasPrefix(m, "o1") || strings.HasPrefix(m, "o3") || strings.HasPrefix(m, "o4") {
		return false
	}
	return true
}
