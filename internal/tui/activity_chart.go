package tui

import (
	"fmt"
	"strings"

	"github.com/NimbleMarkets/ntcharts/sparkline"
	"github.com/charmbracelet/lipgloss"
	"github.com/marcodenic/agentry/internal/glyphs"
)

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

// renderTokenBar draws an animated progress bar for token usage with green-to-red gradient.
// Sets the width and percentage based on current token usage.
func (m Model) renderTokenBar(info *AgentInfo, width int) string {
	// Set the width of the progress bar to fit the sidebar (minus padding)
	barWidth := width - 6 // Account for "  " prefix and some padding
	if barWidth < 10 {
		barWidth = 10 // Minimum width
	}
	if barWidth > 50 {
		barWidth = 50 // Maximum reasonable width
	}
	info.TokenProgress.Width = barWidth

	// Calculate and set the percentage based on current token usage
	maxTokens := 8000
	if info.ModelName != "" && strings.Contains(strings.ToLower(info.ModelName), "gpt-4") {
		maxTokens = 128000
	}

	// Get token count - use streaming count during active streaming, real count otherwise
	actualTokens := 0
	if info.Agent != nil && info.Agent.Cost != nil {
		if info.TokensStarted && info.StreamingResponse != "" {
			actualTokens = info.StreamingTokenCount
		} else {
			actualTokens = info.Agent.Cost.TotalTokens()
		}
	}

	// Calculate percentage (0.0 to 1.0 range for progress bar)
	pct := float64(actualTokens) / float64(maxTokens)
	if pct < 0 {
		pct = 0
	}
	if pct > 1 {
		pct = 1
	}

	// Set the percentage on the existing progress bar
	info.TokenProgress.SetPercent(pct)

	return info.TokenProgress.View()
}

// renderActivityChart shows recent activity levels as a scrolling chart using ntcharts sparkline.
func (m Model) renderActivityChart(activityData []float64, panelWidth int) string {
	if len(activityData) == 0 {
		return ""
	}

	// Calculate available width for the chart:
	// panelWidth - "  " prefix (2 chars) - " XX%" suffix (4 chars) - padding (2 chars) = available width
	availableWidth := panelWidth - 8
	if availableWidth < 10 {
		availableWidth = 10 // Minimum chart width
	}
	if availableWidth > 50 {
		availableWidth = 50 // Maximum chart width for readability
	}

	// Create sparkline chart with height 1 for a single row
	chart := sparkline.New(availableWidth, 1,
		sparkline.WithMaxValue(1.0), // Activity is normalized 0-1
		sparkline.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#22C55E"))),
	)

	// Push the most recent data points to the sparkline
	// Take the last 'availableWidth' data points
	startIdx := len(activityData) - availableWidth
	if startIdx < 0 {
		// If we don't have enough data, pad with zeros at the beginning
		for i := 0; i < availableWidth-len(activityData); i++ {
			chart.Push(0.0)
		}
		startIdx = 0
	}

	// Add the actual data points
	for i := startIdx; i < len(activityData); i++ {
		chart.Push(activityData[i])
	}

	// Draw the Braille sparkline (for smooth, high-resolution appearance)
	chart.DrawBraille()

	// Get the rendered sparkline
	sparklineStr := chart.View()

	// Add percentage indicator
	var result strings.Builder
	result.WriteString(sparklineStr)

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
