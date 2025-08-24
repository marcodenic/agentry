package tui

import (
	"context"
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/tool"
)

func TestNew(t *testing.T) {
	ag := core.New(model.NewMock(), "mock", tool.Registry{}, memory.NewInMemory(), memory.NewInMemoryVector(), nil)
	m := New(ag)
	if len(m.infos) != 1 {
		t.Fatalf("expected one agent")
	}
}

func TestModelBasicInteraction(t *testing.T) {
	mock := &seqMock{}
	ag := core.New(mock, "mock", tool.Registry{}, memory.NewInMemory(), memory.NewInMemoryVector(), nil)

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

func (m *seqMock) Stream(ctx context.Context, msgs []model.ChatMessage, tools []model.ToolSpec) (<-chan model.StreamChunk, error) {
	ch := make(chan model.StreamChunk, 1)
	go func() {
		defer close(ch)
		m.n++
		ch <- model.StreamChunk{ContentDelta: fmt.Sprintf("msg%d", m.n), Done: true}
	}()
	return ch, nil
}

func TestWindowSizing(t *testing.T) {
	mock := &seqMock{}
	ag := core.New(mock, "mock", tool.Registry{}, memory.NewInMemory(), memory.NewInMemoryVector(), nil)

	m := New(ag)

	// Simulate window sizing so viewport renders
	nm, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 20})
	m = nm.(Model)

	// Simulate entering text and pressing enter
	m.input.SetValue("ping")
	nm, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = nm.(Model)

	activeInfo := m.infos[m.active]
	hist := activeInfo.History
	if !strings.Contains(hist, "ping") {
		t.Fatalf("history missing input: %s", hist)
	}

	// Check that the agent is now running (processing the request)
	if activeInfo.Status != StatusRunning {
		t.Fatalf("agent should be running after receiving input, got status: %v", activeInfo.Status)
	}
}
