package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/internal/tool"
)

func TestNewTeam(t *testing.T) {
	ag := core.New(router.Rules{{IfContains: []string{""}, Client: nil}}, tool.Registry{}, memory.NewInMemory(), nil, memory.NewInMemoryVector(), nil)
	tm, err := NewTeam(ag, 2, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tm.vps) != 2 {
		t.Fatalf("expected 2 panes")
	}
	model, _ := tm.Update(tea.WindowSizeMsg{Width: 80, Height: 20})
	tm = model.(TeamModel)
	view := tm.View()
	if !strings.Contains(view, "Agent1") || !strings.Contains(view, "idle") {
		t.Fatalf("panel missing agent info: %s", view)
	}
}

func TestTeamSpinnerAndProgress(t *testing.T) {
	ag := core.New(router.Rules{{IfContains: []string{""}, Client: nil}}, tool.Registry{}, memory.NewInMemory(), nil, memory.NewInMemoryVector(), nil)
	tm, err := NewTeam(ag, 1, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m, _ := tm.Update(startMsg{idx: 0})
	tm = m.(TeamModel)
	if !tm.running[0] {
		t.Fatalf("spinner not started")
	}

	prev := tm.spinners[0].View()
	m, _ = tm.Update(spinner.TickMsg{})
	tm = m.(TeamModel)
	if tm.spinners[0].View() == prev {
		t.Fatalf("spinner not updated")
	}

	m, _ = tm.Update(teamMsg{idx: 0, text: "ok"})
	tm = m.(TeamModel)
	if tm.tokens[0] == 0 {
		t.Fatalf("tokens not recorded")
	}
	if tm.bars[0].Percent() == 0 {
		t.Fatalf("progress not updated")
	}
}
