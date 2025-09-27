package model

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/marcodenic/agentry/internal/debug"
)

type oaFinalizeMode int

const (
	finalizeModeLegacy oaFinalizeMode = iota
	finalizeModeResponses
)

type oaStreamResult struct {
	partials      map[int]*partial
	responseCalls map[string]*partial
	inputTokens   int
	outputTokens  int
	responseID    string
	mode          oaFinalizeMode
}

type openAIStreamReader struct {
	model     string
	msgCount  int
	startTime time.Time
}

func newOpenAIStreamReader(model string, msgCount int, start time.Time) *openAIStreamReader {
	return &openAIStreamReader{model: model, msgCount: msgCount, startTime: start}
}

func (r *openAIStreamReader) Read(ctx context.Context, src io.Reader, emit func(StreamChunk)) (oaStreamResult, error) {
	scanner := bufio.NewScanner(src)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	partials := make(map[int]*partial)
	responseCalls := make(map[string]*partial)
	var inTok, outTok int
	var responseID string
	scanStartTime := time.Now()
	lineCount := 0
	debug.Printf("OpenAI.Stream: starting stream scan, elapsed=%v", time.Since(r.startTime))

	for scanner.Scan() {
		lineCount++
		if lineCount%100 == 0 {
			debug.Printf("OpenAI.Stream: processed %d lines, elapsed=%v", lineCount, time.Since(r.startTime))
		}

		if ctx.Err() != nil {
			debug.Printf("OpenAI.Stream: context error during scan: %v, elapsed=%v", ctx.Err(), time.Since(r.startTime))
			debug.LogEvent("MODEL", "context_error", map[string]interface{}{
				"provider":        "openai",
				"model":           r.model,
				"error":           ctx.Err().Error(),
				"lines_processed": lineCount,
			})
			emit(StreamChunk{Err: ctx.Err()})
			return oaStreamResult{}, ctx.Err()
		}
		line := scanner.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "[DONE]" {
			debug.Printf("OpenAI.Stream: [DONE], finalize (partials=%d responseCalls=%d) scan_duration=%v total_elapsed=%v lines=%d", len(partials), len(responseCalls), time.Since(scanStartTime), time.Since(r.startTime), lineCount)
			debug.LogModelInteraction("openai", r.model, r.msgCount, map[string]int{
				"input_tokens":  inTok,
				"output_tokens": outTok,
			}, time.Since(r.startTime))
			return oaStreamResult{
				partials:      partials,
				responseCalls: responseCalls,
				inputTokens:   inTok,
				outputTokens:  outTok,
				responseID:    responseID,
				mode:          finalizeModeLegacy,
			}, nil
		}
		if payload == "" {
			continue
		}
		var env map[string]any
		if err := json.Unmarshal([]byte(payload), &env); err != nil {
			continue
		}
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
				emit(StreamChunk{ContentDelta: d})
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
						if name, _ := m["name"].(string); name != "" {
							p.Name = name
						}
						if args, _ := m["arguments"].(string); args != "" {
							p.Arguments = append(p.Arguments, []byte(args)...)
						}
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
								"model":     r.model,
								"item_id":   itemID,
								"call_id":   callID,
								"tool_name": name,
							})
						}
					}
				}
			}
		case t == "response.function_call_arguments.delta":
			if itemID, _ := env["item_id"].(string); itemID != "" {
				if p, exists := responseCalls[itemID]; exists {
					if delta, _ := env["delta"].(string); delta != "" {
						p.Arguments = append(p.Arguments, []byte(delta)...)
					}
				}
			}
		case t == "response.function_call_arguments.done":
			if itemID, _ := env["item_id"].(string); itemID != "" {
				if p, exists := responseCalls[itemID]; exists {
					if args, _ := env["arguments"].(string); args != "" {
						p.Arguments = []byte(args)
					}
					debug.Printf("OpenAI.Stream: Function call args done for %s, args length: %d", itemID, len(p.Arguments))
					debug.LogEvent("MODEL", "tool_call_complete", map[string]interface{}{
						"provider":    "openai",
						"model":       r.model,
						"item_id":     itemID,
						"tool_name":   p.Name,
						"args_length": len(p.Arguments),
					})
				}
			}
		case t == "response.completed":
			debug.Printf("OpenAI.Stream: response completed, elapsed=%v", time.Since(r.startTime))
			if u, ok := env["usage"].(map[string]any); ok {
				if iv, ok := u["input_tokens"].(float64); ok {
					inTok = int(iv)
				}
				if ov, ok := u["output_tokens"].(float64); ok {
					outTok = int(ov)
				}
			}
			debug.LogModelInteraction("openai", r.model, r.msgCount, map[string]int{
				"input_tokens":  inTok,
				"output_tokens": outTok,
			}, time.Since(r.startTime))
			return oaStreamResult{
				partials:      partials,
				responseCalls: responseCalls,
				inputTokens:   inTok,
				outputTokens:  outTok,
				responseID:    responseID,
				mode:          finalizeModeResponses,
			}, nil
		default:
			// ignore other events
		}
	}
	scanEndTime := time.Now()
	if err := scanner.Err(); err != nil {
		debug.Printf("OpenAI.Stream: scanner error: %v scan_duration=%v total_elapsed=%v lines=%d", err, scanEndTime.Sub(scanStartTime), scanEndTime.Sub(r.startTime), lineCount)
		debug.LogEvent("MODEL", "scanner_error", map[string]interface{}{
			"provider":        "openai",
			"model":           r.model,
			"error":           err.Error(),
			"lines_processed": lineCount,
		})
		emit(StreamChunk{Err: err})
		return oaStreamResult{}, err
	}
	debug.Printf("OpenAI.Stream: scanner ended normally, finalize (partials=%d responseCalls=%d) scan_duration=%v total_elapsed=%v lines=%d", len(partials), len(responseCalls), scanEndTime.Sub(scanStartTime), scanEndTime.Sub(r.startTime), lineCount)
	debug.LogModelInteraction("openai", r.model, r.msgCount, map[string]int{
		"input_tokens":  inTok,
		"output_tokens": outTok,
	}, time.Since(r.startTime))
	return oaStreamResult{
		partials:      partials,
		responseCalls: responseCalls,
		inputTokens:   inTok,
		outputTokens:  outTok,
		responseID:    responseID,
		mode:          finalizeModeResponses,
	}, nil
}
