package main

import (
	"fmt"
	"strings"
)

func main() {
	// Test the exact scenario from the user's description
	
	// Simulate agent response without proper newline
	history := "┃ hello\n┃good evening, how are you, are you well?"
	fmt.Printf("History before newline check: %q\n", history)
	
	// Check if we need to add newline (this is what agentCompleteMsg should do)
	if len(history) > 0 && !strings.HasSuffix(history, "\n") {
		history += "\n"
	}
	fmt.Printf("History after newline check: %q\n", history)
	
	// Simulate next user input
	nextInput := "yes, I'm doing well"
	history += "┃ " + nextInput + "\n"
	history += "┃ " // AI bar for next response
	
	fmt.Printf("History after next input: %q\n", history)
	
	// This should show proper line separation
	fmt.Println("Visual representation:")
	fmt.Print(history)
}
