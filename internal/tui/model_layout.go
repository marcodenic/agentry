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
	// Update input width to align with effective content width assumptions (see view_render.go)
	m.input.SetWidth(msg.Width - 1)

	// Calculate chat area dimensions
	chatWidth := int(float64(msg.Width)*0.75) - 2 // 75% width for chat area

	// Calculate viewport height using dynamic input rows with spacer and status bar pinned at bottom:
	// Total height - horizontal separator (1) - input section height (dynamic) - spacer line (1) - status bar (1)
	inputRows := m.inputHeight
	if inputRows < 1 {
		inputRows = 1
	}
	viewportHeight := msg.Height - (1 + inputRows + 1 + 1)
	if viewportHeight < 3 {
		viewportHeight = 3
	}

	// Set chat viewport size
	m.vp.Width = chatWidth
	m.vp.Height = viewportHeight

	// Also set debug viewport size
	m.debugVp.Width = chatWidth
	m.debugVp.Height = viewportHeight

	// Set agent panel size (25% width)
	m.tools.SetSize(int(float64(msg.Width)*0.25)-2, viewportHeight)

	// Update status bar size
	m.statusBarModel.SetSize(msg.Width)

	// Update progress bar widths for all agents when window resizes
	panelWidth := int(float64(msg.Width) * 0.25)
	for _, info := range m.infos {
		// Use same width calculation as activity chart: panelWidth - 8
		// This accounts for "  " prefix (2 chars) + " XX%" suffix (4 chars) + padding (2 chars)
		barWidth := panelWidth - 8
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
