package tests

import (
	"context"
	"testing"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/tool"
)

type recordingClient struct {
	calls [][]model.ChatMessage
}

func (r *recordingClient) Stream(ctx context.Context, msgs []model.ChatMessage, tools []model.ToolSpec) (<-chan model.StreamChunk, error) {
	cp := make([]model.ChatMessage, len(msgs))
	copy(cp, msgs)
	r.calls = append(r.calls, cp)
	ch := make(chan model.StreamChunk, 1)
	go func() {
		defer close(ch)
		ch <- model.StreamChunk{ContentDelta: "ack", Done: true}
	}()
	return ch, nil
}

func TestAgentHistoryIncludesPreviousTurn(t *testing.T) {
	client := &recordingClient{}
	memStore := memory.NewInMemory()
	ag := core.New(client, "mock", tool.Registry{}, memStore, memory.NewInMemoryVector(), nil)

	ctx := context.Background()

	if _, err := ag.Run(ctx, "my fave color is purple"); err != nil {
		t.Fatalf("first run failed: %v", err)
	}

	if len(memStore.History()) != 1 {
		t.Fatalf("expected 1 history step, got %d", len(memStore.History()))
	}

	if _, err := ag.Run(ctx, "what's my fave color?"); err != nil {
		t.Fatalf("second run failed: %v", err)
	}

	if len(client.calls) < 2 {
		t.Fatalf("expected at least two calls, got %d", len(client.calls))
	}

	lastMsgs := client.calls[len(client.calls)-1]
	found := false
	for _, m := range lastMsgs {
		if m.Role == "user" && m.Content == "my fave color is purple" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("previous user turn not included in second call: %#v", lastMsgs)
	}
}
