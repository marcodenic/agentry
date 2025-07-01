package tui

import (
	"github.com/charmbracelet/lipgloss"
)

func (m Model) userBar() string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.UserBarColor)).Render("┃")
}

func (m Model) aiBar() string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.AIBarColor)).Render("┃")
}

func (m Model) thinkingBar() string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.AIBarColor)).Render("🤔")
}

// statusBar returns orange horizontal bar for in-progress status updates
func (m Model) statusBar() string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8C00")).Render("┃") // Orange color
}

// completedStatusBar returns green horizontal bar for completed status updates
func (m Model) completedStatusBar() string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#32CD32")).Render("┃") // Green color
}
