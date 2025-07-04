package tui

import (
	"context"
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/google/uuid"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/tool"
)

func TestNew(t *testing.T) {
	ag := core.New(model.NewMock(), "mock", tool.Registry{}, memory.NewInMemory(), nil, memory.NewInMemoryVector(), nil)
	m := New(ag)
	if len(m.infos) != 1 {
		t.Fatalf("expected one agent")
	}
}

func TestCommandFlow(t *testing.T) {
	ag := core.New(model.NewMock(), "mock", tool.Registry{}, memory.NewInMemory(), nil, memory.NewInMemoryVector(), nil)
	m := New(ag)

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

func TestModelBasicInteraction(t *testing.T) {
	mock := &seqMock{}
	ag := core.New(mock, "mock", tool.Registry{}, memory.NewInMemory(), nil, memory.NewInMemoryVector(), nil)

	m := New(ag)

	// Test that we can send a message to the agent
	m.input.SetValue("test message")
	nm, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = nm.(Model)

	// Check that the message appears in the active agent's history
	activeInfo := m.infos[m.active]
	if !strings.Contains(activeInfo.History, "test message") {
		t.Fatalf("agent history should contain user input: %s", activeInfo.History)
	}
}

type seqMock struct{ n int }

func (m *seqMock) Complete(ctx context.Context, msgs []model.ChatMessage, tools []model.ToolSpec) (model.Completion, error) {
	m.n++
	return model.Completion{Content: fmt.Sprintf("msg%d", m.n)}, nil
}

func TestModelMultipleAgents(t *testing.T) {
	mock := &seqMock{}
	ag := core.New(mock, "mock", tool.Registry{}, memory.NewInMemory(), nil, memory.NewInMemoryVector(), nil)

	m := New(ag)

	m, _ = m.handleCommand("/spawn second")
	m, _ = m.handleCommand("/spawn third")
	if len(m.infos) != 3 {
		t.Fatalf("expected 3 agents got %d", len(m.infos))
	}

	var secondID uuid.UUID
	for id, info := range m.infos {
		if info.Name == "second" {
			secondID = id
			break
		}
	}
	if secondID == uuid.Nil {
		t.Fatal("second agent not found")
	}

	m, _ = m.handleCommand("/switch " + secondID.String()[:8])
	if m.active != secondID {
		t.Fatalf("switch failed")
	} // simulate window sizing so viewport renders
	nm, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 20})
	m = nm.(Model)

	// Simulate entering text and pressing enter
	m.input.SetValue("ping")
	nm, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = nm.(Model)

	hist := m.infos[secondID].History
	if !strings.Contains(hist, "ping") {
		t.Fatalf("history missing input: %s", hist)
	}

	// Check that the agent is now running (processing the request)
	if m.infos[secondID].Status != StatusRunning {
		t.Fatalf("agent should be running after receiving input, got status: %v", m.infos[secondID].Status)
	}
}
