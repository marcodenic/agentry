package tui

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/marcodenic/agentry/internal/core"
)

// userBar returns the glyph used to prefix user input.
func (m Model) userBar() string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.UserBarColor)).Render("┃")
}

// aiBar returns the glyph used to prefix agent responses.
func (m Model) aiBar() string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.AIBarColor)).Render("┃")
}

// renderMemory formats an agent's memory history for display.
func renderMemory(ag *core.Agent) string {
	hist := ag.Mem.History()
	var b bytes.Buffer
	for i, s := range hist {
		b.WriteString("Step ")
		b.WriteString(strconv.Itoa(i))
		b.WriteString(": ")
		b.WriteString(s.Output)
		for _, tc := range s.ToolCalls {
			if r, ok := s.ToolResults[tc.ID]; ok {
				b.WriteString(" -> ")
				b.WriteString(tc.Name)
				b.WriteString(": ")
				b.WriteString(r)
			}
		}
		b.WriteString("\n")
	}
	return b.String()
}

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

// helpView returns the static help text displayed at startup.
func helpView() string {
	return strings.Join([]string{
		"AGENTRY TUI - Unified Agent Interface",
		"",
		"Commands:",
		"/spawn <name> [role]     - create a new agent with optional role",
		"/switch <prefix>         - focus an agent by name or ID prefix",
		"/stop <prefix>           - stop an agent",
		"/converse <n> <topic>    - start multi-agent conversation",
		"",
		"Controls:",
		"←→ / Ctrl+P/N           - cycle between agents",
		"Tab                     - switch between chat and memory view",
		"Enter                   - send message / execute command",
		"Ctrl+C / q              - quit",
		"",
		"Agent Panel:",
		"● idle  🟡 running  ❌ error  ⏸️ stopped",
		"[index] shows agent position, ▶ shows active agent",
	}, "\n")
}

// statusDot renders a colored dot based on agent status.
func (m Model) statusDot(st AgentStatus) string {
	color := m.theme.IdleColor
	switch st {
	case StatusRunning:
		color = m.theme.RunningColor
	case StatusError:
		color = m.theme.ErrorColor
	case StatusStopped:
		color = m.theme.StoppedColor
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render("●")
}

// getAdvancedStatusDot renders a status icon with emoji.
func (m Model) getAdvancedStatusDot(status AgentStatus) string {
	switch status {
	case StatusIdle:
		return lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.IdleColor)).Render("●")
	case StatusRunning:
		return lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.RunningColor)).Render("🟡")
	case StatusError:
		return lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.ErrorColor)).Render("❌")
	case StatusStopped:
		return lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.StoppedColor)).Render("⏸️")
	default:
		return "○"
	}
}

// renderTokenBar draws a simple progress bar for token usage.
func (m Model) renderTokenBar(count, max int) string {
	if max <= 0 {
		max = 1
	}
	pct := float64(count) / float64(max)
	if pct > 1 {
		pct = 1
	}
	filled := int(pct * 10)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", 10-filled)
	return fmt.Sprintf("%s %d%%", bar, int(pct*100))
}

// renderSparkline draws a sparkline from the given history values.
func (m Model) renderSparkline(history []int) string {
	if len(history) == 0 {
		return ""
	}
	chars := []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	min, max := history[0], history[0]
	for _, v := range history {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	var b strings.Builder
	for _, v := range history {
		if max == min {
			b.WriteString(chars[0])
		} else {
			n := float64(v-min) / float64(max-min)
			idx := int(n * float64(len(chars)-1))
			b.WriteString(chars[idx])
		}
	}
	return b.String()
}

// renderActivityChart shows recent activity levels as a scrolling chart.
func (m Model) renderActivityChart(activityData []float64, activityTimes []time.Time) string {
	const chartWidth = 30
	chars := []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	chartData := make([]float64, chartWidth)
	now := time.Now()
	for i := 0; i < chartWidth; i++ {
		targetTime := now.Add(-time.Duration(2*(chartWidth-1-i)) * time.Second)
		var activity float64
		if len(activityData) > 0 && len(activityTimes) > 0 {
			minDiff := time.Hour
			for j, t := range activityTimes {
				diff := targetTime.Sub(t)
				if diff < 0 {
					diff = -diff
				}
				if diff < minDiff && diff < 5*time.Second {
					minDiff = diff
					activity = activityData[j]
				}
			}
		}
		chartData[i] = activity
	}

	var result strings.Builder
	for _, activity := range chartData {
		charIdx := int(activity * float64(len(chars)-1))
		if charIdx >= len(chars) {
			charIdx = len(chars) - 1
		}
		if charIdx < 0 {
			charIdx = 0
		}
		char := chars[charIdx]
		var color string
		if activity <= 0.1 {
			color = "#374151"
		} else if activity <= 0.3 {
			color = "#22C55E"
		} else if activity <= 0.6 {
			color = "#FBBF24"
		} else if activity <= 0.8 {
			color = "#F97316"
		} else {
			color = "#EF4444"
		}
		styledChar := lipgloss.NewStyle().
			Foreground(lipgloss.Color(color)).
			Render(char)
		result.WriteString(styledChar)
	}

	if len(activityData) > 0 {
		currentActivity := activityData[len(activityData)-1]
		pctText := fmt.Sprintf(" %2.0f%%", currentActivity*100)
		pctStyled := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Faint(true).
			Render(pctText)
		result.WriteString(pctStyled)
	} else {
		pctStyled := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Faint(true).
			Render("  0%")
		result.WriteString(pctStyled)
	}

	return result.String()
}
