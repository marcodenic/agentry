package model

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Anthropic client uses Anthropic's streaming messages API.
type Anthropic struct {
	key         string
	model       string
	temperature float64
	client      *http.Client
	// simple local rate limiter (approximate) per minute window
	mu          sync.Mutex
	windowStart time.Time
	windowTokens int
}

func NewAnthropic(key, model string) *Anthropic {
	return &Anthropic{key: key, model: model, client: http.DefaultClient}
}

// Stream implements proper Anthropic streaming API
func (a *Anthropic) Stream(ctx context.Context, msgs []ChatMessage, tools []ToolSpec) (<-chan StreamChunk, error) {
	if a.key == "" {
		return nil, errors.New("missing api key")
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
		if len(t.Parameters) > 0 {
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
		// Skip messages with empty content
		if strings.TrimSpace(m.Content) == "" {
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
		"stream":      true, // Enable streaming
	}
	if len(anTools) > 0 {
		reqBody["tools"] = anTools
	}
	if systemPrompt != "" {
		reqBody["system"] = systemPrompt
	}

	// Log context size
	totalChars := 0
	for _, msg := range anMsgs {
		totalChars += len(msg.Content)
	}

	b, _ := json.Marshal(reqBody)

	// crude token estimation (char/4) before sending for rate limiting
	estTokens := len(b) / 4
	a.mu.Lock()
	now := time.Now()
	if a.windowStart.IsZero() || now.Sub(a.windowStart) > time.Minute {
		a.windowStart = now
		a.windowTokens = 0
	}
	limitPerMin := 28000 // keep a safety margin below 30k
	if a.windowTokens+estTokens > limitPerMin {
		a.mu.Unlock()
		return nil, fmt.Errorf("anthropic local rate limiter: estimated tokens %d would exceed per-minute budget (%d/%d used)", estTokens, a.windowTokens, limitPerMin)
	}
	a.windowTokens += estTokens
	a.mu.Unlock()

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.key)
	req.Header.Set("anthropic-version", "2023-06-01")

	out := make(chan StreamChunk, 32)
	go func() {
		defer close(out)
		resp, err := a.client.Do(req)
		if err != nil {
			out <- StreamChunk{Err: err}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 300 {
			body, _ := io.ReadAll(resp.Body)
			if resp.StatusCode == 429 {
				out <- StreamChunk{Err: fmt.Errorf("anthropic API rate limit exceeded: %s", string(body))}
			} else {
				out <- StreamChunk{Err: errors.New(string(body))}
			}
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		var contentBuilder strings.Builder
		var toolCalls []ToolCall
		var currentToolCall *ToolCall
		inputTokens, outputTokens := 0, 0

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				break
			}

			var event struct {
				Type  string `json:"type"`
				Index int    `json:"index,omitempty"`
				Delta struct {
					Type        string `json:"type"`
					Text        string `json:"text"`
					PartialJson string `json:"partial_json"`
				} `json:"delta"`
				ContentBlock struct {
					Type  string          `json:"type"`
					ID    string          `json:"id"`
					Name  string          `json:"name"`
					Input json.RawMessage `json:"input"`
				} `json:"content_block"`
				Usage struct {
					InputTokens  int `json:"input_tokens"`
					OutputTokens int `json:"output_tokens"`
				} `json:"usage"`
			}

			if err := json.Unmarshal([]byte(data), &event); err != nil {
				continue
			}

			switch event.Type {
			case "content_block_delta":
				if event.Delta.Type == "text_delta" && event.Delta.Text != "" {
					contentBuilder.WriteString(event.Delta.Text)
					out <- StreamChunk{ContentDelta: event.Delta.Text}
				} else if event.Delta.Type == "input_json_delta" && currentToolCall != nil {
					// Accumulate tool arguments from streaming deltas
					currentToolCall.Arguments = append(currentToolCall.Arguments, []byte(event.Delta.PartialJson)...)
				}
			case "content_block_start":
				if event.ContentBlock.Type == "tool_use" {
					currentToolCall = &ToolCall{
						ID:        event.ContentBlock.ID,
						Name:      event.ContentBlock.Name,
						Arguments: json.RawMessage{}, // Start empty, will be filled by deltas
					}
				}
			case "content_block_stop":
				if currentToolCall != nil {
					toolCalls = append(toolCalls, *currentToolCall)
					currentToolCall = nil
				}

			case "message_delta":
				if event.Usage.InputTokens > 0 {
					inputTokens = event.Usage.InputTokens
				}
				if event.Usage.OutputTokens > 0 {
					outputTokens = event.Usage.OutputTokens
				}
			}
		}

		if err := scanner.Err(); err != nil {
			out <- StreamChunk{Err: err}
			return
		}

		// Send final response with tool calls and token usage; include model name via special terminal chunk
        out <- StreamChunk{ // final chunk
            Done:         true,
            ToolCalls:    toolCalls,
            InputTokens:  inputTokens,
            OutputTokens: outputTokens,
            ModelName:    "anthropic/" + a.model,
        }
	}()

	return out, nil
}

// ModelName returns the model name used by this Anthropic client
func (a *Anthropic) ModelName() string {
	return a.model
}
