package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/marcodenic/agentry/internal/tool"
)

func main() {
	// Enable debug logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("[DEBUG] Working directory debugging script")

	// Show current working directory
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}
	fmt.Printf("📁 Current working directory: %s\n", cwd)

	// List files in current directory
	files, err := os.ReadDir(".")
	if err != nil {
		log.Fatalf("Failed to read directory: %v", err)
	}
	
	fmt.Printf("📋 Files in current directory:\n")
	for _, file := range files {
		if file.IsDir() {
			fmt.Printf("  📁 %s/\n", file.Name())
		} else {
			fmt.Printf("  📄 %s\n", file.Name())
		}
	}

	// Check specifically for TODO.md
	if _, err := os.Stat("TODO.md"); os.IsNotExist(err) {
		fmt.Printf("❌ TODO.md not found in current directory\n")
	} else {
		fmt.Printf("✅ TODO.md found in current directory\n")
	}

	// Test the view tool with different paths
	registry := tool.DefaultRegistry()
	viewTool, hasView := registry.Use("view")
	if !hasView {
		fmt.Printf("❌ View tool not found\n")
		return
	}

	testPaths := []string{
		"TODO.md",
		"./TODO.md",
		filepath.Join(cwd, "TODO.md"),
	}

	for _, path := range testPaths {
		fmt.Printf("\n🔧 Testing view tool with path: %s\n", path)
		result, err := viewTool.Execute(context.Background(), map[string]any{
			"path": path,
		})
		if err != nil {
			fmt.Printf("❌ Failed: %v\n", err)
		} else {
			fmt.Printf("✅ Success! Result length: %d\n", len(result))
			fmt.Printf("📄 First 100 chars: %s\n", result[:min(100, len(result))])
		}
	}

	fmt.Println("\n💡 Now run TUI mode and compare the working directory and file access")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
