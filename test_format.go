package main

import (
	"fmt"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/tui"
)

func main() {
	// Create a minimal agent
	ag := core.NewAgent("test")
	
	// Create a TUI model
	model := tui.New(ag)
	
	// Test the formatWithBar function
	bar := "┃"
	text := "This is a long message that should wrap to multiple lines when the width is limited."
	width := 40
	
	fmt.Println("Testing formatWithBar:")
	// This won't work because formatWithBar is a method on Model
	// Let me create a standalone test instead
	fmt.Printf("Input: %s\n", text)
	fmt.Printf("Width: %d\n", width)
}
