package tests

import (
	"context"
	"testing"

	"github.com/marcodenic/agentry/internal/tool"
	"github.com/marcodenic/agentry/pkg/flow"
)

func TestFlowEngineSequential(t *testing.T) {
	f := &flow.File{
		Agents: map[string]flow.Agent{
			"tester": {Model: "mock"},
		},
		Tasks: []flow.Task{
			{Sequential: []flow.Task{
				{Agent: "tester", Input: "one"},
				{Agent: "tester", Input: "two"},
			}},
		},
	}
	outs, err := flow.Run(context.Background(), f, tool.DefaultRegistry(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(outs) != 2 {
		t.Fatalf("expected 2 outputs, got %d", len(outs))
	}
	for _, o := range outs {
		if o != "hello" {
			t.Fatalf("unexpected output %s", o)
		}
	}
}

func TestFlowEngineParallel(t *testing.T) {
	f := &flow.File{
		Agents: map[string]flow.Agent{
			"tester": {Model: "mock"},
		},
		Tasks: []flow.Task{
			{Parallel: []flow.Task{
				{Agent: "tester", Input: "a"},
				{Agent: "tester", Input: "b"},
			}},
		},
	}
	outs, err := flow.Run(context.Background(), f, tool.DefaultRegistry(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(outs) != 2 {
		t.Fatalf("expected 2 outputs, got %d", len(outs))
	}
	for _, o := range outs {
		if o != "hello" {
			t.Fatalf("unexpected output %s", o)
		}
	}
}
