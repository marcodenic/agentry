package tui

import (
	"github.com/charmbracelet/lipgloss"
)

func (m Model) userBar() string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.UserBarColor)).Bold(true).Render("┃")
}

func (m Model) aiBar() string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.AIBarColor)).Bold(true).Render("┃")
}

// statusBar returns orange horizontal bar for in-progress status updates
func (m Model) statusBar() string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8C00")).Bold(true).Render("┃") // Orange color
}

// completedStatusBar returns green horizontal bar for completed status updates
func (m Model) completedStatusBar() string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#32CD32")).Bold(true).Render("┃") // Green color
}
