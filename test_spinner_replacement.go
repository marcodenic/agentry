package main

import (
	"fmt"
	"strings"
)

func main() {
	// Test the logic for spinner replacement
	
	// Test case 1: Initial AI bar with space
	history := "┃ "
	fmt.Printf("Initial: %q\n", history)
	
	// Simulate first thinking animation - should replace space with spinner
	if len(history) > 0 {
		lastChar := history[len(history)-1:]
		if lastChar == " " && strings.HasSuffix(history, " ") {
			if len(history) >= 2 && history[len(history)-2] != ' ' {
				history = history[:len(history)-1] // Remove the space
			}
		}
	}
	history += "|"
	fmt.Printf("After first spinner: %q\n", history)
	
	// Simulate spinner animation - should replace spinner with new frame
	if len(history) > 0 {
		lastChar := history[len(history)-1:]
		if lastChar == "|" || lastChar == "/" || lastChar == "-" || lastChar == "\\" {
			history = history[:len(history)-1]
		}
	}
	history += "/"
	fmt.Printf("After spinner animation: %q\n", history)
	
	// Simulate first token - should replace spinner with token
	if len(history) > 0 {
		lastChar := history[len(history)-1:]
		if lastChar == "|" || lastChar == "/" || lastChar == "-" || lastChar == "\\" {
			history = history[:len(history)-1]
		}
	}
	history += "Hello"
	fmt.Printf("After first token: %q\n", history)
	
	// Simulate second token
	history += " world"
	fmt.Printf("After second token: %q\n", history)
	
	// Expected final result should be: "┃Hello world"
	expected := "┃Hello world"
	if history == expected {
		fmt.Println("✅ Test passed! Spinner replacement works correctly.")
	} else {
		fmt.Printf("❌ Test failed! Expected %q, got %q\n", expected, history)
	}
}
