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

	// Define gradient colors - beautiful purple to magenta to blue (matching the stunning image)
	colors := []string{
		"#8A2BE2", // Blue violet (top)
		"#9932CC", // Dark orchid
		"#9F44D3", // Purple
		"#A855F7", // Purple
		"#B566FF", // Light purple
		"#C277FF", // Purple-magenta
		"#CF88FF", // Magenta-purple
		"#DC99FF", // Light magenta
		"#E9AAFF", // Pink-magenta
		"#F6BBFF", // Light pink-magenta
		"#FF88DD", // Pink
		"#FF77CC", // Bright pink
		"#FF66BB", // Pink-blue
		"#FF55AA", // Pink-cyan
		"#4A90E2", // Beautiful blue
		"#00BFFF", // Deep sky blue (bottom)
	}

	totalLines := len(lines)

	// Count non-empty lines to properly distribute gradient (excluding padding)
	nonEmptyLines := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmptyLines++
		}
	}

	currentNonEmptyIndex := 0
	for i, line := range lines {
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
			// Version number in italics with cyan color
			style := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00BFFF")).
				Italic(true)
			styledLines = append(styledLines, style.Render(line))
		} else if strings.Contains(line, "AGENT") && strings.Contains(line, "ORCHESTRATION") && strings.Contains(line, "FRAMEWORK") {
			// Framework title in bold with gradient color
			style := lipgloss.NewStyle().
				Foreground(lipgloss.Color(colors[colorIndex])).
				Bold(true)
			styledLines = append(styledLines, style.Render(line))
		} else {
			// Apply the color to the line with standard styling
			style := lipgloss.NewStyle().
				Foreground(lipgloss.Color(colors[colorIndex]))
			styledLines = append(styledLines, style.Render(line))
		}
	}

	return strings.Join(styledLines, "\n")
}

func main() {
	// Test the beautiful logo gradient function
	rawLogo := `
                                                             
                                                             
                  ████▒               ▒████                  
                    ▒▓███▓▒       ▒▓███▓▒                    
                      ▒█▒████▓▒▓████▓█▒                      
                      ▒█   ▓█████▓▒  █▒                      
                      ▒█▓███▓▓█▓▓███▓█▒                      
                   ▒▓███▓▒   ▒▓▒   ▒▓███▓▒                   
                 ▒███▓▓█     ▒▓▒     █▓▓▓██▒                 
                      ▒█     ▒▓▒     █▒                      
                      ▒█     ▒▓▒     █▒                      
                      ▒█     ▒▓▒     █▒                      
                      ▒█     ▒▓▒     █▒                      
                      ▒█     ▒▓▒     █▒                      
                      ▒█     ▒▓▒     █▒                      
                             ▒▓▒                             
                                                             
                                      v0.2.0                 
                 █▀█ █▀▀ █▀▀ █▀█ ▀█▀ █▀▄ █ █                 
                 █▀█ █ █ █▀▀ █ █  █  █▀▄  █                  
                 ▀ ▀ ▀▀▀ ▀▀▀ ▀ ▀  ▀  ▀ ▀  ▀                  
               AGENT  ORCHESTRATION  FRAMEWORK               
                                                             `

	fmt.Println("💜 Testing Beautiful Purple-Pink-Blue Gradient:")
	fmt.Println("=" + strings.Repeat("=", 75))

	gradientLogo := applyGradientToLogo(rawLogo)
	fmt.Print(gradientLogo)

	fmt.Println("\n" + strings.Repeat("=", 77))
	fmt.Println("✨ Beautiful gradient with proper padding handling!")
}
