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
	// Simplified dark/light detection (feedback: previous logic redundant)
	background := "light"
	if lipgloss.HasDarkBackground() {
		background = "dark"
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

	// Pass through lines unchanged; keep potential bullet glyphs intact
	normalized = append(normalized, lines...)

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
