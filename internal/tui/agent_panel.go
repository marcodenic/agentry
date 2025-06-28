package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// agentPanel renders the sidebar showing all agents and their status.
func (m Model) agentPanel() string {
	var lines []string

	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.PanelTitleColor)).
		Bold(true).
		Render("🤖 AGENTS")
	lines = append(lines, title)

	totalTokens := 0
	runningCount := 0
	for _, ag := range m.infos {
		totalTokens += ag.TokenCount
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
			nameLine = lipgloss.NewStyle().
				Foreground(lipgloss.Color(m.theme.UserBarColor)).
				Bold(true).
				Render("▶ " + nameLine)
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
			toolLine := fmt.Sprintf("  🔧 %s", ag.CurrentTool)
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
		if ag.ModelName != "" && strings.Contains(strings.ToLower(ag.ModelName), "gpt-4") {
			maxTokens = 128000
		}
		tokenPct := float64(ag.TokenCount) / float64(maxTokens) * 100
		tokenLine := fmt.Sprintf("  tokens: %d (%.1f%%)", ag.TokenCount, tokenPct)
		lines = append(lines, tokenLine)
		bar := m.renderTokenBar(ag.TokenCount, maxTokens)
		lines = append(lines, "  "+bar)
		activityChart := m.renderActivityChart(ag.ActivityData, ag.ActivityTimes)
		if activityChart != "" {
			activityPrefix := lipgloss.NewStyle().
				Foreground(lipgloss.Color(m.theme.Palette.Foreground)).
				Faint(true).
				Render("  activity: ")
			lines = append(lines, activityPrefix+activityChart)
		}

		if ag.Agent.Cost != nil && ag.Agent.Cost.TotalCost() > 0 {
			costLine := fmt.Sprintf("  cost: $%.4f", ag.Agent.Cost.TotalCost())
			costLine = lipgloss.NewStyle().
				Foreground(lipgloss.Color(m.theme.AIBarColor)).
				Render(costLine)
			lines = append(lines, costLine)
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
			Render("  ←→ cycle agents"))
		lines = append(lines, lipgloss.NewStyle().
			Foreground(lipgloss.Color(m.theme.Palette.Foreground)).
			Faint(true).
			Render("  Tab switch view"))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
