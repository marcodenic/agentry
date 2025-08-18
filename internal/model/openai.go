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
)

// OpenAI client uses OpenAI's chat completion API.
type OpenAI struct {
	key         string
	model       string
	Temperature *float64
	client      *http.Client
}

// TODO: Replace manual implementation with official github.com/openai/openai-go SDK for
// structured streaming, usage metadata, and reduced maintenance burden.

func NewOpenAI(key, model string) *OpenAI {
	return &OpenAI{key: key, model: model, client: http.DefaultClient}
}

// Internal wire types reused for both non-stream and stream calls.
type openAITool struct {
	Type     string `json:"type"`
	Function struct {
		Name        string         `json:"name"`
		Description string         `json:"description,omitempty"`
		Parameters  map[string]any `json:"parameters"`
	} `json:"function"`
}

type oaToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type oaMessage struct {
	Role       string       `json:"role"`
	Content    string       `json:"content"`
	Name       string       `json:"name,omitempty"`
	ToolCallID string       `json:"tool_call_id,omitempty"`
	ToolCalls  []oaToolCall `json:"tool_calls,omitempty"`
}

func buildOATools(tools []ToolSpec) []openAITool {
	oa := make([]openAITool, len(tools))
	for i, t := range tools {
		oa[i].Type = "function"
		oa[i].Function.Name = t.Name
		oa[i].Function.Description = t.Description
		if len(t.Parameters) > 0 {
			oa[i].Function.Parameters = t.Parameters
		} else {
			oa[i].Function.Parameters = map[string]any{"type": "object", "properties": map[string]any{}}
		}
	}
	return oa
}

func buildOAMessages(msgs []ChatMessage) []oaMessage {
	oa := make([]oaMessage, len(msgs))
	for i, m := range msgs {
		oa[i].Role = m.Role
		oa[i].Content = m.Content
		oa[i].Name = m.Name
		oa[i].ToolCallID = m.ToolCallID
		if len(m.ToolCalls) > 0 {
			oa[i].ToolCalls = make([]oaToolCall, len(m.ToolCalls))
			for j, tc := range m.ToolCalls {
				oa[i].ToolCalls[j].ID = tc.ID
				oa[i].ToolCalls[j].Type = "function"
				oa[i].ToolCalls[j].Function.Name = tc.Name
				oa[i].ToolCalls[j].Function.Arguments = string(tc.Arguments)
			}
		}
	}
	return oa
}

func (o *OpenAI) buildRequest(ctx context.Context, msgs []ChatMessage, tools []ToolSpec, stream bool) (*http.Request, error) {
	body := map[string]any{
		"model":       o.model,
		"messages":    buildOAMessages(msgs),
		"tools":       buildOATools(tools),
		"tool_choice": "auto",
	}
	if stream {
		body["stream"] = true
	}
	if o.Temperature != nil && supportsTemperature(o.model) {
		body["temperature"] = *o.Temperature
	}
	b, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.key)
	return req, nil
}

