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

	// Define gradient colors - purple to blue to green to pink
	colors := []string{
		"#9945FF", // Purple
		"#8B5CF6", // Violet
		"#7C3AED", // Purple-blue
		"#6366F1", // Indigo
		"#3B82F6", // Blue
		"#06B6D4", // Cyan
		"#10B981", // Emerald
		"#84CC16", // Lime
		"#EAB308", // Yellow
		"#F59E0B", // Amber
		"#EF4444", // Red
		"#EC4899", // Pink
		"#D946EF", // Fuchsia
		"#9945FF", // Back to purple
	}

	totalLines := len(lines)

	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			styledLines = append(styledLines, line)
			continue
		}

		// Calculate which color to use based on line position
		colorIndex := (i * len(colors)) / totalLines
		if colorIndex >= len(colors) {
			colorIndex = len(colors) - 1
		}

		// Apply the color to the line
		style := lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors[colorIndex])).
			Bold(true)

		styledLines = append(styledLines, style.Render(line))
	}

	return strings.Join(styledLines, "\n")
}

func main() {
	// Test the logo gradient function
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

	fmt.Println("🎨 Testing Gradient Logo Effect:")
	fmt.Println("=" + strings.Repeat("=", 60))

	gradientLogo := applyGradientToLogo(rawLogo)
	fmt.Print(gradientLogo)

	fmt.Println("\n" + strings.Repeat("=", 62))
	fmt.Println("✨ Gradient effect applied successfully!")
}
