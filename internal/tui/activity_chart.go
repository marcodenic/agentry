package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/marcodenic/agentry/internal/glyphs"
)

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
	return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(true).Render(glyphs.CircleFilled)
}

// getAdvancedStatusDot renders a status icon with emoji.
func (m Model) getAdvancedStatusDot(status AgentStatus) string {
	switch status {
	case StatusIdle:
		return lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.IdleColor)).Bold(true).Render(glyphs.CircleFilled)
	case StatusRunning:
		return lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.RunningColor)).Bold(true).Render(glyphs.CircleFilled)
	case StatusError:
		return lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.ErrorColor)).Bold(true).Render(glyphs.Crossmark)
	case StatusStopped:
		return lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.StoppedColor)).Bold(true).Render(glyphs.CircleEmpty)
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Bold(true).Render(glyphs.CircleEmpty)
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
