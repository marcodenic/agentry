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

	// Define gradient colors - balanced magenta to purple to blue to cyan (mix of both approaches)
	colors := []string{
		"#E844FF", // Bright magenta-purple (less pink)
		"#D842E8", // Purple-magenta
		"#C840D1", // Purple
		"#B83EBA", // Deep purple
		"#A83CA3", // Purple-blue
		"#983A8C", // Blue-purple
		"#883875", // Dark purple-blue
		"#78365E", // Blue-purple transition
		"#683447", // Deep blue-purple
		"#583230", // Blue
		"#483019", // Deep blue
		"#2A4A8F", // Medium blue
		"#1C5AA5", // Bright blue
		"#0E6ABB", // Sky blue
		"#007AD1", // Cyan-blue
		"#008AE7", // Bright cyan
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

		// Apply special styling for specific content
		if strings.Contains(line, "v0.2.0") {
			// Version number in italics with cyan color
			style := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00AAFF")).
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
	// Test the balanced logo gradient function
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

	fmt.Println("🎨 Testing Balanced Gradient + Special Text Styling:")
	fmt.Println("=" + strings.Repeat("=", 75))

	gradientLogo := applyGradientToLogo(rawLogo)
	fmt.Print(gradientLogo)

	fmt.Println("\n" + strings.Repeat("=", 77))
	fmt.Println("✨ Balanced gradient with bold framework text and italic version!")
}
