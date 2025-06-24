package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/marcodenic/agentry/internal/converse"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/router"
)

type seqTeamMock struct{ n int }

func (m *seqTeamMock) Complete(ctx context.Context, msgs []model.ChatMessage, tools []model.ToolSpec) (model.Completion, error) {
	m.n++
	return model.Completion{Content: fmt.Sprintf("msg%d", m.n)}, nil
}

func TestTeamSpawnConversation(t *testing.T) {
	mock := &seqTeamMock{}
	route := router.Rules{{Name: "mock", IfContains: []string{""}, Client: mock}}
	parent := core.New(route, nil, memory.NewInMemory(), nil, memory.NewInMemoryVector(), nil)
	tm, err := converse.NewTeam(parent, 2, "hi")
	if err != nil {
		t.Fatalf("new team: %v", err)
	}
	ctx := context.Background()
	idx, out, err := tm.Step(ctx)
	if err != nil {
		t.Fatalf("step1 error: %v", err)
	}
	if idx != 0 || out != "msg1" {
		t.Fatalf("turn1 expected agent0 msg1 got idx %d %s", idx, out)
	}
	idx, out, err = tm.Step(ctx)
	if err != nil {
		t.Fatalf("step2 error: %v", err)
	}
	if idx != 1 || out != "msg2" {
		t.Fatalf("turn2 expected agent1 msg2 got idx %d %s", idx, out)
	}
}
