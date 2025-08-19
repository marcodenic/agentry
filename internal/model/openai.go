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

// OpenAI client implemented against /v1/responses (legacy chat completions removed).
type OpenAI struct {
	key         string
	model       string
	Temperature *float64
	client      *http.Client
}

func NewOpenAI(key, model string) *OpenAI { return &OpenAI{key: key, model: model, client: http.DefaultClient} }

// Wire types for Responses API
type oaInputItem struct { Role string `json:"role"`; Content []oaContentPart `json:"content"` }
type oaContentPart struct { Type string `json:"type"`; Text string `json:"text"` }

// openAITool matches Responses API tool definition (flattened function schema)
type openAITool struct {
	Type        string         `json:"type"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters"`
}
// oaToolCall matches streamed tool call events (flattened)
type oaToolCall struct { ID, Type, Name, Arguments string; Index *int }

// partial represents an in-progress tool call assembly during streaming.
type partial struct { ToolCall; index int }

func buildOATools(tools []ToolSpec) []openAITool {
	oa := make([]openAITool, len(tools))
	for i, t := range tools {
		oa[i].Type = "function"
		oa[i].Name = t.Name
		oa[i].Description = t.Description
		if len(t.Parameters) > 0 { oa[i].Parameters = t.Parameters } else { oa[i].Parameters = map[string]any{"type":"object","properties":map[string]any{}} }
	}
	return oa
}
func buildOAInput(msgs []ChatMessage) []oaInputItem {
	out := make([]oaInputItem, len(msgs))
	for i, m := range msgs { out[i].Role = m.Role; out[i].Content = []oaContentPart{{Type:"input_text", Text:m.Content}} }
	return out
}

func (o *OpenAI) buildRequest(ctx context.Context, msgs []ChatMessage, tools []ToolSpec, stream bool) (*http.Request, error) {
	if o.key == "" { return nil, errors.New("missing api key") }
	body := map[string]any{"model": o.model, "input": buildOAInput(msgs)}
	if len(tools) > 0 { body["tools"] = buildOATools(tools) }
	if stream { body["stream"] = true }
	if o.Temperature != nil && supportsTemperature(o.model) { body["temperature"] = *o.Temperature }
	b, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/responses", bytes.NewReader(b))
	if err != nil { return nil, err }
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.key)
	return req, nil
}

func (o *OpenAI) Complete(ctx context.Context, msgs []ChatMessage, tools []ToolSpec) (Completion, error) {
	req, err := o.buildRequest(ctx, msgs, tools, false)
	if err != nil { return Completion{}, err }
	resp, err := o.client.Do(req)
	if err != nil { return Completion{}, err }
	defer resp.Body.Close()
	if resp.StatusCode >= 300 { body, _ := io.ReadAll(resp.Body); return Completion{}, errors.New(string(body)) }
	// Read full body for flexible parsing
	data, err := io.ReadAll(resp.Body); if err != nil { return Completion{}, err }
	// Base struct (best effort)
	var base struct {
		Output []struct { Type, Text string } `json:"output"`
		Usage struct { InputTokens int `json:"input_tokens"`; OutputTokens int `json:"output_tokens"` } `json:"usage"`
		ToolCalls []oaToolCall `json:"tool_calls"`
	}
	_ = json.Unmarshal(data, &base)
	// Collect text
	var texts []string
	for _, o := range base.Output { if strings.Contains(o.Type, "output_text") && o.Text != "" { texts = append(texts, o.Text) } }
	if len(texts) == 0 {
		// Generic scan for any objects with type=output_text
		var generic any
		if err := json.Unmarshal(data, &generic); err == nil {
			collectOutputText(generic, &texts)
		}
	}
	var sb strings.Builder; for _, t := range texts { sb.WriteString(t) }
	comp := Completion{Content: sb.String(), InputTokens: base.Usage.InputTokens, OutputTokens: base.Usage.OutputTokens, ModelName: "openai/"+o.model}
	for _, tc := range base.ToolCalls { // support either flattened or nested
		name := tc.Name
		args := tc.Arguments
		if name == "" && args == "" { // attempt nested fallback
			// ignore for now; legacy path removed
		}
		comp.ToolCalls = append(comp.ToolCalls, ToolCall{ID: tc.ID, Name: name, Arguments: []byte(args)})
	}
	return comp, nil
}

