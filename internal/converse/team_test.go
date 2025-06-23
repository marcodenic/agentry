package converse

import (
	"context"
	"fmt"
	"testing"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/router"
)

type seqMock struct{ n int }

func (m *seqMock) Complete(ctx context.Context, msgs []model.ChatMessage, tools []model.ToolSpec) (model.Completion, error) {
	m.n++
	return model.Completion{Content: fmt.Sprintf("msg%d", m.n)}, nil
}

func TestTeamCall(t *testing.T) {
	mock := &seqMock{}
	route := router.Rules{{Name: "mock", IfContains: []string{""}, Client: mock}}
	parent := core.New(route, nil, memory.NewInMemory(), nil, memory.NewInMemoryVector(), nil)

	tm, err := NewTeam(parent, 2, "hi")
	if err != nil {
		t.Fatalf("new team: %v", err)
	}

	out, err := tm.Call(context.Background(), "Agent1", "hello")
	if err != nil {
		t.Fatalf("call error: %v", err)
	}
	if out != "msg1" {
		t.Fatalf("expected msg1 got %s", out)
	}
}

func TestTeamCallUnknown(t *testing.T) {
	mock := &seqMock{}
	route := router.Rules{{Name: "mock", IfContains: []string{""}, Client: mock}}
	parent := core.New(route, nil, memory.NewInMemory(), nil, memory.NewInMemoryVector(), nil)

	tm, err := NewTeam(parent, 1, "hi")
	if err != nil {
		t.Fatalf("new team: %v", err)
	}
	if _, err := tm.Call(context.Background(), "Nope", "hi"); err == nil {
		t.Fatal("expected error for unknown agent")
	}
}
