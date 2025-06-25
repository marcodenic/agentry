package tests

import (
	"testing"

	"github.com/marcodenic/agentry/internal/converse"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/internal/tool"
)

func newTestAgent() *core.Agent {
	route := router.Rules{{Name: "mock", IfContains: []string{""}, Client: staticClient{out: "test response"}}}
	return core.New(route, tool.DefaultRegistry(), memory.NewInMemory(), nil, memory.NewInMemoryVector(), nil)
}

// TestNewTeamContextEmptyStart verifies that NewTeamContext starts with no agents
// while NewTeam pre-spawns agents as before.
func TestNewTeamContextEmptyStart(t *testing.T) {
	parent := newTestAgent()

	// Test NewTeamContext - should start empty
	teamContext, err := converse.NewTeamContext(parent)
	if err != nil {
		t.Fatalf("NewTeamContext failed: %v", err)
	}

	if len(teamContext.Agents()) != 0 {
		t.Errorf("NewTeamContext should start with 0 agents, got %d", len(teamContext.Agents()))
	}

	if len(teamContext.Names()) != 0 {
		t.Errorf("NewTeamContext should start with 0 names, got %d", len(teamContext.Names()))
	}

	// Test NewTeam - should still pre-spawn agents
	teamWithAgents, err := converse.NewTeam(parent, 2, "test topic")
	if err != nil {
		t.Fatalf("NewTeam failed: %v", err)
	}

	if len(teamWithAgents.Agents()) != 2 {
		t.Errorf("NewTeam should pre-spawn 2 agents, got %d", len(teamWithAgents.Agents()))
	}

	if len(teamWithAgents.Names()) != 2 {
		t.Errorf("NewTeam should have 2 names, got %d", len(teamWithAgents.Names()))
	}
}

// TestTeamContextDynamicAgentAddition verifies that agents can be added
// dynamically to a team created with NewTeamContext.
func TestTeamContextDynamicAgentAddition(t *testing.T) {
	parent := newTestAgent()

	team, err := converse.NewTeamContext(parent)
	if err != nil {
		t.Fatalf("NewTeamContext failed: %v", err)
	}

	// Add first agent
	agent1, name1 := team.AddAgent("TestAgent1")
	if agent1 == nil {
		t.Error("AddAgent should return a valid agent")
	}
	if name1 != "TestAgent1" {
		t.Errorf("Expected agent name 'TestAgent1', got '%s'", name1)
	}
	if len(team.Agents()) != 1 {
		t.Errorf("Team should have 1 agent after adding one, got %d", len(team.Agents()))
	}

	// Add second agent
	agent2, name2 := team.AddAgent("TestAgent2")
	if agent2 == nil {
		t.Error("AddAgent should return a valid agent")
	}
	if name2 != "TestAgent2" {
		t.Errorf("Expected agent name 'TestAgent2', got '%s'", name2)
	}
	if len(team.Agents()) != 2 {
		t.Errorf("Team should have 2 agents after adding two, got %d", len(team.Agents()))
	}

	// Verify names list
	names := team.Names()
	if len(names) != 2 {
		t.Errorf("Team should have 2 names, got %d", len(names))
	}
	if names[0] != "TestAgent1" || names[1] != "TestAgent2" {
		t.Errorf("Unexpected names order: %v", names)
	}
}
