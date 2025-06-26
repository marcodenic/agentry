package main

import (
	"fmt"
	"strings"
)

func mainDebug() {
	// Simulate the spinner logic
	history := "┃ user input\n┃ "
	
	// Add spinner
	history += "|"
	fmt.Printf("After adding spinner: %q\n", history)
	
	// Simulate first token arrival - remove spinner
	if strings.HasSuffix(history, "|") {
		history = history[:len(history)-1]
	}
	fmt.Printf("After removing spinner: %q\n", history)
	
	// Add first token
	history += "Hello"
	fmt.Printf("After adding first token: %q\n", history)
	
	// Add more tokens
	history += " world"
	fmt.Printf("After adding more tokens: %q\n", history)
	
	// Final result should show user input on one line, AI response on same line as AI bar
	fmt.Println("\nExpected result:")
	fmt.Println("┃ user input")
	fmt.Println("┃ Hello world")
}
