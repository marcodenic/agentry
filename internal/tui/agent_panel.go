package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/marcodenic/agentry/internal/glyphs"
)

// agentPanel renders the sidebar showing all agents and their status.
func (m Model) agentPanel(panelWidth int) string {
	var lines []string

	// Use our cute robot for the title instead of triangle
	var titleGlyph string
	if m.robot != nil {
		titleGlyph = m.robot.GetStyledFace()
	} else {
		titleGlyph = glyphs.OrangeTriangle()
	}

	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.PanelTitleColor)).
		Bold(true).
		Render(titleGlyph + " AGENTS")
	lines = append(lines, title)

	totalTokens := 0
	runningCount := 0
	for _, ag := range m.infos {
		// Get token count from agent's cost manager for accuracy
		if ag.Agent != nil && ag.Agent.Cost != nil {
			totalTokens += ag.Agent.Cost.TotalTokens()
		}
		if ag.Status == StatusRunning {
			runningCount++
		}
	}
	statsLine := fmt.Sprintf("Total: %d | Running: %d", len(m.infos), runningCount)
	lines = append(lines, lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.Palette.Foreground)).
		Faint(true).
		Render(statsLine))
	lines = append(lines, "")

	for i, id := range m.order {
		ag := m.infos[id]

		var nameLine string
		statusDot := m.getAdvancedStatusDot(ag.Status)
		agentIndex := fmt.Sprintf("[%d]", i)
		if ag.Status == StatusRunning {
			nameLine = fmt.Sprintf("%s %s %s %s", agentIndex, statusDot, ag.Spinner.View(), ag.Name)
		} else {
			nameLine = fmt.Sprintf("%s %s %s", agentIndex, statusDot, ag.Name)
		}
		if id == m.active {
			// Use orange triangle for active agent (including Agent 0)
			nameLine = lipgloss.NewStyle().
				Foreground(lipgloss.Color(m.theme.UserBarColor)).
				Bold(true).
				Render(glyphs.OrangeTriangle() + " " + nameLine)
		}
		lines = append(lines, nameLine)

		if ag.Role != "" {
			roleLine := fmt.Sprintf("  role: %s", ag.Role)
			roleLine = lipgloss.NewStyle().
				Foreground(lipgloss.Color(m.theme.RoleColor)).
				Italic(true).
				Render(roleLine)
			lines = append(lines, roleLine)
		}

		if ag.CurrentTool != "" {
			toolLine := fmt.Sprintf("  %s %s", glyphs.YellowStar(), ag.CurrentTool)
			toolLine = lipgloss.NewStyle().
				Foreground(lipgloss.Color(m.theme.ToolColor)).
				Render(toolLine)
			lines = append(lines, toolLine)
		}

		if ag.ModelName != "" {
			modelLine := fmt.Sprintf("  model: %s", ag.ModelName)
			modelLine = lipgloss.NewStyle().
				Foreground(lipgloss.Color(m.theme.Palette.Foreground)).
				Faint(true).
				Render(modelLine)
			lines = append(lines, modelLine)
		}

		maxTokens := 8000
		if ag.ModelName != "" {
			// Use pricing data to get the actual context limit
			maxTokens = m.pricing.GetContextLimit(ag.ModelName)
		}

		// Get token count - use streaming count during active streaming, real count otherwise
		actualTokens := 0
		if ag.Agent != nil && ag.Agent.Cost != nil {
			if ag.TokensStarted && ag.StreamingResponse != "" {
				actualTokens = ag.StreamingTokenCount
			} else {
				actualTokens = ag.Agent.Cost.TotalTokens()
			}
		}

		tokenPct := float64(actualTokens) / float64(maxTokens) * 100
		tokenLine := fmt.Sprintf("  tokens: %d/%d", actualTokens, maxTokens)
		lines = append(lines, tokenLine)

		// The progress bar percentage will be set in renderTokenBar
		// to avoid double-setting which might cause color issues
		bar := m.renderTokenBar(ag, tokenPct, panelWidth)
		lines = append(lines, "  "+bar)
		activityChart := m.renderActivityChart(ag.ActivityData, panelWidth)
		if activityChart != "" {
			activityLabel := lipgloss.NewStyle().
				Foreground(lipgloss.Color(m.theme.Palette.Foreground)).
				Faint(true).
				Render("  activity:")
			lines = append(lines, activityLabel)
			lines = append(lines, "  "+activityChart)
		}

		if ag.Agent != nil && ag.Agent.Cost != nil {
			// Simply get the current cost directly from the agent's cost manager
			individualCost := ag.Agent.Cost.TotalCost()

			if individualCost > 0 {
				costLine := fmt.Sprintf("  cost: $%.6f", individualCost)
				costLine = lipgloss.NewStyle().
					Foreground(lipgloss.Color(m.theme.AIBarColor)).
					Render(costLine)
				lines = append(lines, costLine)
			}
		}

		lines = append(lines, "")
	}

	if len(m.infos) > 0 {
		lines = append(lines, lipgloss.NewStyle().
			Foreground(lipgloss.Color(m.theme.Palette.Foreground)).
			Faint(true).
			Render("Controls:"))
		lines = append(lines, lipgloss.NewStyle().
			Foreground(lipgloss.Color(m.theme.Palette.Foreground)).
			Faint(true).
			Render("  "+glyphs.ArrowLeft+glyphs.ArrowRight+" cycle agents"))
		lines = append(lines, lipgloss.NewStyle().
			Foreground(lipgloss.Color(m.theme.Palette.Foreground)).
			Faint(true).
			Render("  Tab switch view"))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
