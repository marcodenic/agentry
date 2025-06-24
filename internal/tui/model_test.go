package tui

import (
	"testing"

	"github.com/google/uuid"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/internal/tool"
)

func TestNew(t *testing.T) {
	ag := core.New(router.Rules{{IfContains: []string{""}, Client: nil}}, tool.Registry{}, memory.NewInMemory(), nil, memory.NewInMemoryVector(), nil)
	m := New(ag)
	if m.masterAgent != ag {
		t.Fatalf("agent mismatch")
	}
	if len(m.agents) != 1 {
		t.Fatalf("expected one agent")
	}
}

func TestCommandFlow(t *testing.T) {
	ag := core.New(router.Rules{{IfContains: []string{""}, Client: nil}}, tool.Registry{}, memory.NewInMemory(), nil, memory.NewInMemoryVector(), nil)
	m := New(ag)

	m, _ = m.handleCommand("/spawn helper")
	if len(m.agents) != 2 {
		t.Fatalf("spawn failed")
	}

	var newID uuid.UUID
	for id := range m.agents {
		if id != ag.ID {
			newID = id
		}
	}

	m, _ = m.handleCommand("/switch " + newID.String()[:8])
	if m.active != newID {
		t.Fatalf("switch failed")
	}

	m, _ = m.handleCommand("/stop " + newID.String()[:8])
	if m.agents[newID].Status != StatusStopped {
		t.Fatalf("stop failed")
	}
}
