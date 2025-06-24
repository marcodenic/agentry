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
	m, err := NewChat(ag, 1, "")
	if err != nil {
		t.Fatalf("new chat error: %v", err)
	}
	if len(m.infos) != 1 {
		t.Fatalf("expected one agent")
	}
}

func TestCommandFlow(t *testing.T) {
	ag := core.New(router.Rules{{IfContains: []string{""}, Client: nil}}, tool.Registry{}, memory.NewInMemory(), nil, memory.NewInMemoryVector(), nil)
	m, err := NewChat(ag, 1, "")
	if err != nil {
		t.Fatalf("new chat error: %v", err)
	}

	m, _ = m.handleCommand("/spawn helper")
	if len(m.infos) != 2 {
		t.Fatalf("spawn failed")
	}

	var newID uuid.UUID
	for id := range m.infos {
		if id != m.active {
			newID = id
		}
	}

	m, _ = m.handleCommand("/switch " + newID.String()[:8])
	if m.active != newID {
		t.Fatalf("switch failed")
	}

	m, _ = m.handleCommand("/stop " + newID.String()[:8])
	if m.infos[newID].Status != StatusStopped {
		t.Fatalf("stop failed")
	}
}
