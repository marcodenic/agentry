package tui

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/marcodenic/agentry/internal/core"
)

// renderMemory formats an agent's memory history and trace for display.
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
				Render("� AGENT REASONING & ANALYSIS:"))
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
				toolHeader := fmt.Sprintf("▶ Tool #%d: %s", j+1, tc.Name)
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
						Render("   ❌ No result available"))
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
