package main

import (
	"fmt"
	"strings"

	"github.com/marcodenic/agentry/internal/converse"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/internal/tool"
)

func main() {
	fmt.Println("🧪 Testing coder agent fix...")
	
	// Set up basic infrastructure
	registry := tool.DefaultRegistry()
	route := router.Rules{{Name: "test", Client: nil}}
	
	// Create a base agent
	baseAgent := core.New(route, registry, memory.NewInMemory(), nil, memory.NewInMemoryVector(), nil)
	
	// Create a team 
	team, err := converse.NewTeam(baseAgent, 1, "test")
	if err != nil {
		panic(err)
	}
	
	// Add coder agent - this should load the coder.yaml config
	agent, name := team.AddAgent("coder")
	
	fmt.Printf("✅ Coder agent created with name: %s\n", name)
	fmt.Printf("🔧 Coder agent has %d tools\n", len(agent.Tools))
	
	// Check if the coder agent has the 'agent' tool (it should NOT)
	hasAgentTool := false
	toolNames := []string{}
	for toolName := range agent.Tools {
		toolNames = append(toolNames, toolName)
		if toolName == "agent" {
			hasAgentTool = true
		}
	}
	
	fmt.Printf("📋 Available tools: %s\n", strings.Join(toolNames, ", "))
	
	if hasAgentTool {
		fmt.Printf("❌ PROBLEM: Coder agent still has 'agent' tool - can still try to delegate\n")
	} else {
		fmt.Printf("✅ SUCCESS: Coder agent does NOT have 'agent' tool - cannot delegate\n")
		fmt.Printf("💡 This should fix the 'trying to create view agent' error\n")
	}
	
	// Check that it has the essential tools for file operations
	essentialTools := []string{"view", "write", "edit_range", "search_replace"}
	fmt.Printf("\n🔍 Checking for essential tools:\n")
	for _, tool := range essentialTools {
		if _, exists := agent.Tools[tool]; exists {
			fmt.Printf("  ✅ %s - available\n", tool)
		} else {
			fmt.Printf("  ❌ %s - missing\n", tool)
		}
	}
	
	fmt.Printf("\n🎯 Summary: The coder agent should now work directly with tools instead of trying to delegate to other agents.\n")
}
