package tui

import (
	"bytes"
	"fmt"
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

// thinkingBar returns the glyph used to prefix thinking messages.
func (m Model) thinkingBar() string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.AIBarColor)).Render("🤔")
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

// renderMemory formats an agent's memory history and detailed trace for display.
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
		// Step header with enhanced styling
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
		
		// Agent reasoning/output section
		if s.Output != "" {
			b.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF88")).
				Bold(true).
				Render("� AGENT REASONING & ANALYSIS:"))
			b.WriteString("\n")
			
			// Format the output with better presentation
			outputBox := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#1a1a1a")).
				Padding(1, 2).
				Margin(0, 0, 1, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#444444"))
			
			// Clean and format the output text
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
		
		// Tool execution section
		if len(s.ToolCalls) > 0 {
			b.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF8844")).
				Bold(true).
				Render("🔧 TOOL EXECUTIONS:"))
			b.WriteString("\n\n")
			
			for j, tc := range s.ToolCalls {
				// Tool call header
				toolHeader := fmt.Sprintf("▶ Tool #%d: %s", j+1, tc.Name)
				b.WriteString(lipgloss.NewStyle().
					Foreground(lipgloss.Color("#FF6666")).
					Bold(true).
					Render(toolHeader))
				b.WriteString("\n")
				
				// Tool arguments section
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
				
				// Tool result section
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
					
					// Format result with proper handling of long text
					cleanResult := strings.TrimSpace(result)
					if len(cleanResult) > 500 {
						// Truncate very long results but show key information
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
		
		// Add visual separator between steps
		if i < len(hist)-1 {
			b.WriteString("\n")
			b.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color("#666666")).
				Render(strings.Repeat("═", 80)))
			b.WriteString("\n\n")
		}
	}
	
	// Add footer with summary information
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Italic(true).
		Render(fmt.Sprintf("📊 Summary: %d execution steps completed", len(hist))))
	
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

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// formatWithBar formats text with a vertical bar on the first line and continuation bars on wrapped lines
// This version preserves original formatting including newlines from AI responses
func (m Model) formatWithBar(bar, text string, width int) string {
	if text == "" {
		return bar + " "
	}

	// Clean the text - remove any existing vertical bars but preserve newlines
	cleanText := strings.ReplaceAll(text, "┃", "")
	// Only trim leading/trailing spaces, NOT newlines to preserve formatting
	cleanText = strings.Trim(cleanText, " \t")

	if cleanText == "" {
		return bar + " "
	}

	// Calculate available width for text (subtract bar + space)
	barWidth := lipgloss.Width(bar) + 1 // +1 for space after bar
	textWidth := width - barWidth
	if textWidth <= 20 { // Minimum reasonable width
		textWidth = 60 // Default fallback width
	}

	// Split by existing newlines to preserve AI formatting, then wrap each line
	lines := strings.Split(cleanText, "\n")
	
	var result strings.Builder
	first := true
	
	for _, line := range lines {
		if !first {
			result.WriteString("\n")
		}
		first = false
		
		// Wrap this line if it's too long
		if len(line) <= textWidth {
			result.WriteString(bar + " " + line)
		} else {
			// Word wrap this line
			words := strings.Fields(line)
			var currentLine strings.Builder
			lineFirst := true
			
			for _, word := range words {
				// Check if adding this word would exceed the line width
				testLine := currentLine.String()
				if testLine != "" {
					testLine += " "
				}
				testLine += word

				if len(testLine) <= textWidth {
					// Word fits on current line
					if currentLine.Len() > 0 {
						currentLine.WriteString(" ")
					}
					currentLine.WriteString(word)
				} else {
					// Word doesn't fit, wrap to new line
					if currentLine.Len() > 0 {
						if !lineFirst {
							result.WriteString("\n")
						}
						lineFirst = false
						result.WriteString(bar + " " + currentLine.String())
						currentLine.Reset()
					}
					currentLine.WriteString(word)
				}
			}

			// Add remaining content
			if currentLine.Len() > 0 {
				if !lineFirst {
					result.WriteString("\n")
				}
				result.WriteString(bar + " " + currentLine.String())
			}
		}
	}

	return result.String()
}

// formatHistoryWithBars reformats the entire chat history to apply proper line wrapping with vertical bars
func (m Model) formatHistoryWithBars(history string, width int) string {
	if history == "" {
		return ""
	}

	lines := strings.Split(history, "\n")
	var result strings.Builder

	for i, line := range lines {
		if i > 0 {
			result.WriteString("\n")
		}

		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Check if line starts with a user bar
		userBar := m.userBar()
		aiBar := m.aiBar()

		if strings.HasPrefix(line, userBar+" ") {
			// User message - reformat with proper wrapping
			content := strings.TrimPrefix(line, userBar+" ")
			formatted := m.formatWithBar(userBar, content, width)
			result.WriteString(formatted)
		} else if strings.HasPrefix(line, aiBar+" ") {
			// AI message - reformat with proper wrapping
			content := strings.TrimPrefix(line, aiBar+" ")
			formatted := m.formatWithBar(aiBar, content, width)
			result.WriteString(formatted)
		} else {
			// Other content (status messages, etc.) - pass through as-is
			result.WriteString(line)
		}
	}

	return result.String()
}

