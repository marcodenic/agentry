package tests

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/tool"
	"github.com/marcodenic/agentry/internal/trace"
)

type loopMock struct{}

func (loopMock) Complete(ctx context.Context, msgs []model.ChatMessage, tools []model.ToolSpec) (model.Completion, error) {
	// After a tool result is present, return a final message to end the run.
	if len(msgs) > 0 && msgs[len(msgs)-1].Role == "tool" {
		return model.Completion{Content: "done"}, nil
	}
	args, _ := json.Marshal(map[string]string{"text": "hi"})
	return model.Completion{ToolCalls: []model.ToolCall{{ID: "1", Name: "echo", Arguments: args}}}, nil
}

type captureWriter struct{ events []trace.Event }

func (c *captureWriter) Write(_ context.Context, e trace.Event) { c.events = append(c.events, e) }

func TestAgentRunCompletesWithoutCap(t *testing.T) {
	cw := &captureWriter{}
	ag := core.New(loopMock{}, "mock", tool.DefaultRegistry(), memory.NewInMemory(), memory.NewInMemoryVector(), cw)

	out, err := ag.Run(context.Background(), "start")
	if err != nil {
		t.Fatal(err)
	}
	if out != "done" {
		t.Fatalf("expected final output 'done', got %s", out)
	}
	if len(cw.events) == 0 {
		t.Fatal("no events captured")
	}
	if cw.events[len(cw.events)-1].Type != trace.EventFinal {
		t.Fatalf("expected final event, got %v", cw.events[len(cw.events)-1].Type)
	}
}
