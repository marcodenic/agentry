package tests

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/internal/tool"
	"github.com/marcodenic/agentry/internal/trace"
)

type loopMock struct{}

func (loopMock) Complete(ctx context.Context, msgs []model.ChatMessage, tools []model.ToolSpec) (model.Completion, error) {
	args, _ := json.Marshal(map[string]string{"text": "hi"})
	return model.Completion{ToolCalls: []model.ToolCall{{ID: "1", Name: "echo", Arguments: args}}}, nil
}

type captureWriter struct{ events []trace.Event }

func (c *captureWriter) Write(_ context.Context, e trace.Event) { c.events = append(c.events, e) }

func TestAgentRunYields(t *testing.T) {
	route := router.Rules{{Name: "mock", IfContains: []string{""}, Client: loopMock{}}}
	cw := &captureWriter{}
	ag := core.New(route, tool.DefaultRegistry(), memory.NewInMemory(), nil, memory.NewInMemoryVector(), cw)

	out, err := ag.Run(context.Background(), "start")
	if err != nil {
		t.Fatal(err)
	}
	if out != "" {
		t.Fatalf("expected empty output, got %s", out)
	}
	if len(cw.events) == 0 {
		t.Fatal("no events captured")
	}
	if cw.events[len(cw.events)-1].Type != trace.EventYield {
		t.Fatalf("expected yield event, got %v", cw.events[len(cw.events)-1].Type)
	}
}
