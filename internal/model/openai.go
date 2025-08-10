package model

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

// OpenAI client uses OpenAI's chat completion API.
type OpenAI struct {
	key         string
	model       string
	Temperature *float64
	client      *http.Client
}

func NewOpenAI(key, model string) *OpenAI {
	return &OpenAI{key: key, model: model, client: http.DefaultClient}
}

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
		"model":       o.model,
		"messages":    oaMsgs,
		"tools":       oaTools,
		"tool_choice": "auto",
	}

	// Only include temperature when explicitly set and supported by the model
	if o.Temperature != nil && supportsTemperature(o.model) {
		reqBody["temperature"] = *o.Temperature
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
	var res struct {
		Choices []struct {
			Message struct {
				Content   string `json:"content"`
				ToolCalls []struct {
					ID       string `json:"id"`
					Function struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					} `json:"function"`
				} `json:"tool_calls"`
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

	choice := res.Choices[0].Message
	comp := Completion{
		Content:      choice.Content,
		InputTokens:  res.Usage.PromptTokens,
		OutputTokens: res.Usage.CompletionTokens,
		ModelName:    "openai/" + o.model, // Store as provider/model format
	}
	for _, tc := range choice.ToolCalls {
		comp.ToolCalls = append(comp.ToolCalls, ToolCall{
			ID:        tc.ID,
			Name:      tc.Function.Name,
			Arguments: []byte(tc.Function.Arguments),
		})
	}
	return comp, nil
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
