package tui

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

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

// statusBar returns the glyph used to prefix status messages.
func (m Model) statusBar() string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.AIBarColor)).Render("⚡")
}

// formatToolCompletion creates user-friendly completion messages
func (m Model) formatToolCompletion(toolName string, args map[string]any) string {
	switch toolName {
	case "view", "read":
		return "✅ File read"
	case "write":
		return "✅ File written"
	case "edit", "patch":
		return "✅ File edited"
	case "ls", "list":
		return "✅ Directory listed"
	case "bash", "powershell", "cmd":
		return "✅ Command completed"
	case "agent":
		if agent, ok := args["agent"].(string); ok {
			return fmt.Sprintf("✅ Delegated to %s", agent)
		}
		return "✅ Task delegated"
	case "grep", "search":
		return "✅ Search completed"
	case "fetch":
		return "✅ Data fetched"
	default:
		return "✅ Done"
	}
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
