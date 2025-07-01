package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// applyGradientToLogo applies a beautiful gradient effect to the ASCII logo
func applyGradientToLogo(logo string) string {
	lines := strings.Split(logo, "\n")
	var styledLines []string

	// Vibrant neon-like gradient: Electric magenta → Purple → Blue → Cyan
	// Inspired by synthwave/cyberpunk aesthetics for a stunning modern look
	colors := []string{
		"#FF1493", // Deep hot pink (top)
		"#FF1493", // Deep hot pink
		"#FF00FF", // Bright magenta/fuchsia
		"#E040FB", // Electric purple
		"#D500F9", // Bright purple
		"#C51162", // Dark pink-purple
		"#AA00FF", // Electric violet
		"#9C27B0", // Material purple
		"#8E24AA", // Deep purple
		"#7B1FA2", // Royal purple
		"#673AB7", // Deep violet
		"#5E35B1", // Dark purple-blue
		"#512DA8", // Royal purple-blue
		"#4527A0", // Deep blue-purple
		"#3F51B5", // Indigo
		"#3949AB", // Bright indigo
		"#303F9F", // Deep indigo
		"#283593", // Dark royal blue
		"#1976D2", // Material blue
		"#1565C0", // Bright blue
		"#0D47A1", // Deep blue
		"#01579B", // Dark blue
		"#0277BD", // Light blue
		"#0288D1", // Sky blue
		"#039BE5", // Bright sky blue
		"#03A9F4", // Light blue
		"#00BCD4", // Cyan
		"#00ACC1", // Dark cyan
		"#0097A7", // Teal-cyan (bottom)
	}

	// Count non-empty lines to properly distribute gradient (excluding padding)
	nonEmptyLines := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmptyLines++
		}
	}

	currentNonEmptyIndex := 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			styledLines = append(styledLines, line)
			continue
		}

		// Calculate color based on non-empty line position for better gradient distribution
		colorIndex := (currentNonEmptyIndex * len(colors)) / nonEmptyLines
		if colorIndex >= len(colors) {
			colorIndex = len(colors) - 1
		}
		currentNonEmptyIndex++

		// Apply special styling for specific content
		if strings.Contains(line, "v0.2.0") {
			// Version number in italics with bright cyan
			style := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00E5FF")).
				Italic(true)
			styledLines = append(styledLines, style.Render(line))
		} else if strings.Contains(line, "AGENT") && strings.Contains(line, "ORCHESTRATION") && strings.Contains(line, "FRAMEWORK") {
			// Framework title in bold with current gradient color
			style := lipgloss.NewStyle().
				Foreground(lipgloss.Color(colors[colorIndex])).
				Bold(true)
			styledLines = append(styledLines, style.Render(line))
		} else {
			// Apply the gradient color to the line
			style := lipgloss.NewStyle().
				Foreground(lipgloss.Color(colors[colorIndex]))
			styledLines = append(styledLines, style.Render(line))
		}
	}

	return strings.Join(styledLines, "\n")
}

func main() {
	rawLogo := `

     ██████╗  ██████╗ ███████╗███╗   ██╗████████╗██████╗ ██╗   ██╗
    ██╔══██╗██╔════╝ ██╔════╝████╗  ██║╚══██╔══╝██╔══██╗╚██╗ ██╔╝
    ███████║██║  ███╗█████╗  ██╔██╗ ██║   ██║   ██████╔╝ ╚████╔╝ 
    ██╔══██║██║   ██║██╔══╝  ██║╚██╗██║   ██║   ██╔══██╗  ╚██╔╝  
    ██║  ██║╚██████╔╝███████╗██║ ╚████║   ██║   ██║  ██║   ██║   
    ╚═╝  ╚═╝ ╚═════╝ ╚══════╝╚═╝  ╚═══╝   ╚═╝   ╚═╝  ╚═╝   ╚═╝   
                                                                  
           AGENT ORCHESTRATION FRAMEWORK v0.2.0

`

	styledLogo := applyGradientToLogo(rawLogo)
	fmt.Print(styledLogo)
}
