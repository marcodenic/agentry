package main

import (
	"context"
	"fmt"
	"log"

	"github.com/marcodenic/agentry/internal/converse"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/tool"
)

func main() {
	fmt.Println("Testing agent/tool name validation fixes...")

	// Create a basic agent
	ag := &core.Agent{
		Tools: tool.DefaultRegistry(),
	}

	// Create team
	team, err := converse.NewTeamContext(ag)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Test 1: Try to call a tool name as an agent (should fail)
	fmt.Println("\nTest 1: Attempting to create agent with tool name 'echo'")
	result, err := team.Call(ctx, "echo", "test input")
	if err != nil {
		fmt.Printf("✅ Correctly rejected tool name: %v\n", err)
	} else {
		fmt.Printf("❌ Incorrectly allowed tool name, result: %s\n", result)
	}

	// Test 2: Try to call another tool name as an agent (should fail)
	fmt.Println("\nTest 2: Attempting to create agent with tool name 'read_lines'")
	result, err = team.Call(ctx, "read_lines", "test input")
	if err != nil {
		fmt.Printf("✅ Correctly rejected tool name: %v\n", err)
	} else {
		fmt.Printf("❌ Incorrectly allowed tool name, result: %s\n", result)
	}

	// Test 3: Try to call with invalid agent name (should fail)
	fmt.Println("\nTest 3: Attempting to create agent with invalid name '123invalid'")
	result, err = team.Call(ctx, "123invalid", "test input")
	if err != nil {
		fmt.Printf("✅ Correctly rejected invalid name: %v\n", err)
	} else {
		fmt.Printf("❌ Incorrectly allowed invalid name, result: %s\n", result)
	}

	// Test 4: Try to call with valid agent name (should succeed)
	fmt.Println("\nTest 4: Attempting to create agent with valid name 'TestAgent'")
	result, err = team.Call(ctx, "TestAgent", "Hello, can you help me?")
	if err != nil {
		fmt.Printf("❌ Incorrectly rejected valid name: %v\n", err)
	} else {
		fmt.Printf("✅ Correctly allowed valid name, agent created\n")
		fmt.Printf("Team now has %d agents\n", len(team.Agents()))
	}

	// Test 5: Verify tool names that should be blocked
	toolNames := []string{"web_search", "read_webpage", "api_request", "download_file", "ping", "fetch", "mcp", "agent"}
	fmt.Println("\nTest 5: Testing additional tool names...")
	for _, toolName := range toolNames {
		_, err := team.Call(ctx, toolName, "test")
		if err != nil {
			fmt.Printf("✅ Tool '%s' correctly blocked\n", toolName)
		} else {
			fmt.Printf("❌ Tool '%s' incorrectly allowed\n", toolName)
		}
	}

	fmt.Println("\n🎉 Agent/tool validation tests completed!")
}
