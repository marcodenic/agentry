package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) formatWithBar(bar, text string, width int) string {
	if text == "" {
		return bar + " "
	}
	cleanText := strings.ReplaceAll(text, "┃", "")
	cleanText = strings.Trim(cleanText, " \t")
	if cleanText == "" {
		return bar + " "
	}
	
	barWidth := lipgloss.Width(bar) + 1
	textWidth := width - barWidth
	if textWidth <= 20 {
		textWidth = 60
	}
	
	// Apply markdown rendering for AI responses
	aiBar := m.aiBar()
	if bar == aiBar {
		// This is an AI response - apply markdown rendering if detected
		renderedText := m.RenderMarkdownIfNeeded(cleanText, textWidth)
		if renderedText != cleanText {
			// Markdown was applied - the text is already formatted and wrapped by glamour
			// Just add the bar to each line without additional word wrapping
			lines := strings.Split(renderedText, "\n")
			var result strings.Builder
			first := true
			for _, line := range lines {
				if !first {
					result.WriteString("\n")
				}
				first = false
				result.WriteString(bar + " " + line)
			}
			return result.String()
		}
		// If markdown wasn't applied, continue with normal text wrapping below
	}
	
	lines := strings.Split(cleanText, "\n")
	var result strings.Builder
	first := true
	for _, line := range lines {
		if !first {
			result.WriteString("\n")
		}
		first = false
		if len(line) <= textWidth {
			result.WriteString(bar + " " + line)
		} else {
			words := strings.Fields(line)
			var currentLine strings.Builder
			lineFirst := true
			for _, word := range words {
				testLine := currentLine.String()
				if testLine != "" {
					testLine += " "
				}
				testLine += word
				if len(testLine) <= textWidth {
					if currentLine.Len() > 0 {
						currentLine.WriteString(" ")
					}
					currentLine.WriteString(word)
				} else {
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
		if strings.TrimSpace(line) == "" {
			continue
		}
		userBar := m.userBar()
		aiBar := m.aiBar()
		if strings.HasPrefix(line, userBar+" ") {
			content := strings.TrimPrefix(line, userBar+" ")
			formatted := m.formatWithBar(userBar, content, width)
			result.WriteString(formatted)
		} else if strings.HasPrefix(line, aiBar+" ") {
			content := strings.TrimPrefix(line, aiBar+" ")
			formatted := m.formatWithBar(aiBar, content, width)
			result.WriteString(formatted)
		} else {
			result.WriteString(line)
		}
	}
	return result.String()
}

func (m Model) formatSingleCommand(command string) string {
	return fmt.Sprintf("%s %s", m.statusBar(), command)
}

func (m Model) formatUserInput(bar, text string, width int) string {
	if text == "" {
		return bar + " "
	}
	cleanText := strings.ReplaceAll(text, "┃", "")
	cleanText = strings.Trim(cleanText, " \t")
	if cleanText == "" {
		return bar + " "
	}
	barWidth := lipgloss.Width(bar) + 1
	textWidth := width - barWidth
	if textWidth <= 10 {
		textWidth = 40
	}
	words := strings.Fields(cleanText)
	if len(words) == 0 {
		return bar + " "
	}
	var lines []string
	var currentLine strings.Builder
	for _, word := range words {
		testLine := currentLine.String()
		if testLine != "" {
			testLine += " "
		}
		testLine += word
		if len(testLine) <= textWidth {
			if currentLine.Len() > 0 {
				currentLine.WriteString(" ")
			}
			currentLine.WriteString(word)
		} else {
			if currentLine.Len() > 0 {
				lines = append(lines, currentLine.String())
				currentLine.Reset()
			}
			currentLine.WriteString(word)
		}
	}
	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}
	if len(lines) == 0 {
		return bar + " "
	}
	var result strings.Builder
	for i, line := range lines {
		if i > 0 {
			result.WriteString("\n")
		}
		result.WriteString(bar + " " + line)
	}
	return result.String()
}
