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

// Anthropic client uses Anthropic's messages API.
type Anthropic struct {
	key         string
	model       string
	temperature float64
	client      *http.Client
}

func NewAnthropic(key, model string) *Anthropic {
	return &Anthropic{key: key, model: model, client: http.DefaultClient}
}

func (a *Anthropic) Complete(ctx context.Context, msgs []ChatMessage, tools []ToolSpec) (Completion, error) {
	if a.key == "" {
		return Completion{}, errors.New("missing api key")
	}

	type anthropicTool struct {
		Name        string         `json:"name"`
		Description string         `json:"description,omitempty"`
		InputSchema map[string]any `json:"input_schema"`
	}

	anTools := make([]anthropicTool, len(tools))
	for i, t := range tools {
		anTools[i].Name = t.Name
		anTools[i].Description = t.Description
		if t.Parameters != nil && len(t.Parameters) > 0 {
			anTools[i].InputSchema = t.Parameters
		} else {
			anTools[i].InputSchema = map[string]any{"type": "object", "properties": map[string]any{}}
		}
	}

	type anMessage struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	anMsgs := make([]anMessage, 0, len(msgs))
	var systemPrompt string
	for _, m := range msgs {
		if m.Role == "system" {
			systemPrompt = m.Content
			continue
		}
		// Anthropic only supports user and assistant roles
		role := m.Role
		if role != "user" && role != "assistant" {
			role = "user" // default to user for other roles like tool
		}
		anMsgs = append(anMsgs, anMessage{Role: role, Content: m.Content})
	}

	reqBody := map[string]any{
		"model":       a.model,
		"messages":    anMsgs,
		"max_tokens":  4096,
		"temperature": a.temperature,
	}
	if len(anTools) > 0 {
		reqBody["tools"] = anTools
	}
	if systemPrompt != "" {
		reqBody["system"] = systemPrompt
	}

	b, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(b))
	if err != nil {
		return Completion{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.key)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := a.client.Do(req)
	if err != nil {
		return Completion{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return Completion{}, errors.New(string(body))
	}

	var res struct {
		Content []struct {
			Type  string          `json:"type"`
			Text  string          `json:"text,omitempty"`
			ID    string          `json:"id,omitempty"`
			Name  string          `json:"name,omitempty"`
			Input json.RawMessage `json:"input,omitempty"`
		} `json:"content"`
		StopReason string `json:"stop_reason"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return Completion{}, err
	}

	comp := Completion{}
	var content strings.Builder
	for _, c := range res.Content {
		switch c.Type {
		case "text":
			content.WriteString(c.Text)
		case "tool_use":
			comp.ToolCalls = append(comp.ToolCalls, ToolCall{
				ID:        c.ID,
				Name:      c.Name,
				Arguments: c.Input,
			})
		}
	}
	comp.Content = content.String()

	return comp, nil
}

// ModelName returns the model name used by this Anthropic client
func (a *Anthropic) ModelName() string {
	return a.model
}
