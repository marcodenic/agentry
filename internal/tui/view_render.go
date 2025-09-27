package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	var chatContent string
	if m.layout.activeTab == 0 {
		// Use viewport content directly for proper scrolling
		chatContent = m.view.Chat.Main.View()

		// Special handling for centered logo
		if m.view.Chat.ShowInitialLogo() {
			if info, ok := m.infos[m.active]; ok {
				// Apply centering to the logo content
				logoStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color(uiColorForegroundHex)).
					Width(m.view.Chat.Main.Width).
					Height(m.view.Chat.Main.Height).
					Align(lipgloss.Center, lipgloss.Center)
				chatContent = logoStyle.Render(info.History)
			}
		}
	} else {
		// Use debug viewport for proper scrolling in debug mode
		if info, ok := m.infos[m.active]; ok {
			// Update debug viewport content if not already set
			if m.view.Chat.Debug.TotalLineCount() == 0 {
				debugContent := m.renderDetailedMemory(info.Agent)
				m.view.Chat.Debug.SetContent(debugContent)
			}
		}
		chatContent = m.view.Chat.Debug.View()
	}

	base := lipgloss.NewStyle().
		Foreground(lipgloss.Color(uiColorForegroundHex))

	// Create top section with chat (left) and agents (right)
	// Don't apply extra width constraints to chatContent - let viewport handle it
	left := base.Width(int(float64(m.layout.width) * 0.75)).Render(chatContent)
	rightWidth := int(float64(m.layout.width) * 0.25)
	right := base.Width(rightWidth).Render(m.agentPanel(rightWidth))
	topSection := lipgloss.JoinHorizontal(lipgloss.Top, left, right)

	// Add full-width horizontal line
	horizontalLine := strings.Repeat("─", m.layout.width)
	if m.layout.width < 0 {
		horizontalLine = ""
	}

	// Update dynamic input height based on visual rows (wrap-aware auto-grow)
	rows := 1
	if v := m.view.Input.Value(); v != "" {
		// Use the textarea's actual width to match its internal wrapping
		w := m.layout.width - 3
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
	minChatRows := m.layout.height / 2
	if minChatRows < 8 {
		minChatRows = 8
	}
	// Reserve rows for: horizontal line, spacer line, and the status bar
	reservedRows := 1 + 1 + 1
	maxInputRows := m.layout.height - (reservedRows + minChatRows)
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
	if rows != m.view.Input.Height() {
		m.view.Input.SetHeight(rows)
	}

	// Dynamically update viewport heights based on current input height so layout adapts as you type
	// Subtract: horizontal line (1) + input rows + spacer line (1) + status bar (1)
	viewportHeight := m.layout.height - (1 + rows + 1 + 1)
	if viewportHeight < minChatRows {
		// Keep at least the minimum chat rows when possible
		if m.layout.height > (reservedRows + minChatRows) {
			viewportHeight = minChatRows
		} else if viewportHeight < 3 {
			viewportHeight = 3
		}
	}
	if viewportHeight < 3 {
		viewportHeight = 3
	}
	// Apply only if changed to avoid unnecessary churn
	if m.view.Chat.Main.Height != viewportHeight {
		m.view.Chat.Main.Height = viewportHeight
		m.view.Chat.Debug.Height = viewportHeight
	}

	// Render input and scrub any ANSI sequences that force a black background/foreground
	rawInput := sanitizeInputANSI(m.view.Input.View())
	// Wrap the input with a neutral style so placeholder colouring remains intact
	inputStyle := lipgloss.NewStyle().UnsetBackground()
	inputSection := inputStyle.Render(rawInput)

	// Stack everything vertically
	content := lipgloss.JoinVertical(lipgloss.Left, topSection, horizontalLine, inputSection)

	// Calculate total tokens and cost across all agents
	totalInputTokens := 0
	totalOutputTokens := 0
	totalCost := 0.0
	// Track session-level budgets (assume shared across agents; take max)
	var budgetTokens int
	var budgetDollars float64

	// Sum up tokens and costs, using live streaming counts when available
	for _, info := range m.infos {
		if info.Agent != nil && info.Agent.Cost != nil {
			inputTokens, outputTokens, _ := info.TokenBreakdown()
			totalInputTokens += inputTokens
			totalOutputTokens += outputTokens
			totalCost += info.Agent.Cost.TotalCost()
			if info.Agent.Cost.BudgetTokens > budgetTokens {
				budgetTokens = info.Agent.Cost.BudgetTokens
			}
			if info.Agent.Cost.BudgetDollars > budgetDollars {
				budgetDollars = info.Agent.Cost.BudgetDollars
			}
		}
	}
	totalTokens := totalInputTokens + totalOutputTokens

	// Only show agent/cost/cwd/tokens info in the status bar, not workspace events
	agentsDisplay := fmt.Sprintf("◆ agents: %d", len(m.infos))
	cwdDisplay := fmt.Sprintf("⌂ cwd: %s", m.cwd)
	tokensDisplay := fmt.Sprintf("◈ tokens: %d in / %d out (%d total)", totalInputTokens, totalOutputTokens, totalTokens)
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
	m.view.Status.SetSize(m.layout.width)
	m.view.Status.SetContent(agentsDisplay, cwdDisplay, tokensDisplay, costDisplay)
	footer := m.view.Status.View()

	// Restore: input section, single blank line, then status bar flush at the bottom
	return lipgloss.JoinVertical(lipgloss.Left, content, "", footer)
}
