package tui

import (
	"regexp"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/tool"
)

var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string {
	return ansiRegexp.ReplaceAllString(s, "")
}

func makeTestAgent(t *testing.T, name string) *core.Agent {
	t.Helper()
	ag := core.New(model.NewMock(), name, tool.Registry{}, memory.NewInMemory(), memory.NewInMemoryVector(), nil)
	return ag
}

func attachAgent(m Model, ag *core.Agent, name, history string) Model {
	info := newAgentInfo(ag, history, 120)
	if name != "" {
		info.Name = name
	}
	if history != "" {
		info.History = history
		info.LastContentType = ContentTypeStatusMessage
	}
	m.agents = append(m.agents, ag)
	m.order = append(m.order, ag.ID)
	m.infos[ag.ID] = info
	return m
}

func TestAgentNavigationKeys(t *testing.T) {
	ag0 := makeTestAgent(t, "agent-zero")
	ag1 := makeTestAgent(t, "agent-one")

	m := New(ag0)

	info0 := m.infos[ag0.ID]
	info0.History = "Agent 0 ready"
	info0.LastContentType = ContentTypeStatusMessage
	m.infos[ag0.ID] = info0
	m.view.Chat.Main.SetContent(info0.History)
	m.view.Chat.SetShowInitialLogo(false)

	m = attachAgent(m, ag1, "Agent 1", "Agent 1 ready")

	nm, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlN})
	m = nm.(Model)
	if m.active != ag1.ID {
		t.Fatalf("expected active agent to switch to agent-one")
	}
	if !strings.Contains(stripANSI(m.view.Chat.Main.View()), "Agent 1 ready") {
		t.Fatalf("chat viewport should show agent 1 history after navigation")
	}

	nm, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlP})
	m = nm.(Model)
	if m.active != ag0.ID {
		t.Fatalf("expected active agent to cycle back to agent-zero")
	}
	if !strings.Contains(stripANSI(m.view.Chat.Main.View()), "Agent 0 ready") {
		t.Fatalf("chat viewport should show agent 0 history after cycling back")
	}
}

func TestToggleTabLoadsDebugContent(t *testing.T) {
	ag := makeTestAgent(t, "agent-debug")
	m := New(ag)
	m.view.Chat.SetShowInitialLogo(false)

	info := m.infos[ag.ID]
	info.History = "Agent debug history"
	info.LastContentType = ContentTypeStatusMessage
	info.DebugTrace = append(info.DebugTrace, DebugTraceEvent{Type: "tool", Details: "executed"})
	info.DebugStreamingResponse = "streaming"
	m.infos[ag.ID] = info

	nm, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = nm.(Model)

	if m.layout.activeTab != 1 {
		t.Fatalf("expected to be viewing debug tab, got %d", m.layout.activeTab)
	}

	debugView := stripANSI(m.view.Chat.Debug.View())
	if !strings.Contains(debugView, "EVENT TIMELINE") {
		t.Fatalf("expected debug viewport to include event timeline, got: %s", debugView)
	}
}

func TestWindowResizeAdjustsLayout(t *testing.T) {
	ag := makeTestAgent(t, "agent-layout")
	m := New(ag)

	m.view.Chat.SetShowInitialLogo(false)
	info := m.infos[ag.ID]
	info.History = "Layout history"
	info.LastContentType = ContentTypeStatusMessage
	m.infos[ag.ID] = info
	m.view.Chat.Main.SetContent(info.History)

	nm, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = nm.(Model)

	if m.view.Chat.Main.Width != 88 {
		t.Fatalf("chat width mismatch: got %d", m.view.Chat.Main.Width)
	}
	if m.view.Chat.Main.Height != 36 {
		t.Fatalf("chat height mismatch: got %d", m.view.Chat.Main.Height)
	}
	if m.view.Chat.Debug.Width != 88 || m.view.Chat.Debug.Height != 36 {
		t.Fatalf("debug viewport size mismatch: %+v", struct{ W, H int }{m.view.Chat.Debug.Width, m.view.Chat.Debug.Height})
	}

	info = m.infos[ag.ID]
	if info.TokenProgress.Width != 22 {
		t.Fatalf("expected token progress width to be 22, got %d", info.TokenProgress.Width)
	}
	if m.view.Chat.LastWidth() != 88 {
		t.Fatalf("expected last chat width to be recorded as 88, got %d", m.view.Chat.LastWidth())
	}
}
