package core

import (
	"context"
	"errors"
	"testing"

	"github.com/marcodenic/agentry/internal/model"
)

func TestChunkAggregatorCollectsFinalValues(t *testing.T) {
	ag := &Agent{ModelName: "openai/gpt-4"}
	s := &conversationSession{agent: ag, ctx: context.Background(), tracer: noopConversationTracer{}}

	aggregator := newChunkAggregator(s)
	stream := make(chan model.StreamChunk, 2)
	stream <- model.StreamChunk{ContentDelta: "hello "}
	stream <- model.StreamChunk{
		ContentDelta: "world",
		Done:         true,
		InputTokens:  5,
		OutputTokens: 3,
		ToolCalls:    []model.ToolCall{{ID: "1", Name: "echo"}},
		ModelName:    "custom-model",
		ResponseID:   "resp-123",
	}
	close(stream)

	if err := aggregator.Collect(stream); err != nil {
		t.Fatalf("Collect returned error: %v", err)
	}
	completion, responseID := aggregator.Result()

	if completion.Content != "hello world" {
		t.Fatalf("unexpected content: %q", completion.Content)
	}
	if len(completion.ToolCalls) != 1 || completion.ToolCalls[0].ID != "1" {
		t.Fatalf("unexpected tool calls: %#v", completion.ToolCalls)
	}
	if completion.InputTokens != 5 || completion.OutputTokens != 3 {
		t.Fatalf("unexpected token accounting: %#v", completion)
	}
	if completion.ModelName != "custom-model" {
		t.Fatalf("expected model override, got %q", completion.ModelName)
	}
	if responseID != "resp-123" {
		t.Fatalf("unexpected response ID: %q", responseID)
	}
}

func TestChunkAggregatorPropagatesChunkError(t *testing.T) {
	ag := &Agent{}
	s := &conversationSession{agent: ag, ctx: context.Background(), tracer: noopConversationTracer{}}

	aggregator := newChunkAggregator(s)
	boom := errors.New("boom")
	stream := make(chan model.StreamChunk, 1)
	stream <- model.StreamChunk{Err: boom}
	close(stream)

	err := aggregator.Collect(stream)
	if !errors.Is(err, boom) {
		t.Fatalf("expected original error, got %v", err)
	}
}
