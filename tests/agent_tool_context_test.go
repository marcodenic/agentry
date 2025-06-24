package tests

import (
	"context"
	"testing"

	"github.com/marcodenic/agentry/internal/converse"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/internal/teamctx"
	"github.com/marcodenic/agentry/internal/tool"
)

// simpleMock returns a simple text completion.
type agentMock struct{}

func (agentMock) Complete(ctx context.Context, msgs []model.ChatMessage, tools []model.ToolSpec) (model.Completion, error) {
	return model.Completion{Content: "Hello from Agent1"}, nil
}

func TestAgentToolContext(t *testing.T) {
	route := router.Rules{{Name: "mock", IfContains: []string{""}, Client: agentMock{}}}
	ag := core.New(route, tool.DefaultRegistry(), memory.NewInMemory(), nil, memory.NewInMemoryVector(), nil)
	team, err := converse.NewTeam(ag, 1, "hi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ctx := context.WithValue(context.Background(), teamctx.Key{}, team)
	tl, ok := tool.DefaultRegistry()["agent"]
	if !ok {
		t.Fatalf("agent tool missing")
	}
	out, err := tl.Execute(ctx, map[string]any{"agent": "Agent1", "input": "foo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == "" {
		t.Fatalf("expected output, got empty")
	}
}
