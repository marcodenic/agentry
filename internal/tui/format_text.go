package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// getBarSpacing returns the appropriate spacing for a given bar type
func (m Model) getBarSpacing(bar string) (string, int) {
	aiBar := m.aiBar()
	if bar == aiBar {
		return "  ", 2 // AI responses use 2 spaces (accounts for Glamour padding)
	}
	return "    ", 4 // User inputs and status messages use 4 spaces
}

// calculateTextWidth calculates the available text width given total width and bar
func (m Model) calculateTextWidth(totalWidth int, bar string, minWidth int) int {
	_, spacingWidth := m.getBarSpacing(bar)
	barWidth := lipgloss.Width(bar) + spacingWidth
	textWidth := totalWidth - barWidth

	// Use fallback width if calculation results in too narrow text area
	if textWidth <= minWidth {
		return 72 // Fallback to reasonable width (80% of default viewport width)
	}
	return textWidth
}

// wrapTextToLines wraps text to fit within the specified width
func wrapTextToLines(text string, maxWidth int) []string {
	if text == "" {
		return nil
	}

	lines := strings.Split(text, "\n")
	var result []string

	for _, line := range lines {
		if len(line) <= maxWidth {
			result = append(result, line)
		} else {
			// Wrap long lines
			words := strings.Fields(line)
			var currentLine strings.Builder

			for _, word := range words {
				testLine := currentLine.String()
				if testLine != "" {
					testLine += " "
				}
				testLine += word

				if len(testLine) <= maxWidth {
					if currentLine.Len() > 0 {
						currentLine.WriteString(" ")
					}
					currentLine.WriteString(word)
				} else {
					if currentLine.Len() > 0 {
						result = append(result, currentLine.String())
						currentLine.Reset()
					}
					currentLine.WriteString(word)
				}
			}

			if currentLine.Len() > 0 {
				result = append(result, currentLine.String())
			}
		}
	}

	return result
}

func (m Model) formatWithBar(bar, text string, width int) string {
	spacing, _ := m.getBarSpacing(bar)

	if text == "" {
		return bar + spacing
	}

	cleanText := strings.ReplaceAll(text, "┃", "")
	cleanText = strings.Trim(cleanText, " \t")
	if cleanText == "" {
		return bar + spacing
	}

	textWidth := m.calculateTextWidth(width, bar, 20)
	aiBar := m.aiBar()

	// Apply markdown rendering for AI responses
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
				result.WriteString(bar + spacing + line)
			}
			return result.String()
		}
		// If markdown wasn't applied, continue with normal text wrapping below
	}

	// Use the consolidated wrapping function
	lines := wrapTextToLines(cleanText, textWidth)
	var result strings.Builder
	first := true
	for _, line := range lines {
		if !first {
			result.WriteString("\n")
		}
		first = false
		result.WriteString(bar + spacing + line)
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
	spacing, _ := m.getBarSpacing(m.statusBar())
	return fmt.Sprintf("%s%s%s", m.statusBar(), spacing, command)
}

func (m Model) formatUserInput(bar, text string, width int) string {
	spacing, _ := m.getBarSpacing(bar)

	if text == "" {
		return bar + spacing
	}

	cleanText := strings.ReplaceAll(text, "┃", "")
	cleanText = strings.Trim(cleanText, " \t")
	if cleanText == "" {
		return bar + spacing
	}

	textWidth := m.calculateTextWidth(width, bar, 10)
	lines := wrapTextToLines(cleanText, textWidth)

	if len(lines) == 0 {
		return bar + spacing
	}

	var result strings.Builder
	for i, line := range lines {
		if i > 0 {
			result.WriteString("\n")
		}
		result.WriteString(bar + spacing + line)
	}
	return result.String()
}
