package tests

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/converse"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/internal/team"
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

// TestTUIAgentToolIntegration ensures the agent tool works within the unified TUI model.
func TestTUIAgentToolIntegration(t *testing.T) {
	client := &agentCallClient{}
	route := router.Rules{{Name: "mock", IfContains: []string{""}, Client: client}}
	ag := core.New(route, tool.DefaultRegistry(), memory.NewInMemory(), nil, memory.NewInMemoryVector(), nil)	// Use the unified Model instead of deprecated TeamModel
	_ = tui.New(ag)
	// Create team context for agent tool to work properly
	tm, err := converse.NewTeam(ag, 1, "")
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}
	
	// Simulate user input to trigger agent call with proper team context
	ctx := team.WithContext(context.Background(), tm)
	_, err = ag.Run(ctx, "ping")
	if err != nil {
		t.Fatalf("agent run failed: %v", err)
	}

	// Check that the agent tool was called
	if client.call < 2 {
		t.Fatalf("agent tool did not execute, calls: %d", client.call)
	}
}
