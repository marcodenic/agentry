package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func applyGradientToLogo(logo string) string {
	lines := strings.Split(logo, "\n")
	var styledLines []string

	colors := []string{
		"#8B5FBF",
		"#8B5FBF",
		"#6B76CF",
		"#5B82D7",
		"#4B8EDF",
		"#3B9AE7",
		"#2BA6EF",
		"#1BB2F7",
		"#0BBEFF",
		"#00CAF7",
		"#00D6EF",
		"#00E2E7",
	}

	totalLines := len(lines)
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			styledLines = append(styledLines, line)
			continue
		}
		colorIndex := (i * len(colors)) / totalLines
		if colorIndex >= len(colors) {
			colorIndex = len(colors) - 1
		}
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(colors[colorIndex]))
		styledLines = append(styledLines, style.Render(line))
	}

	return strings.Join(styledLines, "\n")
}
