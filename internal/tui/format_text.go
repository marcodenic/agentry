package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) formatWithBar(bar, text string, width int) string {
	if text == "" {
		return bar + "  " // Use 2 spaces for AI responses
	}
	cleanText := strings.ReplaceAll(text, "┃", "")
	cleanText = strings.Trim(cleanText, " \t")
	if cleanText == "" {
		return bar + "  " // Use 2 spaces for AI responses
	}
	
	barWidth := lipgloss.Width(bar) + 2  // Account for the "  " spacing we add
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
			// Markdown was applied - clean up glamour's extra padding and spacing
			lines := strings.Split(renderedText, "\n")
			var cleanedLines []string
			
			for _, line := range lines {
				// Remove leading whitespace but preserve ANSI formatting
				trimmed := strings.TrimLeft(line, " \t\u00A0\u2002\u2003\u2004\u2005\u2006\u2007\u2008\u2009\u200A\u202F\u205F\u3000")
				cleanedLines = append(cleanedLines, trimmed)
			}
			
			// Remove empty lines at the beginning and end
			for len(cleanedLines) > 0 && strings.TrimSpace(cleanedLines[0]) == "" {
				cleanedLines = cleanedLines[1:]
			}
			for len(cleanedLines) > 0 && strings.TrimSpace(cleanedLines[len(cleanedLines)-1]) == "" {
				cleanedLines = cleanedLines[:len(cleanedLines)-1]
			}
			
			// Add the bar to each line with consistent spacing
			var result strings.Builder
			first := true
			for _, line := range cleanedLines {
				if !first {
					result.WriteString("\n")
				}
				first = false
				result.WriteString(bar + "  " + line) // Use 2 spaces for markdown rendered content
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
			result.WriteString(bar + "  " + line) // Use 2 spaces for non-markdown content
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
						result.WriteString(bar + "  " + currentLine.String()) // Use 2 spaces for consistency
						currentLine.Reset()
					}
					currentLine.WriteString(word)
				}
			}
			if currentLine.Len() > 0 {
				if !lineFirst {
					result.WriteString("\n")
				}
				result.WriteString(bar + "  " + currentLine.String()) // Use 2 spaces for consistency
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
		return bar + "    " // Use 4 spaces to match AI spacing
	}
	cleanText := strings.ReplaceAll(text, "┃", "")
	cleanText = strings.Trim(cleanText, " \t")
	if cleanText == "" {
		return bar + "    " // Use 4 spaces to match AI spacing
	}
	barWidth := lipgloss.Width(bar) + 4  // Account for more spacing to match AI output
	textWidth := width - barWidth
	if textWidth <= 10 {
		textWidth = 40
	}
	words := strings.Fields(cleanText)
	if len(words) == 0 {
		return bar + "    " // Use 4 spaces to match AI spacing
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
		return bar + "    " // Use 4 spaces to match AI spacing
	}
	var result strings.Builder
	for i, line := range lines {
		if i > 0 {
			result.WriteString("\n")
		}
		// Add consistent left padding to match AI responses 
		result.WriteString(bar + "    " + line) // Use "    " (4 spaces) to match AI indentation
	}
	return result.String()
}
