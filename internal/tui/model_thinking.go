package tui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// handleThinkingMessage processes reasoning/thinking content from reasoning models
// Displays as a marquee-style scrolling line that disappears when real response starts
func (m Model) handleThinkingMessage(msg thinkingMsg) (Model, tea.Cmd) {
	info := m.infos[msg.id]
	if info.Status == StatusStopped {
		return m, nil
	}

	// Accumulate thinking content
	info.ThinkingContent += msg.delta

	// Throttle display updates to prevent overwhelming the terminal
	// Only update every 50ms (20 fps max) or every 10 characters
	now := time.Now()
	shouldUpdate := len(info.ThinkingContent)%10 == 0 ||
		info.LastThinkingUpdate.IsZero() ||
		now.Sub(info.LastThinkingUpdate) > 50*time.Millisecond

	if shouldUpdate {
		info.LastThinkingUpdate = now
		// Update display if this is the active agent
		if msg.id == m.active {
			m.updateThinkingDisplay(info)
		}
	}

	// Save updated info
	m.infos[msg.id] = info

	// Continue reading events
	return m, m.runtime.ReadCmd(&m, msg.id)
}

// updateThinkingDisplay renders the marquee-style thinking line
func (m *Model) updateThinkingDisplay(info *AgentInfo) {
	if info.ThinkingContent == "" {
		return
	}

	// Clean up the thinking content - remove newlines, collapse spaces
	clean := strings.ReplaceAll(info.ThinkingContent, "\n", " ")
	clean = strings.ReplaceAll(clean, "\r", "")
	clean = strings.Join(strings.Fields(clean), " ")

	// Use thinking bar (same as AI bar but with thinking emoji prefix)
	thinkingBar := lipgloss.NewStyle().Foreground(lipgloss.Color("#9370DB")).Bold(true).Render("┃") // Purple for thinking
	prefix := thinkingBar + " 💭 "
	prefixLen := 6 // bar + space + emoji + space

	// Calculate available width for marquee (accounting for bar and prefix)
	maxWidth := m.view.Chat.Main.Width - prefixLen - 4

	if maxWidth < 20 {
		maxWidth = 40 // minimum reasonable width
	}

	// Marquee effect: show the last N characters, scrolling left
	var displayText string
	runeContent := []rune(clean)
	if len(runeContent) > maxWidth {
		// Show the last maxWidth characters (scrolling effect)
		displayText = string(runeContent[len(runeContent)-maxWidth:])
	} else {
		displayText = clean
	}

	// Style the thinking line (dimmed, italic)
	thinkingStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")). // dim gray
		Italic(true)

	thinkingLine := prefix + thinkingStyle.Render(displayText)

	// Update viewport with thinking line at the bottom
	displayHistory := info.History
	if displayHistory != "" {
		displayHistory += "\n"
	}
	displayHistory += thinkingLine

	m.view.Chat.Main.SetContent(displayHistory)
	m.view.Chat.Main.GotoBottom()
}
