package tests

import (
	"context"
	"testing"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/tool"
)

func TestAgentBuiltin(t *testing.T) {
	reg := tool.DefaultRegistry()
	client := model.NewMock()
	parent := core.New(client, "mock", reg, memory.NewInMemory(), memory.NewInMemoryVector(), nil)

	agentTool, _ := reg.Use("agent")
	spawn := func(ctx context.Context, q string) (string, error) {
		sub := parent.Spawn()
		return sub.Run(ctx, q)
	}
	out, err := agentTool.Execute(tool.WithSpawn(context.Background(), spawn), map[string]any{"query": "hi"})
	if err != nil {
		t.Fatal(err)
	}
	if out != "hello" {
		t.Fatalf("expected hello, got %s", out)
	}
}
