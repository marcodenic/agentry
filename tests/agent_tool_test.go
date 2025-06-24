package tests

import (
	"context"
	"testing"

	"github.com/marcodenic/agentry/internal/converse"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/internal/team"
	"github.com/marcodenic/agentry/internal/tool"
)

// staticClient returns a fixed response.
type staticClient struct{ out string }

func (s staticClient) Complete(ctx context.Context, msgs []model.ChatMessage, tools []model.ToolSpec) (model.Completion, error) {
	return model.Completion{Content: s.out}, nil
}

func newTestTeam(t *testing.T, reply string) *converse.Team {
	route := router.Rules{{Name: "mock", IfContains: []string{""}, Client: staticClient{out: reply}}}
	ag := core.New(route, tool.DefaultRegistry(), memory.NewInMemory(), nil, memory.NewInMemoryVector(), nil)
	tm, err := converse.NewTeam(ag, 2, "")
	if err != nil {
		t.Fatal(err)
	}
	return tm
}

func TestAgentToolDelegates(t *testing.T) {
	tm := newTestTeam(t, "ok")
	ctx := team.WithContext(context.Background(), tm)
	tl, ok := tool.DefaultRegistry().Use("agent")
	if !ok {
		t.Fatal("agent tool missing")
	}
	out, err := tl.Execute(ctx, map[string]any{"agent": "Agent1", "input": "hi"})
	if err != nil {
		t.Fatalf("exec failed: %v", err)
	}
	if out != "ok" {
		t.Fatalf("unexpected output %s", out)
	}
}

func TestAgentToolUnknown(t *testing.T) {
	tm := newTestTeam(t, "ignore")
	ctx := team.WithContext(context.Background(), tm)
	tl, _ := tool.DefaultRegistry().Use("agent")
	_, err := tl.Execute(ctx, map[string]any{"agent": "Bogus", "input": "hi"})
	if err == nil {
		t.Fatal("expected unknown agent error")
	}
	expectedErr := "unknown agent: Bogus"
	if err.Error() != expectedErr {
		t.Fatalf("expected '%s' error, got %v", expectedErr, err)
	}
}
