package model

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Google client for Gemini models.
type Google struct {
	key         string
	model       string
	Temperature *float64
	client      *genai.Client
}

func NewGoogle(key, model string) *Google {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(key))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating Google client: %v\n", err)
		return nil
	}
	return &Google{key: key, model: model, client: client}
}

func (g *Google) Stream(ctx context.Context, msgs []ChatMessage, tools []ToolSpec) (<-chan StreamChunk, error) {
	if g.client == nil {
		return nil, fmt.Errorf("google client not initialized")
	}

	model := g.client.GenerativeModel(g.model)
	if g.Temperature != nil {
		t := float32(*g.Temperature)
		model.SetTemperature(t)
	}

	if len(tools) > 0 {
		model.Tools = []*genai.Tool{buildGoogleTools(tools)}
	}

	var history []*genai.Content
	var systemInstruction *genai.Content

	// Filter out system messages first
	var chatMsgs []ChatMessage
	for _, m := range msgs {
		if m.Role == "system" {
			// Concatenate multiple system messages if any
			if systemInstruction == nil {
				systemInstruction = genai.NewUserContent(genai.Text(m.Content))
			} else {
				systemInstruction.Parts = append(systemInstruction.Parts, genai.Text(m.Content))
			}
		} else {
			chatMsgs = append(chatMsgs, m)
		}
	}

	if systemInstruction != nil {
		model.SystemInstruction = systemInstruction
	}

	if len(chatMsgs) == 0 {
		return nil, fmt.Errorf("no messages to send")
	}

	lastMsg := chatMsgs[len(chatMsgs)-1]
	historyMsgs := chatMsgs[:len(chatMsgs)-1]

	for _, m := range historyMsgs {
		content := convertMessageToContent(m)
		if content != nil {
			history = append(history, content)
		}
	}

	cs := model.StartChat()
	cs.History = history

	// Prepare the parts for the last message
	lastParts := convertMessageToParts(lastMsg)
	if len(lastParts) == 0 {
		lastParts = []genai.Part{genai.Text(" ")}
	}

	iter := cs.SendMessageStream(ctx, lastParts...)
	return g.processStream(iter)
}

func convertMessageToContent(m ChatMessage) *genai.Content {
	role := "user"
	if m.Role == "assistant" {
		role = "model"
	} else if m.Role == "tool" {
		role = "function"
	}

	parts := convertMessageToParts(m)
	if len(parts) == 0 {
		return nil
	}

	return &genai.Content{
		Role:  role,
		Parts: parts,
	}
}

func convertMessageToParts(m ChatMessage) []genai.Part {
	var parts []genai.Part

	if m.Role == "tool" {
		var response map[string]any
		if err := json.Unmarshal([]byte(m.Content), &response); err != nil {
			response = map[string]any{"result": m.Content}
		}
		
		parts = append(parts, genai.FunctionResponse{
			Name:     m.Name,
			Response: response,
		})
		return parts
	}

	if m.Content != "" {
		parts = append(parts, genai.Text(m.Content))
	}

	if len(m.ToolCalls) > 0 {
		for _, tc := range m.ToolCalls {
			var args map[string]any
			if len(tc.Arguments) > 0 {
				_ = json.Unmarshal(tc.Arguments, &args)
			}
			parts = append(parts, genai.FunctionCall{
				Name: tc.Name,
				Args: args,
			})
		}
	}

	return parts
}

func (g *Google) processStream(iter *genai.GenerateContentResponseIterator) (<-chan StreamChunk, error) {
	out := make(chan StreamChunk)
	go func() {
		defer close(out)
		for {
			resp, err := iter.Next()
			if err != nil {
				if err == iterator.Done {
					break
				}
				break
			}
			
			for _, cand := range resp.Candidates {
				if cand.Content != nil {
					for _, part := range cand.Content.Parts {
						if txt, ok := part.(genai.Text); ok {
							out <- StreamChunk{
								ContentDelta: string(txt),
							}
						}
						if fn, ok := part.(genai.FunctionCall); ok {
							argsBytes, _ := json.Marshal(fn.Args)
							out <- StreamChunk{
								ToolCalls: []ToolCall{{
									Name:      fn.Name,
									Arguments: argsBytes,
									ID:        "call_" + fn.Name, 
								}},
							}
						}
					}
				}
			}
		}
		out <- StreamChunk{Done: true, ModelName: "google/" + g.model}
	}()
	return out, nil
}

func buildGoogleTools(tools []ToolSpec) *genai.Tool {
	var fds []*genai.FunctionDeclaration
	for _, t := range tools {
		fds = append(fds, &genai.FunctionDeclaration{
			Name:        t.Name,
			Description: t.Description,
			Parameters:  schemaFromMap(t.Parameters),
		})
	}
	return &genai.Tool{FunctionDeclarations: fds}
}

func schemaFromMap(m map[string]any) *genai.Schema {
	s := &genai.Schema{}
	if t, ok := m["type"].(string); ok {
		s.Type = typeFromString(t)
	}
	if d, ok := m["description"].(string); ok {
		s.Description = d
	}
	if props, ok := m["properties"].(map[string]any); ok {
		s.Properties = make(map[string]*genai.Schema)
		for k, v := range props {
			if vm, ok := v.(map[string]any); ok {
				s.Properties[k] = schemaFromMap(vm)
			}
		}
	}
	if req, ok := m["required"].([]any); ok {
		for _, r := range req {
			if rs, ok := r.(string); ok {
				s.Required = append(s.Required, rs)
			}
		}
	}
	if items, ok := m["items"].(map[string]any); ok {
		s.Items = schemaFromMap(items)
	}
	
	return s
}

func typeFromString(t string) genai.Type {
	switch t {
	case "string":
		return genai.TypeString
	case "number":
		return genai.TypeNumber
	case "integer":
		return genai.TypeInteger
	case "boolean":
		return genai.TypeBoolean
	case "array":
		return genai.TypeArray
	case "object":
		return genai.TypeObject
	default:
		return genai.TypeString
	}
}

func (g *Google) Clone() Client {
	return NewGoogle(g.key, g.model)
}

func (g *Google) ModelName() string { return g.model }