func (o *OpenAI) Stream(ctx context.Context, msgs []ChatMessage, tools []ToolSpec) (<-chan StreamChunk, error) {
	req, err := o.buildRequest(ctx, msgs, tools, true); if err != nil { return nil, err }
	out := make(chan StreamChunk, 32)
	go func() {
		defer close(out)
		resp, err := o.client.Do(req); if err != nil { out <- StreamChunk{Err: err}; return }
		defer resp.Body.Close()
		if resp.StatusCode >= 300 { body, _ := io.ReadAll(resp.Body); out <- StreamChunk{Err: errors.New(string(body))}; return }
		scanner := bufio.NewScanner(resp.Body); scanner.Buffer(make([]byte,0,64*1024), 1024*1024)
		partials := map[int]*partial{}
		var inTok, outTok int
		for scanner.Scan() {
			if ctx.Err() != nil { out <- StreamChunk{Err: ctx.Err()}; return }
			line := scanner.Text(); if !strings.HasPrefix(line, "data:") { continue }
			payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if payload == "[DONE]" { finalizeOpenAI(partials, out, inTok, outTok); return }
			if payload == "" { continue }
			var env map[string]any; if err := json.Unmarshal([]byte(payload), &env); err != nil { continue }
			t, _ := env["type"].(string)
			switch {
			case strings.HasSuffix(t, ".delta") && strings.Contains(t, "output_text"):
				if d, ok := env["delta"].(string); ok && d != "" { out <- StreamChunk{ContentDelta: d} }
			case strings.HasSuffix(t, ".delta") && strings.Contains(t, "tool_calls"):
				if arr, ok := env["tool_calls"].([]any); ok { for _, v := range arr { if m, ok := v.(map[string]any); ok {
					idx := 0; if iv, ok := m["index"].(float64); ok { idx = int(iv) }
					p := partials[idx]; if p == nil { p = &partial{index: idx}; partials[idx] = p }
					if id, _ := m["id"].(string); id != "" { p.ID = id }
					// flattened fields
					if name, _ := m["name"].(string); name != "" { p.Name = name }
					if args, _ := m["arguments"].(string); args != "" { p.Arguments = append(p.Arguments, []byte(args)...) }
					// nested legacy fallback
					if fn, ok := m["function"].(map[string]any); ok {
						if name, _ := fn["name"].(string); name != "" { p.Name = name }
						if args, _ := fn["arguments"].(string); args != "" { p.Arguments = append(p.Arguments, []byte(args)...) }
					}
				} } }
			case t == "response.completed":
				if u, ok := env["usage"].(map[string]any); ok { if iv, ok := u["input_tokens"].(float64); ok { inTok = int(iv) }; if ov, ok := u["output_tokens"].(float64); ok { outTok = int(ov) } }
				finalizeOpenAI(partials, out, inTok, outTok); return
			default: /* ignore */
			}
		}
		if err := scanner.Err(); err != nil { out <- StreamChunk{Err: err} } else { finalizeOpenAI(partials, out, inTok, outTok) }
	}()
	return out, nil
}

func finalizeOpenAI(partials map[int]*partial, out chan<- StreamChunk, inTok, outTok int) {
	if len(partials) == 0 { out <- StreamChunk{Done: true, InputTokens: inTok, OutputTokens: outTok}; return }
	idxs := make([]int,0,len(partials)); for i := range partials { idxs = append(idxs, i) }; sort.Ints(idxs)
	final := make([]ToolCall,0,len(partials)); for _, i := range idxs { p := partials[i]; final = append(final, ToolCall{ID:p.ID, Name:p.Name, Arguments:p.Arguments}) }
	out <- StreamChunk{Done: true, ToolCalls: final, InputTokens: inTok, OutputTokens: outTok}
}

func (o *OpenAI) ModelName() string { return o.model }
func supportsTemperature(model string) bool { m := strings.ToLower(model); if strings.HasPrefix(m, "gpt-5") || strings.HasPrefix(m, "o1") || strings.HasPrefix(m, "o3") || strings.HasPrefix(m, "o4") { return false }; return true }

// collectOutputText walks arbitrary decoded JSON and extracts any objects
// with type == "output_text" and a non-empty text field.
func collectOutputText(node any, out *[]string) {
	switch v := node.(type) {
	case map[string]any:
		if t, ok := v["type"].(string); ok && strings.Contains(t, "output_text") {
			if txt, _ := v["text"].(string); txt != "" { *out = append(*out, txt) }
		}
		for _, vv := range v { collectOutputText(vv, out) }
	case []any:
		for _, itm := range v { collectOutputText(itm, out) }
	}
}
