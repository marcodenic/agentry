package tui

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/glyphs"
)

// renderDetailedMemory provides a detailed trace when debug data is available.
func (m Model) renderDetailedMemory(ag *core.Agent) string {
	info := m.infos[ag.ID]
	if info == nil || len(info.DebugTrace) == 0 {
		return renderMemory(ag)
	}

	var b bytes.Buffer

	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FFFF")).
		Bold(true).
		Render("🔍 COMPREHENSIVE AGENT DEBUG TRACE"))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Italic(true).
		Render("Real-time trace of all agent decisions, tool calls, and system events"))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("═", 80))
	b.WriteString("\n\n")

	if info.DebugStreamingResponse != "" {
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("#88FF88")).
			Bold(true).
			Render(glyphs.BlueCircle() + " CURRENT RESPONSE BEING GENERATED:"))
		b.WriteString("\n")

		responseBox := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#1a3a1a")).
			Padding(1, 2).
			Margin(0, 0, 1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#44AA44"))

		b.WriteString(responseBox.Render(truncateString(info.DebugStreamingResponse, 300)))
		b.WriteString("\n\n")
	}

	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Italic(true).
		Render("📈 EVENT TIMELINE:"))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", 80))
	b.WriteString("\n\n")

	currentStep := 0
	tokenCount := 0
	for i, event := range info.DebugTrace {
		if event.StepNum > currentStep {
			currentStep = event.StepNum
			stepHeader := fmt.Sprintf("📋 STEP %d", currentStep)
			b.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFF00")).
				Bold(true).
				Background(lipgloss.Color("#333333")).
				Padding(0, 1).
				Render(stepHeader))
			b.WriteString("\n")
			b.WriteString(strings.Repeat("─", 60))
			b.WriteString("\n")
		}

		timeStr := event.Timestamp.Format("15:04:05.000")
		eventHeader := fmt.Sprintf("[%s] %s", timeStr, event.Type)

		var eventColor string
		var eventIcon string
		switch event.Type {
		case "model_start":
			eventColor = "#88FF88"
			eventIcon = "🤖"
		case "step_start":
			eventColor = "#FFFF88"
			eventIcon = "🚀"
		case "tool_start":
			eventColor = "#FF8888"
			eventIcon = "🔧"
		case "tool_end":
			eventColor = "#88FFFF"
			eventIcon = "✅"
		case "token":
			eventColor = "#CCCCCC"
			eventIcon = "💬"
			tokenCount++
		case "final":
			eventColor = "#FF88FF"
			eventIcon = "🏁"
		default:
			eventColor = "#AAAAAA"
			eventIcon = "📌"
		}

		if event.Type == "token" {
			if tokenCount <= 5 {
				if strings.Contains(event.Details, "Character:") {
					tokenCount++
					continue
				}
			} else if tokenCount > 5 {
				hasMoreTokens := i < len(info.DebugTrace)-1
				if hasMoreTokens {
					nextEvent := info.DebugTrace[i+1]
					if nextEvent.Type == "token" {
						tokenCount++
						continue
					}
				}
				if tokenCount > 5 {
					b.WriteString(fmt.Sprintf("  💬 %s\n",
						lipgloss.NewStyle().Foreground(lipgloss.Color(eventColor)).Render(
							fmt.Sprintf("Response Generated [%d tokens streamed]", tokenCount))))
					tokenCount = 0
					continue
				}
			}
		}

		b.WriteString(fmt.Sprintf("  %s ", eventIcon))
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color(eventColor)).
			Bold(true).
			Render(eventHeader))
		b.WriteString("\n")

		if event.Details != "" {
			detailsBox := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#1a1a1a")).
				Padding(0, 1).
				Margin(0, 0, 0, 4).
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("#444444"))

			details := event.Details
			if len(details) > 200 {
				details = details[:197] + "..."
			}

			b.WriteString(detailsBox.Render(details))
			b.WriteString("\n")
		}

		b.WriteString("\n")
	}

	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Italic(true).
		Render(fmt.Sprintf("📊 Total events: %d | Steps completed: %d", len(info.DebugTrace), currentStep)))

	return b.String()
}
