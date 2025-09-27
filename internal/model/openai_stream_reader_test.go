package model

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestOpenAIStreamReaderLegacyCompletion(t *testing.T) {
	ctx := context.Background()
	reader := newOpenAIStreamReader("gpt-test", 2, time.Now())
	input := strings.Join([]string{
		"data: {\"type\":\"response.output_text.delta\",\"delta\":\"Hello\"}",
		"data: {\"type\":\"response.output_text.delta\",\"delta\":\" world\"}",
		"data: [DONE]",
	}, "\n\n") + "\n"

	ch := make(chan StreamChunk, 8)
	result, err := reader.Read(ctx, strings.NewReader(input), func(chunk StreamChunk) {
		ch <- chunk
	})
	if err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}
	if result.mode != finalizeModeLegacy {
		t.Fatalf("expected finalizeModeLegacy, got %v", result.mode)
	}
	finalizeOpenAI(result.partials, ch, result.inputTokens, result.outputTokens, "gpt-test", result.responseID)
	close(ch)

	var chunks []StreamChunk
	for c := range ch {
		chunks = append(chunks, c)
	}
	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks, got %d", len(chunks))
	}
	if chunks[0].ContentDelta != "Hello" {
		t.Fatalf("unexpected first delta: %q", chunks[0].ContentDelta)
	}
	if chunks[1].ContentDelta != " world" {
		t.Fatalf("unexpected second delta: %q", chunks[1].ContentDelta)
	}
	final := chunks[2]
	if !final.Done {
		t.Fatalf("expected final chunk to mark done")
	}
	if final.ResponseID != "" {
		t.Fatalf("expected blank response ID, got %s", final.ResponseID)
	}
	if final.InputTokens != 0 || final.OutputTokens != 0 {
		t.Fatalf("expected zero tokens, got %d/%d", final.InputTokens, final.OutputTokens)
	}
}

func TestOpenAIStreamReaderResponsesCompletion(t *testing.T) {
	ctx := context.Background()
	reader := newOpenAIStreamReader("gpt-test", 1, time.Now())
	input := strings.Join([]string{
		"data: {\"type\":\"response.output_item.added\",\"item\":{\"type\":\"function_call\",\"id\":\"item_1\",\"name\":\"call_tool\",\"call_id\":\"call_123\"}}",
		"data: {\"type\":\"response.function_call_arguments.done\",\"item_id\":\"item_1\",\"arguments\":\"{\\\"foo\\\":1,\\\"bar\\\":2}\"}",
		"data: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_456\"},\"usage\":{\"input_tokens\":11,\"output_tokens\":7}}",
	}, "\n\n") + "\n"

	ch := make(chan StreamChunk, 8)
	result, err := reader.Read(ctx, strings.NewReader(input), func(chunk StreamChunk) {
		ch <- chunk
	})
	if err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}
	if result.mode != finalizeModeResponses {
		t.Fatalf("expected finalizeModeResponses, got %v", result.mode)
	}
	finalizeWithResponses(result.partials, result.responseCalls, ch, result.inputTokens, result.outputTokens, "gpt-test", result.responseID)
	close(ch)

	var final StreamChunk
	for c := range ch {
		final = c
	}
	if !final.Done {
		t.Fatalf("expected final chunk, got %+v", final)
	}
	if final.ResponseID != "resp_456" {
		t.Fatalf("unexpected response ID: %s", final.ResponseID)
	}
	if final.InputTokens != 11 || final.OutputTokens != 7 {
		t.Fatalf("unexpected token counts: %d/%d", final.InputTokens, final.OutputTokens)
	}
	if len(final.ToolCalls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(final.ToolCalls))
	}
	call := final.ToolCalls[0]
	if call.ID != "call_123" || call.Name != "call_tool" {
		t.Fatalf("unexpected tool call: %+v", call)
	}
	if string(call.Arguments) != "{\"foo\":1,\"bar\":2}" {
		t.Fatalf("unexpected arguments: %s", string(call.Arguments))
	}
}
