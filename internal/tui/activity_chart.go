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
func (m Model) renderTokenBar(info *AgentInfo, tokenPct float64, panelWidth int) string {
	// Calculate width using the EXACT same method as renderActivityChart
	// to ensure perfect alignment
	chartWidth := panelWidth - 8 // Same calculation as activity chart
	if chartWidth < 10 {
		chartWidth = 10 // Same minimum as activity chart
	}
	if chartWidth > 50 {
		chartWidth = 50 // Same maximum as activity chart
	}

	// Set the width to match activity chart exactly
	info.TokenProgress.Width = chartWidth

	return info.TokenProgress.View()
}

// renderActivityChart shows recent activity levels as a scrolling chart using ntcharts sparkline.
func (m Model) renderActivityChart(activityData []float64, panelWidth int) string {
	if len(activityData) == 0 {
		return ""
	}

	// Calculate available width for the chart:
	// panelWidth - "  " prefix (2 chars) - " XX%" suffix (4 chars) - padding (2 chars) = available width
	chartWidth := panelWidth - 8
	if chartWidth < 10 {
		chartWidth = 10 // Minimum chart width
	}
	if chartWidth > 50 {
		chartWidth = 50 // Maximum chart width for readability
	}
	
	// Adjust sparkline width to match token bar behavior:
	// Token bar adds " 0%" (4 chars) automatically, so reduce sparkline width accordingly
	availableWidth := chartWidth - 4

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
		pctText := fmt.Sprintf(" %2.0f%%", currentActivity*100) // Removed extra spaces to match token bar
		pctStyled := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Faint(true).
			Render(pctText)
		result.WriteString(pctStyled)
	} else {
		pctStyled := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Faint(true).
			Render("  0%") // Removed extra spaces to match token bar
		result.WriteString(pctStyled)
	}

	return result.String()
}
