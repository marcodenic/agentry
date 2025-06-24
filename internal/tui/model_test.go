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

type seqMock struct{ n int }

func (m *seqMock) Complete(ctx context.Context, msgs []model.ChatMessage, tools []model.ToolSpec) (model.Completion, error) {
	m.n++
	return model.Completion{Content: fmt.Sprintf("msg%d", m.n)}, nil
}

func TestChatModelMultipleAgents(t *testing.T) {
	mock := &seqMock{}
	route := router.Rules{{Name: "mock", IfContains: []string{""}, Client: mock}}
	ag := core.New(route, tool.Registry{}, memory.NewInMemory(), nil, memory.NewInMemoryVector(), nil)

	m, err := NewChat(ag, 1, "")
	if err != nil {
		t.Fatalf("new chat error: %v", err)
	}

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
	}

	// simulate window sizing so viewport renders
	nm, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 20})
	m = nm.(ChatModel)

	m, _ = m.callActive("ping")
	hist := m.infos[secondID].History
	if !strings.Contains(hist, "ping") {
		t.Fatalf("history missing input: %s", hist)
	}
	view := m.vps[m.indexOf(secondID)].View()
	if !strings.Contains(view, "msg1") {
		t.Fatalf("viewport not updated: %s", view)
	}
}
