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

// renderMarkdownIfNeeded applies markdown rendering to AI responses
func (m Model) RenderMarkdownIfNeeded(content string, width int) string {
	// For AI responses, always try markdown rendering to ensure proper formatting
	// The glamour library handles plain text gracefully, so this is safe
	rendered, err := m.renderMarkdown(content, width)
	if err != nil {
		// If markdown rendering fails, return original content
		return content
	}

	return rendered
}
