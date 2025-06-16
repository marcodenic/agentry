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

type seqMock struct{ n int }

func (m *seqMock) Complete(ctx context.Context, msgs []model.ChatMessage, tools []model.ToolSpec) (model.Completion, error) {
	m.n++
	return model.Completion{Content: fmt.Sprintf("msg%d", m.n)}, nil
}

func TestConverseRunner(t *testing.T) {
	mock := &seqMock{}
	route := router.Rules{{Name: "mock", IfContains: []string{""}, Client: mock}}
	parent := core.New(route, nil, memory.NewInMemory(), nil)

	out, err := converse.Run(context.Background(), parent, 3, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 10 {
		t.Fatalf("expected 10 turns, got %d", len(out))
	}
	for i, msg := range out {
		exp := fmt.Sprintf("msg%d", i+1)
		if msg != exp {
			t.Fatalf("turn %d want %s got %s", i, exp, msg)
		}
	}
}
