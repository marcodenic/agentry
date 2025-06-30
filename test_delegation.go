package main

import (
	"context"
	"fmt"
	"log"

	"github.com/marcodenic/agentry/internal/config"
	"github.com/marcodenic/agentry/internal/team"
)

func main() {
	// Load config
	cfg, err := config.LoadFile("smart-config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	// Build agent using the same function as TUI
	agent, err := buildAgent(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Create team
	tm, err := team.NewTeam(agent, 10, "")
	if err != nil {
		log.Fatal(err)
	}

	// Set team context
	ctx := team.WithContext(context.Background(), tm)

	// Test delegation
	fmt.Println("🧪 Testing agent delegation...")
	
	// First, check if agent tool is available
	if agentTool, exists := agent.Tools["agent"]; exists {
		fmt.Println("✅ Agent tool is available:", agentTool.Description())
		
		// Try to call the agent tool directly
		args := map[string]any{
			"agent": "coder",
			"input": "Hello from test",
		}
		
		result, err := agentTool.Call(ctx, args)
		if err != nil {
			fmt.Printf("❌ Agent tool call failed: %v\n", err)
		} else {
			fmt.Printf("✅ Agent tool call succeeded: %s\n", result)
		}
	} else {
		fmt.Println("❌ Agent tool not found!")
	}
}
