package tui

import (
	"testing"

	"github.com/bradleyjkemp/cupaloy/v2"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/internal/tool"
)

// newTestAgent creates a minimal agent for TUI tests.
func newTestAgent() *core.Agent {
	return core.New(router.Rules{{IfContains: []string{""}, Client: nil}}, tool.Registry{}, memory.NewInMemory(), nil, memory.NewInMemoryVector(), nil)
}

func TestThemeRenderingSnapshots(t *testing.T) {
	ag := newTestAgent()
	for _, mode := range []string{"dark", "light"} {
		t.Run(mode, func(t *testing.T) {
			t.Setenv("AGENTRY_THEME", mode)
			m := New(ag)
			nm, _ := m.Update(tea.WindowSizeMsg{Width: 40, Height: 10})
			m = nm.(Model)
			cupaloy.SnapshotT(t, m.View())
		})
	}
}
