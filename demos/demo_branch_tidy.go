package main

import (
	"context"
	"fmt"

	"github.com/marcodenic/agentry/internal/tool"
)

func main() {
	registry := tool.DefaultRegistry()
	branchTidy, exists := registry.Use("branch-tidy")
	if !exists {
		fmt.Println("branch-tidy tool not found")
		return
	}

	fmt.Println("=== Branch Tidy Tool Demo ===")
	fmt.Println("\nTool Description:", branchTidy.Description())
	
	schema := branchTidy.JSONSchema()
	fmt.Println("\nTool Schema:")
	fmt.Printf("  Type: %v\n", schema["type"])
	
	if props, ok := schema["properties"].(map[string]any); ok {
		fmt.Println("  Properties:")
		for name, prop := range props {
			if propMap, ok := prop.(map[string]any); ok {
				fmt.Printf("    - %s: %v\n", name, propMap["description"])
			}
		}
	}

	if example, ok := schema["example"].(map[string]any); ok {
		fmt.Println("\nExample usage:")
		for key, value := range example {
			fmt.Printf("  %s: %v\n", key, value)
		}
	}

	fmt.Println("\n=== Dry Run Demo ===")
	result, err := branchTidy.Execute(context.Background(), map[string]any{
		"dry-run": true,
		"force":   false,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Result:\n%s\n", result)
	}
}