func (o *OpenAI) Complete(ctx context.Context, msgs []ChatMessage, tools []ToolSpec) (Completion, error) {
	if o.key == "" {
		return Completion{}, errors.New("missing api key")
	}
	req, err := o.buildRequest(ctx, msgs, tools, false)
	if err != nil {
		return Completion{}, err
	}
	resp, err := o.client.Do(req)
	if err != nil {
		return Completion{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return Completion{}, errors.New(string(body))
	}
	var res struct {
		Choices []struct {
			Message struct {
				Content   string       `json:"content"`
				ToolCalls []oaToolCall `json:"tool_calls"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return Completion{}, err
	}
	if len(res.Choices) == 0 {
		return Completion{}, errors.New("no choices")
	}
	msg := res.Choices[0].Message
	comp := Completion{Content: msg.Content, InputTokens: res.Usage.PromptTokens, OutputTokens: res.Usage.CompletionTokens, ModelName: "openai/" + o.model}
	for _, tc := range msg.ToolCalls {
		comp.ToolCalls = append(comp.ToolCalls, ToolCall{ID: tc.ID, Name: tc.Function.Name, Arguments: []byte(tc.Function.Arguments)})
	}
	return comp, nil
}

// Stream implements incremental streaming for OpenAI Chat Completions API using SSE.
// NOTE: This currently targets the older /v1/chat/completions streaming format.
func (o *OpenAI) Stream(ctx context.Context, msgs []ChatMessage, tools []ToolSpec) (<-chan StreamChunk, error) {
	if o.key == "" {
		return nil, errors.New("missing api key")
	}
	req, err := o.buildRequest(ctx, msgs, tools, true)
	if err != nil {
		return nil, err
	}
	out := make(chan StreamChunk, 32)
	go func() {
		defer close(out)
		resp, err := o.client.Do(req)
		if err != nil {
			out <- StreamChunk{Err: err}
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 300 {
			body, _ := io.ReadAll(resp.Body)
			out <- StreamChunk{Err: errors.New(string(body))}
			return
		}
		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		// Accumulate tool call deltas by index; OpenAI streams partial name/argument pieces.
		type partialToolCall struct {
			ToolCall
			index int
		}
		// index -> partial
		partials := map[int]*partialToolCall{}
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				out <- StreamChunk{Err: ctx.Err()}
				return
			default:
			}
			line := scanner.Text()
			if !strings.HasPrefix(line, "data:") {
				continue
			}
			payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if payload == "[DONE]" {
				// Early termination: assemble partials collected so far.
				if len(partials) > 0 {
					idxs := make([]int, 0, len(partials))
					for i := range partials { idxs = append(idxs, i) }
					sort.Ints(idxs)
					final := make([]ToolCall, 0, len(partials))
					for _, i := range idxs { final = append(final, partials[i].ToolCall) }
					out <- StreamChunk{Done: true, ToolCalls: final}
				} else {
					out <- StreamChunk{Done: true}
				}
				return
			}
			var evt struct {
				Choices []struct {
					Delta struct {
						Content   string `json:"content"`
						ToolCalls []struct {
							Index int    `json:"index"`
							ID    string `json:"id"`
							Type  string `json:"type"`
							Function struct {
								Name      string `json:"name"`
								Arguments string `json:"arguments"`
							} `json:"function"`
						} `json:"tool_calls"`
					} `json:"delta"`
				} `json:"choices"`
			}
			if err := json.Unmarshal([]byte(payload), &evt); err != nil {
				continue
			}
			if len(evt.Choices) == 0 {
				continue
			}
			d := evt.Choices[0].Delta
			if len(d.ToolCalls) > 0 {
				for _, tc := range d.ToolCalls {
					// Retrieve or create partial by index.
					p := partials[tc.Index]
					if p == nil {
						p = &partialToolCall{index: tc.Index}
						partials[tc.Index] = p
					}
					if tc.ID != "" { // ID only appears once usually.
						p.ID = tc.ID
					}
					if tc.Function.Name != "" { // Name appears on first delta.
						p.Name = tc.Function.Name
					}
					if tc.Function.Arguments != "" { // Arguments stream piecewise.
						p.Arguments = append(p.Arguments, []byte(tc.Function.Arguments)...)
					}
				}
			}
			if d.Content != "" {
				out <- StreamChunk{ContentDelta: d.Content}
			}
		}
		if err := scanner.Err(); err != nil {
			out <- StreamChunk{Err: err}
		} else {
			// Consolidate ordered tool calls with fully aggregated arguments.
			if len(partials) > 0 {
				idxs := make([]int, 0, len(partials))
				for i := range partials { idxs = append(idxs, i) }
				sort.Ints(idxs)
				final := make([]ToolCall, 0, len(partials))
				for _, i := range idxs {
					p := partials[i]
					final = append(final, p.ToolCall)
				}
				out <- StreamChunk{Done: true, ToolCalls: final}
			} else {
				out <- StreamChunk{Done: true}
			}
		}
	}()
	return out, nil
}

// ModelName returns the model name used by this OpenAI client
func (o *OpenAI) ModelName() string {
	return o.model
}

// supportsTemperature returns whether this model supports the temperature parameter.
// Some reasoning-oriented models (e.g., gpt-5, o1 family) don’t accept temperature.
func supportsTemperature(model string) bool {
	m := strings.ToLower(model)
	// Blocklist known families that reject temperature
	if strings.HasPrefix(m, "gpt-5") || strings.HasPrefix(m, "o1") || strings.HasPrefix(m, "o3") || strings.HasPrefix(m, "o4") {
		return false
	}
	return true
}
