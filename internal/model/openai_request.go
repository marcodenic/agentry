package model

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/marcodenic/agentry/internal/debug"
)

type oaRequestBuilder struct {
	client *OpenAI
	msgs   []ChatMessage
	tools  []ToolSpec
}

func newOARequestBuilder(o *OpenAI, msgs []ChatMessage, tools []ToolSpec) *oaRequestBuilder {
	return &oaRequestBuilder{client: o, msgs: msgs, tools: tools}
}

func (b *oaRequestBuilder) Build(ctx context.Context, stream bool) (*http.Request, error) {
	o := b.client
	if o.key == "" {
		return nil, errors.New("missing api key")
	}

	body := map[string]any{}
	endpoint := "https://api.openai.com/v1/responses"

	fnOutputs := b.functionOutputs()
	if len(fnOutputs) > 0 {
		body["model"] = o.model
		if o.previousResponseID != "" {
			body["previous_response_id"] = o.previousResponseID
		} else {
			debug.Printf("OpenAIConversation.buildRequest: missing previous_response_id for tool outputs; proceeding without linkage")
		}
		body["input"] = fnOutputs
		if stream {
			body["stream"] = true
		}
	} else {
		body["model"] = o.model
		body["input"] = buildOAInput(b.msgs)
		if len(b.tools) > 0 {
			body["tools"] = buildOATools(b.tools)
			body["tool_choice"] = "auto"
		}
		if stream {
			body["stream"] = true
		}
		if o.previousResponseID == "" {
			debug.Printf("OpenAIConversation.buildRequest: No previous response ID available, starting new conversation")
		}
	}
	if o.Temperature != nil && supportsTemperature(o.model) {
		body["temperature"] = *o.Temperature
	}

	payload, _ := json.Marshal(body)
	debug.Printf("OpenAIConversation.buildRequest: Request body: %s", string(payload))

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("OpenAI-Beta", "responses=v1")
	req.Header.Set("Authorization", "Bearer "+o.key)
	return req, nil
}

func (b *oaRequestBuilder) functionOutputs() []map[string]any {
	o := b.client
	if o.previousResponseID == "" {
		return nil
	}
	pending := make([]ChatMessage, 0)
	for i := len(b.msgs) - 1; i >= 0; i-- {
		m := b.msgs[i]
		if m.Role != "tool" {
			break
		}
		pending = append(pending, m)
	}
	if len(pending) == 0 {
		return nil
	}
	for i := 0; i < len(pending)/2; i++ {
		j := len(pending) - 1 - i
		pending[i], pending[j] = pending[j], pending[i]
	}
	outputs := make([]map[string]any, 0, len(pending))
	for _, m := range pending {
		if strings.TrimSpace(m.ToolCallID) == "" || strings.TrimSpace(m.Content) == "" {
			continue
		}
		outputs = append(outputs, map[string]any{
			"type":    "function_call_output",
			"call_id": m.ToolCallID,
			"output":  m.Content,
		})
	}
	return outputs
}
