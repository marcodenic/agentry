package main

import (
	"context"
	"fmt"

	"github.com/marcodenic/agentry/internal/tool"
)

func main() {
	// Set sandbox to disabled
	tool.SetSandboxEngine("disabled")
	
	// Create context
	ctx := context.Background()
	
	// Create tool registry
	reg := tool.DefaultRegistry()
	
	// Test powershell tool directly with the command that might be causing issues
	if powershellTool, exists := reg.Use("powershell"); exists {
		fmt.Println("Testing powershell tool to read README.md...")
		result, err := powershellTool.Execute(ctx, map[string]any{
			"command": "Get-Content README.md | Select-Object -First 5",
		})
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
		} else {
			fmt.Printf("SUCCESS: %s\n", result)
		}
		
		// Test a simpler command  
		fmt.Println("\nTesting simpler powershell command...")
		result, err = powershellTool.Execute(ctx, map[string]any{
			"command": "echo 'hello world'",
		})
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
		} else {
			fmt.Printf("SUCCESS: %s\n", result)
		}
	} else {
		fmt.Println("ERROR: powershell tool not found")
	}
}
