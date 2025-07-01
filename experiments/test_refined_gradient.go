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

	// Define gradient colors - subtle purple to blue to teal (matching the image style)
	colors := []string{
		"#8B5FBF", // Soft purple
		"#7B6BC7", // Purple-blue
		"#6B76CF", // Lavender blue
		"#5B82D7", // Medium blue
		"#4B8EDF", // Light blue
		"#3B9AE7", // Sky blue
		"#2BA6EF", // Bright blue
		"#1BB2F7", // Cyan blue
		"#0BBEFF", // Light cyan
		"#00CAF7", // Teal cyan
		"#00D6EF", // Soft teal
		"#00E2E7", // Light teal
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

		// Apply the color to the line with subtle styling
		style := lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors[colorIndex]))

		styledLines = append(styledLines, style.Render(line))
	}

	return strings.Join(styledLines, "\n")
}

func main() {
	// Test the refined logo gradient function
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

	fmt.Println("🎨 Testing Refined Gradient (Purple → Blue → Teal):")
	fmt.Println("=" + strings.Repeat("=", 65))

	gradientLogo := applyGradientToLogo(rawLogo)
	fmt.Print(gradientLogo)

	fmt.Println("\n" + strings.Repeat("=", 67))
	fmt.Println("✨ Subtle gradient effect applied - matches the image aesthetic!")
}
