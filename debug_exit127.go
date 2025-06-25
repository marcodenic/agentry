package main

import (
	"context"
	"fmt"

	"github.com/marcodenic/agentry/internal/tool"
)

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
	
	// List available tools
	fmt.Println("Available tools:")
	for name := range reg {
		fmt.Printf("  - %s\n", name)
	}
	
	// Test powershell tool directly with the command that might be causing issues
	if powershellTool, exists := reg.Use("powershell"); exists {
		fmt.Println("\nTesting powershell tool to read README.md...")
		result, err := powershellTool.Execute(ctx, map[string]any{
			"command": "Get-Content README.md | Select-Object -First 5",
		})
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
		} else {
			fmt.Printf("SUCCESS: First 100 chars: %s\n", result[:min(100, len(result))])
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
