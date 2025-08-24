package tui

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/glyphs"
)

// renderMemory formats an agent's memory history and trace for display.
// Falls back to basic memory if debug data is unavailable.
func renderMemory(ag *core.Agent) string {
	var b bytes.Buffer

	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FFFF")).
		Bold(true).
		Render("🔍 AGENT DEBUG TRACE & MEMORY"))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Italic(true).
		Render("Detailed view of agent reasoning, tool calls, and execution flow"))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("═", 80))
	b.WriteString("\n\n")

	hist := ag.Mem.History()
	if len(hist) == 0 {
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Italic(true).
			Render("No execution history available yet..."))
		return b.String()
	}

	for i, s := range hist {
		stepHeader := fmt.Sprintf("📋 EXECUTION STEP %d", i+1)
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF00")).
			Bold(true).
			Background(lipgloss.Color("#333333")).
			Padding(0, 1).
			Render(stepHeader))
		b.WriteString("\n")
		b.WriteString(strings.Repeat("─", 60))
		b.WriteString("\n\n")

		if s.Output != "" {
			b.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF88")).
				Bold(true).
				Render("🧠 AGENT REASONING & ANALYSIS:"))
			b.WriteString("\n")

			outputBox := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#1a1a1a")).
				Padding(1, 2).
				Margin(0, 0, 1, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#444444"))

			cleanOutput := strings.TrimSpace(s.Output)
			lines := strings.Split(cleanOutput, "\n")
			var formattedLines []string
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				if trimmed != "" {
					formattedLines = append(formattedLines, trimmed)
				}
			}

			if len(formattedLines) > 0 {
				b.WriteString(outputBox.Render(strings.Join(formattedLines, "\n")))
				b.WriteString("\n\n")
			}
		}

		if len(s.ToolCalls) > 0 {
			b.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF8844")).
				Bold(true).
				Render("🔧 TOOL EXECUTIONS:"))
			b.WriteString("\n\n")

			for j, tc := range s.ToolCalls {
				toolHeader := fmt.Sprintf(glyphs.OrangeTriangle()+" Tool #%d: %s", j+1, tc.Name)
				b.WriteString(lipgloss.NewStyle().
					Foreground(lipgloss.Color("#FF6666")).
					Bold(true).
					Render(toolHeader))
				b.WriteString("\n")

				if len(tc.Arguments) > 0 {
					b.WriteString(lipgloss.NewStyle().
						Foreground(lipgloss.Color("#AAAAAA")).
						Render("   📝 Arguments:"))
					b.WriteString("\n")

					argsBox := lipgloss.NewStyle().
						Foreground(lipgloss.Color("#CCCCCC")).
						Background(lipgloss.Color("#2a2a2a")).
						Padding(0, 1).
						Margin(0, 0, 0, 6).
						Border(lipgloss.NormalBorder()).
						BorderForeground(lipgloss.Color("#555555"))

					b.WriteString(argsBox.Render(string(tc.Arguments)))
					b.WriteString("\n\n")
				}

				if result, ok := s.ToolResults[tc.ID]; ok {
					b.WriteString(lipgloss.NewStyle().
						Foreground(lipgloss.Color("#44FF88")).
						Bold(true).
						Render("   ✅ Result:"))
					b.WriteString("\n")

					resultBox := lipgloss.NewStyle().
						Foreground(lipgloss.Color("#FFFFFF")).
						Background(lipgloss.Color("#1a3a1a")).
						Padding(1, 2).
						Margin(0, 0, 1, 6).
						Border(lipgloss.RoundedBorder()).
						BorderForeground(lipgloss.Color("#44AA44"))

					cleanResult := strings.TrimSpace(result)
					if len(cleanResult) > 500 {
						cleanResult = cleanResult[:497] + "..."
					}

					b.WriteString(resultBox.Render(cleanResult))
					b.WriteString("\n")
				} else {
					b.WriteString(lipgloss.NewStyle().
						Foreground(lipgloss.Color("#FF4444")).
						Render("   " + glyphs.RedCrossmark() + " No result available"))
					b.WriteString("\n\n")
				}
			}
		}

		if i < len(hist)-1 {
			b.WriteString("\n")
			b.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color("#666666")).
				Render(strings.Repeat("═", 80)))
			b.WriteString("\n\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Italic(true).
		Render(fmt.Sprintf("📊 Summary: %d execution steps completed", len(hist))))

	return b.String()
}

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
