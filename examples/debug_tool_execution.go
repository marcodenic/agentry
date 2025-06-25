package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/marcodenic/agentry/internal/tool"
)

func main() {
	fmt.Printf("Testing tools on %s...\n", runtime.GOOS)
	
	// Set sandbox to disabled
	tool.SetSandboxEngine("disabled")
	
	ctx := context.Background()
	registry := tool.DefaultRegistry()
	
	// Test view tool to read ROADMAP.md
	fmt.Println("Testing view tool...")
	viewTool, exists := registry.Use("view")
	if !exists {
		log.Fatal("view tool not found")
	}
	
	// Check if ROADMAP.md exists
	if _, err := os.Stat("ROADMAP.md"); os.IsNotExist(err) {
		log.Fatal("ROADMAP.md not found")
	}
	
	viewArgs := map[string]any{
		"path": "ROADMAP.md",
	}
	
	result, err := viewTool.Execute(ctx, viewArgs)
	if err != nil {
		log.Fatalf("view tool failed: %v", err)
	}
	
	fmt.Printf("Successfully read ROADMAP.md (first 200 chars): %s...\n", 
		truncate(result, 200))
	
	// Test ls tool
	fmt.Println("Testing ls tool...")
	lsTool, exists := registry.Use("ls")
	if !exists {
		log.Fatal("ls tool not found")
	}
	
	lsArgs := map[string]any{
		"path": ".",
	}
	
	result, err = lsTool.Execute(ctx, lsArgs)
	if err != nil {
		log.Fatalf("ls tool failed: %v", err)
	}
	
	fmt.Printf("Successfully listed directory (first 200 chars): %s...\n", 
		truncate(result, 200))
	
	fmt.Println("All tools working correctly!")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
