package tui

import (
	"regexp"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

var snapshotAnsiStrip = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSISnapshot(s string) string { return snapshotAnsiStrip.ReplaceAllString(s, "") }

func TestStatusBarShowsTokenBudgetWarnings(t *testing.T) {
	ag := newTestAgent()
	ag.Cost.BudgetTokens = 50
	ag.Cost.AddModelUsage("openai/gpt-4", 45, 0)

	m := New(ag)
	m.view.Chat.SetShowInitialLogo(false)

	nm, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = nm.(Model)
	view := stripANSISnapshot(m.View())

	if !strings.Contains(view, "tokens: 45 in / 0 out (45 total)") {
		t.Fatalf("expected rendered view to list token usage, got: %s", view)
	}
	if !strings.Contains(view, "⚠") {
		t.Fatalf("expected warning indicator when usage over 80%% budget, got: %s", view)
	}
}

func TestViewDisplaysAgentCountAndCwd(t *testing.T) {
	ag := newTestAgent()
	m := New(ag)
	m.view.Chat.SetShowInitialLogo(false)

	nm, _ := m.Update(tea.WindowSizeMsg{Width: 90, Height: 28})
	m = nm.(Model)

	view := stripANSISnapshot(m.View())
	if !strings.Contains(view, "agents: 1") {
		t.Fatalf("expected rendered view to include agents summary, got: %s", view)
	}
	if !strings.Contains(view, "cwd:") {
		t.Fatalf("expected rendered view to include cwd path, got: %s", view)
	}
}
