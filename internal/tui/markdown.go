package tui

import (
	"errors"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// renderMarkdown renders the markdown content with glamour, maintaining proper width and theme
func (m Model) renderMarkdown(content string, width int) (string, error) {
	if strings.TrimSpace(content) == "" {
		return content, nil
	}

	// Pre-process content to normalize markdown syntax
	content = normalizeMarkdownSyntax(content)

	// Calculate effective width for markdown rendering
	// Leave some margin for proper word wrapping
	markdownWidth := width
	if markdownWidth < 20 {
		markdownWidth = 60 // Fallback minimum width
	}

	// Create glamour renderer with appropriate settings
	// Use standard style with dark/light theme detection for proper formatting
	background := "dark"
	if lipgloss.HasDarkBackground() {
		background = "dark"
	} else {
		background = "light"
	}

	renderer, err := glamour.NewTermRenderer(
		glamour.WithWordWrap(markdownWidth),
		glamour.WithStandardStyle(background),
	)
	if err != nil {
		// If glamour fails, return original content
		return content, errors.New("failed to create markdown renderer")
	}

	// Render the markdown
	rendered, err := renderer.Render(content)
	if err != nil {
		// If rendering fails, return original content
		return content, errors.New("failed to render markdown")
	}

	// Clean up the rendered output
	// Remove trailing whitespace and ensure consistent line endings
	rendered = strings.TrimRightFunc(rendered, func(r rune) bool {
		return r == ' ' || r == '\t' || r == '\n' || r == '\r'
	})

	return rendered, nil
}

// normalizeMarkdownSyntax converts common alternative syntax to standard markdown
func normalizeMarkdownSyntax(content string) string {
	lines := strings.Split(content, "\n")
	var normalized []string

	for _, line := range lines {
		// Keep bullet point symbols as-is since they're already markdown-ish
		// Don't convert • to - to avoid double conversion
		// Just pass through the content
		normalized = append(normalized, line)
	}

	return strings.Join(normalized, "\n")
}

// isLikelyMarkdown performs a simple heuristic check to see if content contains markdown
func IsLikelyMarkdown(content string) bool {
	// Check for common markdown patterns
	markers := []string{
		"#",   // Headers
		"**",  // Bold
		"*",   // Italic/Bold
		"`",   // Code
		"```", // Code blocks
		"[",   // Links
		"1.",  // Numbered lists
		"-",   // Bullet lists
		"•",   // Unicode bullet points (common in AI output)
		">",   // Blockquotes
		"|",   // Tables (but not our UI bars)
	}

	content = strings.TrimSpace(content)
	if len(content) == 0 {
		return false
	}

	// Don't treat content as markdown if it only contains our UI bars
	lines := strings.Split(content, "\n")
	nonBarLines := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "┃") && !strings.HasPrefix(trimmed, "|") {
			nonBarLines++
		}
	}

	// If most lines are UI bars, don't treat as markdown
	if len(lines) > 0 && float64(nonBarLines)/float64(len(lines)) < 0.3 {
		return false
	}

	// Check for markdown markers
	for _, marker := range markers {
		if strings.Contains(content, marker) {
			// Special handling for "-", "•" and "|" to avoid false positives
			if marker == "-" {
				// Only count as markdown if it's at start of line (list item)
				if strings.Contains(content, "\n- ") || strings.HasPrefix(content, "- ") {
					return true
				}
			} else if marker == "•" {
				// Only count as markdown if it's used as a list item
				if strings.Contains(content, "\n• ") || strings.HasPrefix(content, "• ") {
					return true
				}
			} else if marker == "|" {
				// Only count as markdown if it looks like a table
				if strings.Contains(content, " | ") {
					return true
				}
			} else {
				return true
			}
		}
	}

	return false
}

// renderMarkdownIfNeeded conditionally renders markdown based on content analysis
func (m Model) RenderMarkdownIfNeeded(content string, width int) string {
	if !IsLikelyMarkdown(content) {
		return content
	}

	rendered, err := m.renderMarkdown(content, width)
	if err != nil {
		// If markdown rendering fails, return original content
		return content
	}

	return rendered
}
