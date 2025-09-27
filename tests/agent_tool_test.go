package tests

import (
	"context"
	"testing"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/team"
	"github.com/marcodenic/agentry/internal/tool"
)

// staticClient returns a fixed response.
type staticClient struct{ out string }

func (s staticClient) Clone() model.Client { return staticClient{out: s.out} }

func (s staticClient) Stream(ctx context.Context, msgs []model.ChatMessage, tools []model.ToolSpec) (<-chan model.StreamChunk, error) {
	ch := make(chan model.StreamChunk, 1)
	ch <- model.StreamChunk{ContentDelta: s.out, Done: true}
	close(ch)
	return ch, nil
}

func newTestTeam(t *testing.T, reply string) (*team.Team, tool.Registry) {
	registry := tool.DefaultRegistry()
	ag := core.New(staticClient{out: reply}, "mock", registry, memory.NewInMemory(), memory.NewInMemoryVector(), nil)
	tm, err := team.NewTeam(ag, 2, "")
	if err != nil {
		t.Fatal(err)
	}
	// Register the team's agent tool to replace the placeholder
	t.Logf("Registering agent tool with team...")
	tm.RegisterAgentTool(registry)
	t.Logf("Agent tool registered")
	return tm, registry
}

func TestAgentToolDelegates(t *testing.T) {
	t.Logf("Creating test team...")
	tm, registry := newTestTeam(t, "ok")
	t.Logf("Setting up context...")
	ctx := team.WithContext(context.Background(), tm)
	t.Logf("Getting agent tool...")
	tl, ok := registry.Use("agent")
	if !ok {
		t.Fatal("agent tool missing")
	}
	t.Logf("Agent tool description: %s", tl.Description())
	t.Logf("Executing agent tool...")
	out, err := tl.Execute(ctx, map[string]any{"agent": "Agent1", "input": "hi"})
	t.Logf("Agent tool execution completed")
	if err != nil {
		t.Fatalf("exec failed: %v", err)
	}
	if out != "ok" {
		t.Fatalf("unexpected output %s", out)
	}
}

func TestAgentToolUnknown(t *testing.T) {
	tm, registry := newTestTeam(t, "ok")
	t.Logf("Initial agent count: %d", len(tm.GetTeamAgents()))
	ctx := team.WithContext(context.Background(), tm)
	tl, _ := registry.Use("agent")
	out, err := tl.Execute(ctx, map[string]any{"agent": "coder", "input": "hi"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "ok" {
		t.Fatalf("unexpected output %s", out)
	}
	t.Logf("Final agent count: %d", len(tm.GetTeamAgents()))
	if len(tm.GetTeamAgents()) != 1 {
		t.Fatalf("agent not spawned, got %d", len(tm.GetTeamAgents()))
	}
}
