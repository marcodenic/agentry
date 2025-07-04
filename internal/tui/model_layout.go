package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// handleWindowResize processes window resize messages and updates layout
func (m Model) handleWindowResize(msg tea.WindowSizeMsg) (Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height
	m.input.Width = msg.Width - 2 // Full width minus padding for input

	// Calculate chat area dimensions
	chatWidth := int(float64(msg.Width)*0.75) - 2 // 75% width for chat area

	// Calculate viewport height more accurately:
	// Total height - top section margin - horizontal separator (1) - input section height - footer section height - padding
	viewportHeight := msg.Height - 5 // Leave space for separator, input, footer, and padding

	m.vp.Width = chatWidth
	m.vp.Height = viewportHeight

	// Also set debug viewport size
	m.debugVp.Width = chatWidth
	m.debugVp.Height = viewportHeight

	// Set agent panel size (25% width)
	m.tools.SetSize(int(float64(msg.Width)*0.25)-2, viewportHeight)

	// Update progress bar widths for all agents when window resizes
	panelWidth := int(float64(msg.Width) * 0.25)
	for _, info := range m.infos {
		barWidth := panelWidth - 6 // Account for "  " prefix and some padding
		if barWidth < 10 {
			barWidth = 10 // Minimum width
		}
		if barWidth > 50 {
			barWidth = 50 // Maximum reasonable width
		}
		info.TokenProgress.Width = barWidth
	}

	// Refresh the viewport content with proper sizing - avoid expensive reformatting
	if info, ok := m.infos[m.active]; ok {
		// For normal history, only re-wrap if width changed significantly
		if !m.showInitialLogo && info.History != "" {
			// Only reformat if width changed by more than 10 characters to avoid constant reformatting
			if m.lastWidth == 0 || (chatWidth != m.lastWidth && abs(chatWidth-m.lastWidth) > 10) {
				reformattedHistory := m.formatHistoryWithBars(info.History, chatWidth)
				m.vp.SetContent(reformattedHistory)
				m.lastWidth = chatWidth
			} else {
				// Width didn't change much, just update viewport size without reformatting
				m.vp.SetContent(info.History)
			}
		} else {
			// For initial logo or empty history, use as-is
			m.vp.SetContent(info.History)
		}
		m.vp.GotoBottom() // Ensure we're at the bottom after resize
	}
	return m, nil
}
