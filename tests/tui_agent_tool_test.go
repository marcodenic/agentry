package tests

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/internal/tool"
	"github.com/marcodenic/agentry/internal/tui"
)

// agentCallClient triggers the agent tool on the first call then returns a final message.
type agentCallClient struct{ call int }

func (c *agentCallClient) Complete(ctx context.Context, msgs []model.ChatMessage, tools []model.ToolSpec) (model.Completion, error) {
	c.call++
	if c.call == 1 {
		args, _ := json.Marshal(map[string]string{"agent": "Agent1", "input": "ping"})
		return model.Completion{ToolCalls: []model.ToolCall{{ID: "1", Name: "agent", Arguments: args}}}, nil
	}
	return model.Completion{Content: "done"}, nil
}

// TestTUIAgentToolIntegration ensures the agent tool works within the TUI team model.
func TestTUIAgentToolIntegration(t *testing.T) {
	client := &agentCallClient{}
	route := router.Rules{{Name: "mock", IfContains: []string{""}, Client: client}}
	ag := core.New(route, tool.DefaultRegistry(), memory.NewInMemory(), nil, memory.NewInMemoryVector(), nil)

	tm, err := tui.NewTeam(ag, 1, "hi")
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	// Run one update cycle manually.
	cmd := tm.Init()
	if cmd == nil {
		t.Fatal("init returned nil cmd")
	}
	msg := cmd()
	m, _ := tm.Update(msg)
	tm = m.(tui.TeamModel)

	// Check internal error field via reflection.
	errVal := reflect.ValueOf(tm).FieldByName("err")
	if errVal.IsValid() && !errVal.IsNil() {
		t.Fatalf("unexpected error: %v", errVal.Interface())
	}

	if client.call < 2 {
		t.Fatalf("agent tool did not execute, calls: %d", client.call)
	}
}
