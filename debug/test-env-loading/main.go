package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/marcodenic/agentry/internal/env"
)

func main() {
	fmt.Println("=== Environment Loading Debug ===")
	
	// Print current working directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting working directory: %v\n", err)
	} else {
		fmt.Printf("Working directory: %s\n", wd)
	}
	
	// Check if .env.local exists in current directory
	envPath := filepath.Join(wd, ".env.local")
	if _, err := os.Stat(envPath); err == nil {
		fmt.Printf(".env.local found at: %s\n", envPath)
	} else {
		fmt.Printf(".env.local not found at: %s (error: %v)\n", envPath, err)
	}
	
	// Print OPENAI_KEY before loading
	fmt.Printf("OPENAI_KEY before env.Load(): '%s'\n", os.Getenv("OPENAI_KEY"))
	
	// Load environment
	fmt.Println("Calling env.Load()...")
	env.Load()
	
	// Print OPENAI_KEY after loading
	fmt.Printf("OPENAI_KEY after env.Load(): '%s'\n", os.Getenv("OPENAI_KEY"))
	
	// Check if it's set to something meaningful
	openaiKey := os.Getenv("OPENAI_KEY")
	if openaiKey == "" {
		fmt.Println("❌ OPENAI_KEY is empty or not set")
	} else if len(openaiKey) < 10 {
		fmt.Println("❌ OPENAI_KEY seems too short to be valid")
	} else {
		fmt.Printf("✅ OPENAI_KEY is set (length: %d)\n", len(openaiKey))
	}
}
