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

	// Define gradient colors - vibrant magenta to purple to blue to cyan (matching the new image)
	colors := []string{
		"#FF44FF", // Bright neon magenta
		"#F542F5", // Magenta
		"#EB40EB", // Pink-magenta
		"#E13EE1", // Purple-pink
		"#D73CD7", // Purple-magenta
		"#CD3ACD", // Purple
		"#C338C3", // Deep purple
		"#B936B9", // Purple-blue
		"#AF34AF", // Blue-purple
		"#A532A5", // Purple-blue
		"#9B309B", // Blue-purple
		"#912E91", // Blue
		"#872C87", // Deep blue-purple
		"#7D2A7D", // Blue
		"#732873", // Blue-cyan
		"#44AAFF", // Bright cyan-blue
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
	// Test the vibrant logo gradient function
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

	fmt.Println("🎨 Testing Vibrant Gradient (Magenta → Purple → Blue → Cyan):")
	fmt.Println("=" + strings.Repeat("=", 70))

	gradientLogo := applyGradientToLogo(rawLogo)
	fmt.Print(gradientLogo)

	fmt.Println("\n" + strings.Repeat("=", 72))
	fmt.Println("✨ Vibrant neon-style gradient applied - matches the icon aesthetic!")
}
