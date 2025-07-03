package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	var chatContent string
	if m.activeTab == 0 {
		// Use viewport content directly for proper scrolling
		chatContent = m.vp.View()

		// Special handling for centered logo
		if m.showInitialLogo {
			if info, ok := m.infos[m.active]; ok {
				// Apply centering to the logo content
				logoStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color(m.theme.Palette.Foreground)).
					Width(m.vp.Width).
					Height(m.vp.Height).
					Align(lipgloss.Center, lipgloss.Center)
				chatContent = logoStyle.Render(info.History)
			}
		}
	} else {
		// Use debug viewport for proper scrolling in debug mode
		if info, ok := m.infos[m.active]; ok {
			// Update debug viewport content if not already set
			if m.debugVp.TotalLineCount() == 0 {
				debugContent := m.renderDetailedMemory(info.Agent)
				m.debugVp.SetContent(debugContent)
			}
		}
		chatContent = m.debugVp.View()
	}

	base := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.Palette.Foreground))

	// Create top section with chat (left) and agents (right)
	// Don't apply extra width constraints to chatContent - let viewport handle it
	left := base.Width(int(float64(m.width) * 0.75)).Render(chatContent)
	rightWidth := int(float64(m.width) * 0.25)
	right := base.Width(rightWidth).Render(m.agentPanel(rightWidth))
	topSection := lipgloss.JoinHorizontal(lipgloss.Top, left, right)

	// Add full-width horizontal line
	horizontalLine := strings.Repeat("─", m.width)
	if m.width < 0 {
		horizontalLine = ""
	}

	// Add full-width input
	inputSection := base.Width(m.width).Render(m.input.View())

	// Stack everything vertically
	content := lipgloss.JoinVertical(lipgloss.Left, topSection, horizontalLine, inputSection)

	// Calculate total tokens and cost across all agents
	totalTokens := 0
	totalCost := 0.0

	// Calculate total cost by summing individual agent costs (same logic as agent panel)
	const CostPerToken = 0.000002 // Same as in agent_panel.go
	for _, info := range m.infos {
		totalTokens += info.TokenCount
		if info.Agent.Cost != nil {
			// Use the same cost calculation as individual agent display
			individualCost := float64(info.TokenCount) * CostPerToken
			totalCost += individualCost
		}
	}

	footerText := fmt.Sprintf("cwd: %s | agents: %d | tokens: %d cost: $%.4f", m.cwd, len(m.infos), totalTokens, totalCost)
	footer := base.Width(m.width).Align(lipgloss.Right).Render(footerText)

	// Add empty line between input and footer
	return lipgloss.JoinVertical(lipgloss.Left, content, "", footer)
}
