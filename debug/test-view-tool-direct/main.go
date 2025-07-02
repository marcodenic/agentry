package main

import (
	"context"
	"fmt"

	"github.com/marcodenic/agentry/internal/tool"
)

func main() {
	fmt.Printf("🔧 Testing view tool directly...\n")

	// Get default registry
	registry := tool.DefaultRegistry()

	// Check if view tool exists
	viewTool, hasView := registry.Use("view")
	if !hasView {
		fmt.Printf("❌ View tool not found in registry\n")
		return
	}

	fmt.Printf("✅ View tool found: %s\n", viewTool.Description())

	// Test the view tool with TODO.md
	fmt.Printf("\n📖 Testing view tool with TODO.md...\n")
	result, err := viewTool.Execute(context.Background(), map[string]any{
		"path": "TODO.md",
	})

	if err != nil {
		fmt.Printf("❌ View tool failed: %v\n", err)
	} else {
		fmt.Printf("✅ View tool succeeded! Result length: %d\n", len(result))
		fmt.Printf("📝 First 200 chars of result:\n%s\n", result[:minInt(200, len(result))])
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
