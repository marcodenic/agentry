package tui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/NimbleMarkets/ntcharts/sparkline"
	"github.com/charmbracelet/lipgloss"
	"github.com/marcodenic/agentry/internal/glyphs"
)

// getAdvancedStatusDot renders a status icon with emoji.
func (m Model) getAdvancedStatusDot(status AgentStatus) string {
	switch status {
	case StatusIdle:
		return lipgloss.NewStyle().Foreground(lipgloss.Color(uiColorIdleHex)).Bold(true).Render(glyphs.CircleFilled)
	case StatusRunning:
		return lipgloss.NewStyle().Foreground(lipgloss.Color(uiColorRunningHex)).Bold(true).Render(glyphs.CircleFilled)
	case StatusError:
		return lipgloss.NewStyle().Foreground(lipgloss.Color(uiColorErrorHex)).Bold(true).Render(glyphs.Crossmark)
	case StatusStopped:
		return lipgloss.NewStyle().Foreground(lipgloss.Color(uiColorStoppedHex)).Bold(true).Render(glyphs.CircleEmpty)
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Bold(true).Render(glyphs.CircleEmpty)
	}
}

// renderTokenBar draws an animated progress bar for token usage with green-to-red gradient.
// renderTokenBar now computes tokenPct internally to avoid duplicated logic at call sites.
func (m Model) renderTokenBar(info *AgentInfo, panelWidth int) string {
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

	// Get the base progress bar (without percentage)
	barStr := info.TokenProgress.View()

	// Re-compute percentage here to remove duplication in caller
	maxTokens := 8000
	if info.ModelName != "" {
		maxTokens = m.pricing.GetContextLimit(info.ModelName)
	}
	liveTokens := info.LiveTokenCount()
	var tokenPct float64
	if maxTokens > 0 {
		tokenPct = float64(liveTokens) / float64(maxTokens) * 100
	}
	// Add our own styled percentage to match the activity chart
	pctText := fmt.Sprintf(" %2.0f%%", tokenPct)
	pctStyled := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Faint(true).
		Render(pctText)

	return barStr + pctStyled
}

// renderActivityChart shows recent activity levels for both output and input tokens.
func (m Model) renderActivityChart(info *AgentInfo, panelWidth int) string {
	outputData := info.OutputActivityData
	inputData := info.InputActivityData
	if len(outputData) == 0 && len(inputData) == 0 {
		return ""
	}

	// Calculate available width for the chart:
	chartWidth := panelWidth - 8
	if chartWidth < 10 {
		chartWidth = 10 // Minimum chart width
	}
	if chartWidth > 50 {
		chartWidth = 50 // Maximum chart width for readability
	}

	labelWidth := 3                               // width of "out"/"in" labels (already padded to 3 chars)
	sparklineWidth := chartWidth - labelWidth - 1 // account for label and separating space
	if sparklineWidth < 4 {
		sparklineWidth = 4
	}
	effectiveWidth := sparklineWidth + 1 // ntcharts appends a trailing space we trim later
	spacePattern := regexp.MustCompile(` \x1b\[0m$`)

	renderSeries := func(label string, data []float64, style lipgloss.Style) string {
		chart := sparkline.New(effectiveWidth, 1,
			sparkline.WithMaxValue(1.0),
			sparkline.WithStyle(style),
		)

		series := data
		if len(series) > effectiveWidth {
			series = series[len(series)-effectiveWidth:]
		}
		if len(series) < effectiveWidth {
			padding := effectiveWidth - len(series)
			for i := 0; i < padding; i++ {
				chart.Push(0.0)
			}
		}
		for _, v := range series {
			chart.Push(v)
		}

		chart.DrawBraille()
		sparklineStr := chart.View()
		sparklineStr = spacePattern.ReplaceAllString(sparklineStr, "\x1b[0m")

		currentValue := 0.0
		if len(data) > 0 {
			currentValue = data[len(data)-1]
		}
		pctText := fmt.Sprintf(" %2.0f%%", currentValue*100)
		pctStyled := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Faint(true).
			Render(pctText)

		labelStyled := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Faint(true).
			Render(fmt.Sprintf("%-3s", label))

		var builder strings.Builder
		builder.WriteString(labelStyled)
		builder.WriteString(" ")
		builder.WriteString(sparklineStr)
		builder.WriteString(pctStyled)
		return builder.String()
	}

	lines := []string{}
	lines = append(lines, renderSeries("out", outputData, lipgloss.NewStyle().Foreground(lipgloss.Color("#22C55E"))))
	lines = append(lines, renderSeries("in", inputData, lipgloss.NewStyle().Foreground(lipgloss.Color("#3B82F6"))))

	return strings.Join(lines, "\n")
}
