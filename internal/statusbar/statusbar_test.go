package statusbar

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func TestStatusbarViewRespectsWidth(t *testing.T) {
	cfg := ColorConfig{Foreground: lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#ffffff"}, Background: lipgloss.AdaptiveColor{Light: "#111111", Dark: "#111111"}}
	sb := New(cfg, cfg, cfg, cfg)
	sb.SetSize(40)
	sb.SetContent("LEFT", strings.Repeat("middle content ", 2), "RIGHT", "END")

	view := sb.View()
	if len(view) == 0 {
		t.Fatalf("expected view to render when width > 0")
	}
	if lipgloss.Width(view) != 40 {
		t.Fatalf("expected rendered width 40, got %d", lipgloss.Width(view))
	}
}

func TestStatusbarUpdateHandlesWindowSizeMsg(t *testing.T) {
	cfg := ColorConfig{}
	sb := New(cfg, cfg, cfg, cfg)

	msg := tea.WindowSizeMsg{Width: 55}
	updated, _ := sb.Update(msg)
	if updated.Width != 55 {
		t.Fatalf("expected width to be set from window size msg")
	}
}

func TestEmptyWidthReturnsEmptyView(t *testing.T) {
	cfg := ColorConfig{}
	sb := New(cfg, cfg, cfg, cfg)
	if sb.View() != "" {
		t.Fatalf("expected zero width statusbar to render empty string")
	}
}
