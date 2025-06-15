package model

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/marcodenic/agentry/internal/trace"
)

// OpenAI client uses OpenAI's chat completion API.
const defaultMaxTokens = 120

type OpenAI struct {
	key         string
	temperature float64
	maxTokens   int
	client      *http.Client
}

func NewOpenAI(key string) *OpenAI {
	return &OpenAI{key: key, temperature: 0.7, maxTokens: defaultMaxTokens, client: http.DefaultClient}
}

func (o *OpenAI) SetTemperature(t float64) { o.temperature = t }

func (o *OpenAI) SetMaxTokens(t int) { o.maxTokens = t }

func (o *OpenAI) Complete(ctx context.Context, msgs []ChatMessage, tools []ToolSpec) (Completion, error) {
	if o.key == "" {
		return Completion{}, errors.New("missing api key")
	}

	type openAITool struct {
		Type     string `json:"type"`
		Function struct {
			Name        string         `json:"name"`
			Description string         `json:"description,omitempty"`
			Parameters  map[string]any `json:"parameters"`
		} `json:"function"`
	}

	oaTools := make([]openAITool, len(tools))
	for i, t := range tools {
		oaTools[i].Type = "function"
		oaTools[i].Function.Name = t.Name
		oaTools[i].Function.Description = t.Description
		if t.Parameters != nil && len(t.Parameters) > 0 {
			oaTools[i].Function.Parameters = t.Parameters
		} else {
			oaTools[i].Function.Parameters = map[string]any{"type": "object", "properties": map[string]any{}}
		}
	}

	type oaMessage struct {
		Role       string `json:"role"`
		Content    string `json:"content"`
		Name       string `json:"name,omitempty"`
		ToolCallID string `json:"tool_call_id,omitempty"`
		ToolCalls  []struct {
			ID       string `json:"id"`
			Type     string `json:"type"`
			Function struct {
				Name      string `json:"name"`
				Arguments string `json:"arguments"`
			} `json:"function"`
		} `json:"tool_calls,omitempty"`
	}

	oaMsgs := make([]oaMessage, len(msgs))
	for i, m := range msgs {
		oaMsgs[i].Role = m.Role
		oaMsgs[i].Content = m.Content
		oaMsgs[i].Name = m.Name
		oaMsgs[i].ToolCallID = m.ToolCallID
		if len(m.ToolCalls) > 0 {
			oaMsgs[i].ToolCalls = make([]struct {
				ID       string `json:"id"`
				Type     string `json:"type"`
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			}, len(m.ToolCalls))
			for j, tc := range m.ToolCalls {
				oaMsgs[i].ToolCalls[j].ID = tc.ID
				oaMsgs[i].ToolCalls[j].Type = "function"
				oaMsgs[i].ToolCalls[j].Function.Name = tc.Name
				oaMsgs[i].ToolCalls[j].Function.Arguments = string(tc.Arguments)
			}
		}
	}

	reqBody := map[string]any{
		"model":       "gpt-4o",
		"messages":    oaMsgs,
		"tools":       oaTools,
		"tool_choice": "auto",
		"temperature": o.temperature,
		"max_tokens":  o.maxTokens,
		"stream":      true,
	}

	b, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(b))
	if err != nil {
		return Completion{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.key)
	resp, err := o.client.Do(req)
	if err != nil {
		return Completion{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return Completion{}, errors.New(string(body))
	}

	tr := trace.WriterFrom(ctx)
	var final strings.Builder
	var toolCalls []ToolCall
	sc := bufio.NewScanner(resp.Body)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "[DONE]" {
			break
		}
		var chunk struct {
			Choices []struct {
				Delta struct {
					Content   string `json:"content"`
					ToolCalls []struct {
						ID       string `json:"id"`
						Type     string `json:"type"`
						Function struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						} `json:"function"`
					} `json:"tool_calls"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			continue
		}
		if len(chunk.Choices) == 0 {
			continue
		}
		delta := chunk.Choices[0].Delta
		if delta.Content != "" {
			final.WriteString(delta.Content)
			if tr != nil {
				tr.Write(ctx, trace.Event{Type: trace.EventToken, Data: delta.Content, Timestamp: trace.Now()})
			}
		}
		for _, tc := range delta.ToolCalls {
			toolCalls = append(toolCalls, ToolCall{
				ID:        tc.ID,
				Name:      tc.Function.Name,
				Arguments: []byte(tc.Function.Arguments),
			})
		}
	}
	if err := sc.Err(); err != nil {
		return Completion{}, err
	}
	comp := Completion{Content: final.String(), ToolCalls: toolCalls}
	return comp, nil
}
