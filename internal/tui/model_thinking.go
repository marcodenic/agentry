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
	
	// Only wrap and update display every ~100 characters to reduce CPU usage
	if len(info.ThinkingContent)%100 == 0 || len(msg.delta) > 10 {
		thinkingWidth := m.view.Chat.Main.Width - 8 // Account for border and padding
		if thinkingWidth < 20 {
			thinkingWidth = 20
		}
		wrappedContent := lipgloss.NewStyle().Width(thinkingWidth).Render(info.ThinkingContent)
		info.ThinkingViewport.SetContent(wrappedContent)
		info.ThinkingViewport.GotoBottom()
		
		// Update main viewport with thinking box (inline to avoid repeated expensive calls)
		if msg.id == m.active {
			// Create thinking section with border
			thinkingStyle := lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#9370DB")).
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
			
			m.view.Chat.Main.SetContent(displayContent)
		}
	}

	// Save updated info
	m.infos[msg.id] = info

	// Continue reading events
	return m, m.runtime.ReadCmd(&m, msg.id)
}


