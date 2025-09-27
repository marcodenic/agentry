package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/marcodenic/agentry/internal/glyphs"
)

// agentPanel renders the sidebar showing all agents and their status.
func (m Model) agentPanel(panelWidth int) string {
	var lines []string

	// Use our cute robot for the title instead of triangle
	var titleGlyph string
	if m.view.Robot != nil {
		titleGlyph = m.view.Robot.GetStyledFace()
	} else {
		titleGlyph = glyphs.OrangeTriangle()
	}

	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color(uiColorPanelTitleHex)).
		Bold(true).
		Render(titleGlyph + " AGENTS")
	lines = append(lines, title)

	runningCount := 0
	for _, ag := range m.infos {
		if ag.Status == StatusRunning {
			runningCount++
		}
	}
	statsLine := fmt.Sprintf("Total: %d | Running: %d", len(m.infos), runningCount)
	lines = append(lines, lipgloss.NewStyle().
		Foreground(lipgloss.Color(uiColorForegroundHex)).
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
				Foreground(lipgloss.Color(uiColorUserAccentHex)).
				Bold(true).
				Render(glyphs.OrangeTriangle() + " " + nameLine)
		}
		lines = append(lines, nameLine)

		if ag.Role != "" {
			roleLine := fmt.Sprintf("  role: %s", ag.Role)
			roleLine = lipgloss.NewStyle().
				Foreground(lipgloss.Color(uiColorRoleAccentHex)).
				Italic(true).
				Render(roleLine)
			lines = append(lines, roleLine)
		}

		if ag.CurrentTool != "" {
			toolLine := fmt.Sprintf("  %s %s", glyphs.YellowStar(), ag.CurrentTool)
			toolLine = lipgloss.NewStyle().
				Foreground(lipgloss.Color(uiColorToolAccentHex)).
				Render(toolLine)
			lines = append(lines, toolLine)
		}

		if ag.ModelName != "" {
			modelLine := fmt.Sprintf("  model: %s", ag.ModelName)
			modelLine = lipgloss.NewStyle().
				Foreground(lipgloss.Color(uiColorForegroundHex)).
				Faint(true).
				Render(modelLine)
			lines = append(lines, modelLine)
		}

		maxTokens := 8000
		if ag.ModelName != "" {
			// Use pricing data to get the actual context limit
			maxTokens = m.pricing.GetContextLimit(ag.ModelName)
		}

		inputTokens, outputTokens, totalTokens := ag.TokenBreakdown()

		tokenLine := fmt.Sprintf("  tokens: %d in / %d out (%d/%d)", inputTokens, outputTokens, totalTokens, maxTokens)
		lines = append(lines, tokenLine)

		// The progress bar percentage is computed inside renderTokenBar
		bar := m.renderTokenBar(ag, panelWidth)
		lines = append(lines, "  "+bar)
		activityChart := m.renderActivityChart(ag, panelWidth)
		if activityChart != "" {
			activityLabel := lipgloss.NewStyle().
				Foreground(lipgloss.Color(uiColorForegroundHex)).
				Faint(true).
				Render("  activity:")
			lines = append(lines, activityLabel)
			for _, chartLine := range strings.Split(activityChart, "\n") {
				lines = append(lines, "  "+chartLine)
			}
		}

		if ag.Agent != nil && ag.Agent.Cost != nil {
			// Simply get the current cost directly from the agent's cost manager
			individualCost := ag.Agent.Cost.TotalCost()

			if individualCost > 0 {
				costLine := fmt.Sprintf("  cost: $%.6f", individualCost)
				costLine = lipgloss.NewStyle().
					Foreground(lipgloss.Color(uiColorAIAccentHex)).
					Render(costLine)
				lines = append(lines, costLine)
			}
		}

		lines = append(lines, "")
	}

	// Diagnostics summary block
	if len(m.view.Diagnostics.Entries) > 0 || m.view.Diagnostics.Running {
		title := lipgloss.NewStyle().
			Foreground(lipgloss.Color(uiColorPanelTitleHex)).
			Bold(true).
			Render("🩺 DIAGNOSTICS")
		lines = append(lines, title)
		if m.view.Diagnostics.Running {
			lines = append(lines, "  running...")
		}
		if len(m.view.Diagnostics.Entries) > 0 {
			// Aggregate counts
			errs := 0
			warns := 0
			perFile := map[string]int{}
			for _, d := range m.view.Diagnostics.Entries {
				if d.Severity == "warning" {
					warns++
				} else {
					errs++
				}
				perFile[d.File]++
			}
			lines = append(lines, fmt.Sprintf("  errors: %d  warnings: %d", errs, warns))
			// Show top 3 files
			shown := 0
			for f, c := range perFile {
				lines = append(lines, fmt.Sprintf("  • %s (%d)", f, c))
				shown++
				if shown >= 3 {
					break
				}
			}
			// Show first diagnostic preview
			d := m.view.Diagnostics.Entries[0]
			preview := fmt.Sprintf("  %s %s:%d:%d %s", severityGlyph(d.Severity), d.File, d.Line, d.Col, d.Message)
			lines = append(lines, lipgloss.NewStyle().Faint(true).Render(preview))
		}
		lines = append(lines, lipgloss.NewStyle().Faint(true).Render("  press "+m.keys.Diagnostics+" to run"))
		lines = append(lines, "")
	}

	if len(m.infos) > 0 {
		lines = append(lines, lipgloss.NewStyle().
			Foreground(lipgloss.Color(uiColorForegroundHex)).
			Faint(true).
			Render("Controls:"))
		lines = append(lines, lipgloss.NewStyle().
			Foreground(lipgloss.Color(uiColorForegroundHex)).
			Faint(true).
			Render("  "+glyphs.ArrowLeft+glyphs.ArrowRight+" cycle agents"))
		lines = append(lines, lipgloss.NewStyle().
			Foreground(lipgloss.Color(uiColorForegroundHex)).
			Faint(true).
			Render("  Tab switch view"))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func severityGlyph(sev string) string {
	switch sev {
	case "warning":
		return "⚠️"
	case "info":
		return "ℹ️"
	default:
		return "❌"
	}
}
