package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/router"
)

// simpleClient is a minimal model.Client for testing.
type simpleClient struct {
	out string
	err error
}

func (s *simpleClient) Complete(ctx context.Context, msgs []model.ChatMessage, tools []model.ToolSpec) (model.Completion, error) {
	if s.err != nil {
		return model.Completion{}, s.err
	}
	return model.Completion{Content: s.out}, nil
}

func newAgent(out string, err error) *core.Agent {
	c := &simpleClient{out: out, err: err}
	route := router.Rules{{Name: "mock", IfContains: []string{""}, Client: c}}
	return core.New(route, nil, memory.NewInMemory(), nil)
}

func TestRunParallelAggregatesErrors(t *testing.T) {
	ctx := context.Background()
	errBoom := errors.New("boom")
	ag1 := newAgent("ok", nil)
	ag2 := newAgent("", errBoom)

	out, err := core.RunParallel(ctx, []*core.Agent{ag1, ag2}, []string{"a", "b"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, errBoom) {
		t.Fatalf("expected aggregated error to contain original; got %v", err)
	}
	if out[0] != "ok" || out[1] != "" {
		t.Fatalf("unexpected outputs: %#v", out)
	}
}

func TestRunParallelMultipleErrors(t *testing.T) {
	ctx := context.Background()
	err1 := errors.New("fail1")
	err2 := errors.New("fail2")
	ag1 := newAgent("", err1)
	ag2 := newAgent("", err2)

	_, err := core.RunParallel(ctx, []*core.Agent{ag1, ag2}, []string{"x", "y"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, err1) || !errors.Is(err, err2) {
		t.Fatalf("expected both errors aggregated, got: %v", err)
	}
}