// formatCommandGroup wraps related commands with proper spacing and visual grouping
func (m Model) formatCommandGroup(commands []string) string {
	if len(commands) == 0 {
		return ""
	}

	// Create a visual separator for command groups
	separator := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.AIBarColor)).
		Render("─────────────────────────────────────────────────")

	var formatted strings.Builder
	formatted.WriteString("\n")
	formatted.WriteString(separator)
	formatted.WriteString("\n")

	for i, cmd := range commands {
		formatted.WriteString(fmt.Sprintf("%s %s", m.statusBar(), cmd))
		if i < len(commands)-1 {
			formatted.WriteString("\n")
		}
	}

	formatted.WriteString("\n")
	formatted.WriteString(separator)
	formatted.WriteString("\n\n")

	return formatted.String()
}

// formatSingleCommand formats a single command with proper spacing
func (m Model) formatSingleCommand(command string) string {
	return fmt.Sprintf("\n%s %s\n", m.statusBar(), command)
}

// formatUserInput formats user input with proper word wrapping
func (m Model) formatUserInput(bar, text string, width int) string {
	if text == "" {
		return bar + " "
	}

	// Clean the text - remove any existing vertical bars  
	cleanText := strings.ReplaceAll(text, "┃", "")
	// Only trim leading/trailing spaces, NOT newlines to preserve formatting
	cleanText = strings.Trim(cleanText, " \t")

	if cleanText == "" {
		return bar + " "
	}

	// Calculate available width for text (subtract bar + space)
	barWidth := lipgloss.Width(bar) + 1 // +1 for space after bar
	textWidth := width - barWidth
	if textWidth <= 10 { // Minimum reasonable width
		textWidth = 40
	}

	// Split text into words for wrapping
	words := strings.Fields(cleanText)
	if len(words) == 0 {
		return bar + " "
	}

	var lines []string
	var currentLine strings.Builder

	for _, word := range words {
		// Check if adding this word would exceed the line width
		testLine := currentLine.String()
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		if len(testLine) <= textWidth {
			// Word fits on current line
			if currentLine.Len() > 0 {
				currentLine.WriteString(" ")
			}
			currentLine.WriteString(word)
		} else {
			// Word doesn't fit, start new line
			if currentLine.Len() > 0 {
				lines = append(lines, currentLine.String())
				currentLine.Reset()
			}
			currentLine.WriteString(word)
		}
	}

	// Add remaining content
	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	// Format lines with bars
	if len(lines) == 0 {
		return bar + " "
	}

	// Format each line with the bar
	var result strings.Builder
	for i, line := range lines {
		if i > 0 {
			result.WriteString("\n")
		}
		result.WriteString(bar + " " + line)
	}

	return result.String()
}

// renderDetailedMemory provides even more detailed trace information when debug data is available
func (m Model) renderDetailedMemory(ag *core.Agent) string {
	info := m.infos[ag.ID]
	if info == nil || len(info.DebugTrace) == 0 {
		// Fall back to standard memory view
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
	
	// Show current streaming response if available
	if info.DebugStreamingResponse != "" {
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("#88FF88")).
			Bold(true).
			Render("🔄 CURRENT RESPONSE BEING GENERATED:"))
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
		// Show step progression
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
		
		// Event timestamp and type
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
		
		// Skip excessive token events for readability, but count them
		if event.Type == "token" {
			// Show first few tokens individually, then aggregate
			if tokenCount <= 5 {
				// Show individual meaningful tokens
				if strings.Contains(event.Details, "Character:") {
					tokenCount++
					continue // Skip individual characters
				}
			} else if tokenCount > 5 {
				// Look ahead to see if more tokens are coming
				hasMoreTokens := i < len(info.DebugTrace)-1
				if hasMoreTokens {
					nextEvent := info.DebugTrace[i+1]
					if nextEvent.Type == "token" {
						tokenCount++
						continue // Skip this token, we'll show a summary
					}
				}
				// Show token summary when we reach the end of a token sequence
				if tokenCount > 5 {
					b.WriteString(fmt.Sprintf("  💬 %s\n", 
						lipgloss.NewStyle().Foreground(lipgloss.Color(eventColor)).Render(
							fmt.Sprintf("Response Generated [%d tokens streamed]", tokenCount))))
					tokenCount = 0 // Reset after showing summary
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
		
		// Event details
		if event.Details != "" {
			detailsBox := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#1a1a1a")).
				Padding(0, 1).
				Margin(0, 0, 0, 4).
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("#444444"))
			
			// Limit detail length for readability
			details := event.Details
			if len(details) > 200 {
				details = details[:197] + "..."
			}
			
			b.WriteString(detailsBox.Render(details))
			b.WriteString("\n")
		}
		
		b.WriteString("\n")
	}
	
	// Summary footer
	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Italic(true).
		Render(fmt.Sprintf("📊 Total events: %d | Steps completed: %d", len(info.DebugTrace), currentStep)))
	
	return b.String()
}
