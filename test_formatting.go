package main

import (
	"fmt"
	"strings"
)

// Simple test function to verify line wrapping logic
func formatWithBar(bar, text string, width int) string {
	if text == "" {
		return bar + " "
	}
	
	// Calculate available width for text (subtract bar + space)
	barWidth := len(bar) + 1 // +1 for space after bar
	textWidth := width - barWidth
	if textWidth <= 0 {
		textWidth = 1
	}
	
	// Split text into words
	words := strings.Fields(text)
	if len(words) == 0 {
		return bar + " "
	}
	
	var lines []string
	var currentLine strings.Builder
	
	for _, word := range words {
		// Check if adding this word would exceed the line width
		testLine := currentLine.String()
		if testLine != "" {
			testLine += " "
		}
		testLine += word
		
		if len(testLine) <= textWidth {
			// Word fits on current line
			if currentLine.Len() > 0 {
				currentLine.WriteString(" ")
			}
			currentLine.WriteString(word)
		} else {
			// Word doesn't fit, start new line
			if currentLine.Len() > 0 {
				lines = append(lines, currentLine.String())
				currentLine.Reset()
			}
			currentLine.WriteString(word)
		}
	}
	
	// Add remaining content
	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}
	
	// Format lines with bars
	if len(lines) == 0 {
		return bar + " "
	}
	
	// First line gets the original bar
	result := bar + " " + lines[0]
	
	// Continuation lines get the same bar
	for i := 1; i < len(lines); i++ {
		result += "\n" + bar + " " + lines[i]
	}
	
	return result
}

func main() {
	bar := "┃"
	text := "This is a very long message that should wrap to multiple lines when displayed with a narrow width to demonstrate the line continuation formatting with vertical bars."
	width := 40
	
	fmt.Println("Testing formatWithBar function:")
	fmt.Printf("Bar: %s\n", bar)
	fmt.Printf("Text: %s\n", text)
	fmt.Printf("Width: %d\n", width)
	fmt.Println("\nFormatted result:")
	result := formatWithBar(bar, text, width)
	fmt.Println(result)
	fmt.Println("\nEnd of result")
}
