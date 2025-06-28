package tui

import (
	"fmt"
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

func (m Model) statusBar() string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.AIBarColor)).Render("⚡")
}

func (m Model) formatToolCompletion(toolName string, args map[string]any) string {
	switch toolName {
	case "view", "read":
		return "✅ File read"
	case "write":
		return "✅ File written"
	case "edit", "patch":
		return "✅ File edited"
	case "ls", "list":
		return "✅ Directory listed"
	case "bash", "powershell", "cmd":
		return "✅ Command completed"
	case "agent":
		if agent, ok := args["agent"].(string); ok {
			return fmt.Sprintf("✅ Delegated to %s", agent)
		}
		return "✅ Task delegated"
	case "grep", "search":
		return "✅ Search completed"
	case "fetch":
		return "✅ Data fetched"
	default:
		return "✅ Done"
	}
}
