package main

import (
	"fmt"
	"log"

	"github.com/marcodenic/agentry/internal/converse"
)

func main() {
	fmt.Println("🧪 Testing role configuration loading...")
	
	// Test loading agent_0 config
	config, err := converse.LoadRoleConfig("agent_0")
	if err != nil {
		log.Fatalf("Failed to load agent_0 config: %v", err)
	}
	
	fmt.Printf("✅ Agent 0 config loaded:\n")
	fmt.Printf("   Name: %s\n", config.Name)
	fmt.Printf("   Prompt length: %d chars\n", len(config.Prompt))
	fmt.Printf("   Tools: %v\n", config.Tools)
	
	// Test loading coder config
	config, err = converse.LoadRoleConfig("coder")
	if err != nil {
		log.Fatalf("Failed to load coder config: %v", err)
	}
	
	fmt.Printf("✅ Coder config loaded:\n")
	fmt.Printf("   Name: %s\n", config.Name)
	fmt.Printf("   Prompt length: %d chars\n", len(config.Prompt))
	fmt.Printf("   Tools: %v\n", config.Tools)
	
	// Verify that agent_0 has agent tool but coder doesn't
	agent0HasAgentTool := false
	coderHasAgentTool := false
	
	agent0Config, _ := converse.LoadRoleConfig("agent_0")
	for _, tool := range agent0Config.Tools {
		if tool == "agent" {
			agent0HasAgentTool = true
			break
		}
	}
	
	coderConfig, _ := converse.LoadRoleConfig("coder")
	for _, tool := range coderConfig.Tools {
		if tool == "agent" {
			coderHasAgentTool = true
			break
		}
	}
	
	fmt.Printf("\n🔍 Tool access verification:\n")
	fmt.Printf("   Agent 0 has 'agent' tool: %v\n", agent0HasAgentTool)
	fmt.Printf("   Coder has 'agent' tool: %v\n", coderHasAgentTool)
	
	if agent0HasAgentTool && !coderHasAgentTool {
		fmt.Printf("🎉 SUCCESS: Tool restriction is configured correctly!\n")
		fmt.Printf("   - Agent 0 can delegate (has 'agent' tool)\n")
		fmt.Printf("   - Coder cannot delegate (no 'agent' tool)\n")
		fmt.Printf("   - This should prevent infinite recursion\n")
	} else {
		fmt.Printf("❌ ISSUE: Tool restriction not configured correctly\n")
	}
}
