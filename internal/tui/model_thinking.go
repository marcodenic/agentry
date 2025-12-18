package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// handleThinkingMessage processes reasoning/thinking content from reasoning models
// Displays in a dedicated scrolling viewport that collapses when real response starts
func (m Model) handleThinkingMessage(msg thinkingMsg) (Model, tea.Cmd) {
	info := m.infos[msg.id]
	if info.Status == StatusStopped {
		return m, nil
	}

	// Accumulate thinking content
	info.ThinkingContent += msg.delta
	
	// Show the thinking viewport and update its content
	info.ShowThinking = true
	info.ThinkingViewport.SetContent(info.ThinkingContent)
	info.ThinkingViewport.GotoBottom() // Auto-scroll thinking viewport to bottom

	// Save updated info
	m.infos[msg.id] = info

	// Don't update main viewport on every delta to prevent flashing
	// The thinking content will be visible through the next render cycle

	// Continue reading events
	return m, m.runtime.ReadCmd(&m, msg.id)
}

// getDisplayContent returns the content to display in the main viewport
// including history and optional thinking box (for rendering only)
func (m *Model) getDisplayContent(info *AgentInfo) string {
	if !info.ShowThinking || info.ThinkingContent == "" {
		return info.History
	}
	
	// Create thinking section with border
	thinkingStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#9370DB")). // Purple for thinking
		Padding(0, 1)
	
	thinkingHeader := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9370DB")).
		Bold(true).
		Render("💭 Thinking...")
	
	thinkingBody := info.ThinkingViewport.View()
	thinkingBox := thinkingStyle.Render(thinkingHeader + "\n" + thinkingBody)
	
	// Combine history with thinking box
	displayContent := info.History
	if displayContent != "" && !strings.HasSuffix(displayContent, "\n") {
		displayContent += "\n"
	}
	displayContent += "\n" + thinkingBox
	
	return displayContent
}


