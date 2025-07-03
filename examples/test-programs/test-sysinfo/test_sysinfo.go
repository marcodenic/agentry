package main

import (
	"context"
	"fmt"

	"github.com/marcodenic/agentry/internal/tool"
)

func main() {
	registry := tool.DefaultRegistry()
	
	// Test if sysinfo tool exists
	sysinfoTool, exists := registry.Use("sysinfo")
	if !exists {
		fmt.Println("sysinfo tool not found")
		return
	}
	
	fmt.Println("sysinfo tool found successfully!")
	fmt.Printf("Tool name: %s\n", sysinfoTool.Name())
	fmt.Printf("Tool description: %s\n", sysinfoTool.Description())
	
	// Test execution (this might fail if sandbox restrictions exist)
	fmt.Println("\nTesting sysinfo execution...")
	result, err := sysinfoTool.Execute(context.Background(), map[string]any{})
	if err != nil {
		fmt.Printf("sysinfo execution failed: %v\n", err)
		fmt.Printf("Error type: %T\n", err)
	} else {
		fmt.Printf("sysinfo result (first 200 chars): %.200s...\n", result)
	}
}
