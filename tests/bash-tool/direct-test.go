package main

import (
	"context"
	"fmt"
	"log"

	"github.com/marcodenic/agentry/internal/tool"
)

func main() {
	ctx := context.Background()
	
	// Create a simple tool manager
	tm := tool.New(".", "", nil)
	
	fmt.Println("Testing bash tool directly...")
	
	result, err := tm.Call(ctx, "bash", map[string]any{
		"command": "echo hello world",
	})
	
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		log.Printf("Success: %s", result)
	}
}
