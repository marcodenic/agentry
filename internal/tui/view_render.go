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

	// Update dynamic input height based on visual rows (wrap-aware auto-grow)
	rows := 1
	if v := m.input.Value(); v != "" {
		// Use the textarea's actual width to match its internal wrapping
		w := m.width - 3
		if w < 1 {
			w = 1
		}
		rows = 0
		for _, line := range strings.Split(v, "\n") {
			lw := lipgloss.Width(line)
			if lw <= 0 {
				rows += 1
				continue
			}
			// ceil(lw / w) visual lines for this paragraph
			rows += (lw + w - 1) / w
		}
		if rows < 1 {
			rows = 1
		}
	}

	// Ensure the input doesn't push the rest of the UI off-screen.
	// Reserve a minimum viewport height for the chat/agents section.
	minChatRows := m.height / 2
	if minChatRows < 8 {
		minChatRows = 8
	}
	// Reserve rows for: horizontal line, spacer line, and the status bar
	reservedRows := 1 + 1 + 1
	maxInputRows := m.height - (reservedRows + minChatRows)
	// Hard cap input to 10 visual rows per requirement
	if maxInputRows > 10 {
		maxInputRows = 10
	}
	if maxInputRows < 1 {
		maxInputRows = 1
	}
	if rows > maxInputRows {
		rows = maxInputRows
	}

	// Keep input height exactly equal to calculated row count (no cushion) to avoid blank lines
	if rows != m.inputHeight {
		m.inputHeight = rows
		m.input.SetHeight(rows)
	}

	// Dynamically update viewport heights based on current input height so layout adapts as you type
	// Subtract: horizontal line (1) + input rows + spacer line (1) + status bar (1)
	viewportHeight := m.height - (1 + rows + 1 + 1)
	if viewportHeight < minChatRows {
		// Keep at least the minimum chat rows when possible
		if m.height > (reservedRows + minChatRows) {
			viewportHeight = minChatRows
		} else if viewportHeight < 3 {
			viewportHeight = 3
		}
	}
	if viewportHeight < 3 {
		viewportHeight = 3
	}
	// Apply only if changed to avoid unnecessary churn
	if m.vp.Height != viewportHeight {
		m.vp.Height = viewportHeight
		m.debugVp.Height = viewportHeight
	}

	// Render input as-is to avoid double-wrapping/cropping by lipgloss
	inputSection := m.input.View()

	// Stack everything vertically
	content := lipgloss.JoinVertical(lipgloss.Left, topSection, horizontalLine, inputSection)

	// Calculate total tokens and cost across all agents
	totalTokens := 0
	totalCost := 0.0
	// Track session-level budgets (assume shared across agents; take max)
	var budgetTokens int
	var budgetDollars float64

	// Sum up tokens and costs, using live streaming counts when available
	for _, info := range m.infos {
		if info.Agent != nil && info.Agent.Cost != nil {
			// Use streaming token count during active streaming, real count otherwise
			if info.TokensStarted && info.StreamingResponse != "" {
				totalTokens += info.StreamingTokenCount
			} else {
				totalTokens += info.Agent.Cost.TotalTokens()
			}
			totalCost += info.Agent.Cost.TotalCost()
			if info.Agent.Cost.BudgetTokens > budgetTokens {
				budgetTokens = info.Agent.Cost.BudgetTokens
			}
			if info.Agent.Cost.BudgetDollars > budgetDollars {
				budgetDollars = info.Agent.Cost.BudgetDollars
			}
		}
	}

	// Only show agent/cost/cwd/tokens info in the status bar, not workspace events
	agentsDisplay := fmt.Sprintf("◆ agents: %d", len(m.infos))
	cwdDisplay := fmt.Sprintf("⌂ cwd: %s", m.cwd)
	tokensDisplay := fmt.Sprintf("◈ tokens: %d", totalTokens)
	costDisplay := fmt.Sprintf("◎ cost: $%.6f", totalCost)

	// Append budget warnings (soft at 80%, hard at 100%)
	softPct := 0.80
	// Budget by tokens
	if budgetTokens > 0 {
		if totalTokens >= budgetTokens {
			tokensDisplay += " ⛔"
		} else if float64(totalTokens) >= float64(budgetTokens)*softPct {
			tokensDisplay += " ⚠"
		}
	}
	// Budget by dollars
	if budgetDollars > 0.0 {
		if totalCost >= budgetDollars {
			costDisplay += " ⛔"
		} else if totalCost >= budgetDollars*softPct {
			costDisplay += " ⚠"
		}
	}
	m.statusBarModel.SetSize(m.width)
	m.statusBarModel.SetContent(agentsDisplay, cwdDisplay, tokensDisplay, costDisplay)
	footer := m.statusBarModel.View()

	// Restore: input section, single blank line, then status bar flush at the bottom
	return lipgloss.JoinVertical(lipgloss.Left, content, "", footer)
}
