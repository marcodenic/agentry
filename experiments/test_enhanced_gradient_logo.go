package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// applyEnhancedGradientToLogo applies a more sophisticated gradient effect with character-level coloring
func applyEnhancedGradientToLogo(logo string) string {
	lines := strings.Split(logo, "\n")
	var styledLines []string

	// Define gradient colors - matching the image's purple-blue-green-pink style
	colors := []string{
		"#9945FF", // Bright Purple
		"#8B5CF6", // Violet
		"#7C3AED", // Purple-blue
		"#6366F1", // Indigo
		"#3B82F6", // Blue
		"#06B6D4", // Cyan
		"#10B981", // Emerald
		"#84CC16", // Lime
		"#F59E0B", // Amber
		"#EF4444", // Red
		"#EC4899", // Pink
		"#D946EF", // Fuchsia
	}

	totalLines := len(lines)

	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			styledLines = append(styledLines, line)
			continue
		}

		// Apply different gradient strategies based on content
		if strings.Contains(line, "████") || strings.Contains(line, "█") {
			// Logo art gets character-by-character gradient
			styledLines = append(styledLines, applyCharacterGradient(line, colors, i, totalLines))
		} else if strings.Contains(line, "v0.2.0") {
			// Version gets special highlighting
			style := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#06B6D4")).
				Bold(true).
				Italic(true)
			styledLines = append(styledLines, style.Render(line))
		} else if strings.Contains(line, "AGENT") || strings.Contains(line, "ORCHESTRATION") || strings.Contains(line, "FRAMEWORK") {
			// Title gets rainbow effect
			styledLines = append(styledLines, applyRainbowEffect(line, colors))
		} else {
			// Regular line-based gradient
			colorIndex := (i * len(colors)) / totalLines
			if colorIndex >= len(colors) {
				colorIndex = len(colors) - 1
			}

			style := lipgloss.NewStyle().
				Foreground(lipgloss.Color(colors[colorIndex])).
				Bold(true)

			styledLines = append(styledLines, style.Render(line))
		}
	}

	return strings.Join(styledLines, "\n")
}

// applyCharacterGradient applies gradient to each character individually
func applyCharacterGradient(line string, colors []string, lineIndex, totalLines int) string {
	if strings.TrimSpace(line) == "" {
		return line
	}

	runes := []rune(line)
	var styledChars []string

	for j, r := range runes {
		if r == ' ' {
			styledChars = append(styledChars, " ")
			continue
		}

		// Calculate color based on both line position and character position
		colorProgress := float64(lineIndex)/float64(totalLines) + float64(j)/float64(len(runes)*4)
		colorIndex := int(colorProgress*float64(len(colors))) % len(colors)

		style := lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors[colorIndex])).
			Bold(true)

		styledChars = append(styledChars, style.Render(string(r)))
	}

	return strings.Join(styledChars, "")
}

// applyRainbowEffect applies rainbow colors to words
func applyRainbowEffect(line string, colors []string) string {
	words := strings.Fields(line)
	var styledWords []string

	for i, word := range words {
		colorIndex := i % len(colors)
		style := lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors[colorIndex])).
			Bold(true)
		styledWords = append(styledWords, style.Render(word))
	}

	// Preserve original spacing
	result := strings.Join(styledWords, " ")

	// Add leading spaces to maintain alignment
	originalLeading := len(line) - len(strings.TrimLeft(line, " "))
	return strings.Repeat(" ", originalLeading) + strings.TrimLeft(result, " ")
}

func main() {
	// Test the enhanced logo gradient function
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

	fmt.Println("🌈 Testing Enhanced Gradient Logo Effect:")
	fmt.Println("=" + strings.Repeat("=", 70))

	gradientLogo := applyEnhancedGradientToLogo(rawLogo)
	fmt.Print(gradientLogo)

	fmt.Println("\n" + strings.Repeat("=", 72))
	fmt.Println("✨ Enhanced gradient effect with character-level coloring applied!")
}
