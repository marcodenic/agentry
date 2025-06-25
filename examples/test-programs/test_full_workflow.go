package main

import (
	"context"
	"fmt"
	"log"

	"github.com/marcodenic/agentry/internal/config"
	"github.com/marcodenic/agentry/internal/converse"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/internal/team"
	"github.com/marcodenic/agentry/internal/tool"
)

func main() {
	fmt.Println("🧪 Testing Complete Agent Spawning Workflow...")
	
	// Load actual config with disabled sandbox
	cfg, err := config.Load(".agentry.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	
	fmt.Printf("✅ Config loaded successfully\n")
	fmt.Printf("   Sandbox engine: %s\n", cfg.Sandbox.Engine)
	fmt.Printf("   Tools configured: %d\n", len(cfg.Tools))
	
	// Create tool registry from config
	reg, err := tool.BuildRegistry(cfg.Tools)
	if err != nil {
		log.Fatalf("Failed to build tool registry: %v", err)
	}
	
	fmt.Printf("✅ Tool registry built with %d tools\n", len(reg))
	for name := range reg {
		fmt.Printf("   - %s\n", name)
	}
	
	// Set sandbox engine
	tool.SetSandboxEngine(cfg.Sandbox.Engine)
	fmt.Printf("✅ Sandbox engine set to: %s\n", cfg.Sandbox.Engine)
	
	// Create Agent 0 with proper router (using mock for this test)
	route := router.Rules{{Name: "mock", IfContains: []string{""}, Client: &mockClient{}}}
	
	agent0 := core.New(route, reg, memory.NewInMemory(), nil, memory.NewInMemoryVector(), nil)
	
	// Load Agent 0 prompt from config
	agent0.Prompt = `You are Agent 0, the system orchestrator in a multi-agent environment. Your primary job is
to analyse user requests and either act directly or delegate work to specialised agents.

## Core Responsibilities
1. **Natural Language Analysis** – Understand each request and determine scope and complexity.
2. **Agent Management** – Spawn and coordinate sub‑agents when tasks require special skills.
3. **Team Orchestration** – Create teams for complex workflows and keep them aligned.
4. **Task Delegation** – Use the agent tool to assign work to other agents.
5. **Direct Assistance** – Handle simple questions yourself using available tools.

## Decision Framework
- **Simple tasks** – Handle yourself.
- **Specialised tasks** – Delegate to the appropriate agent (coder, researcher, etc.).
- **Complex projects** – Form a team and coordinate execution.
- **Parallel work** – Spawn multiple agents to run at the same time when tasks are independent.

## Available Tools for Orchestration
- agent – Delegate tasks. Usage example: {"agent": "agent_name", "input": "task"}
- Standard tools – ls, view, write, edit, bash, fetch, etc.

## Agent Types You Can Spawn
- **coder** – Software development
- **researcher** – Information gathering
- **analyst** – Data analysis
- **writer** – Documentation and content
- **planner** – Project planning
- **tester** – Quality assurance
- **devops** – Deployment and automation

## Behavioural Guidelines
- **Be proactive** – Spawn agents without asking when needed.
- **Think step-by-step** – Break large tasks into smaller pieces.
- **Delegate wisely** – Match tasks to the right agent.
- **Stay coordinated** – Track progress of sub‑agents and keep them focused.
- **Be efficient** – Use parallel execution for independent tasks.

Remember: users expect you to manage the entire system efficiently. Do not over‑explain your decisions – execute the optimal strategy.`

	fmt.Printf("✅ Agent 0 created with prompt length: %d chars\n", len(agent0.Prompt))
	
	// Create team
	tm, err := converse.NewTeam(agent0, 1, "test")
	if err != nil {
		log.Fatalf("Failed to create team: %v", err)
	}
	ctx := team.WithContext(context.Background(), tm)
	
	fmt.Printf("✅ Team created successfully\n")
	
	// Test delegation
	fmt.Println("\n🚀 Testing agent delegation...")
	input := "I need help creating a Python web scraper"
	
	output, err := agent0.Run(ctx, input)
	
	if err != nil {
		fmt.Printf("❌ Agent execution failed: %v\n", err)
	} else {
		fmt.Printf("✅ Agent execution completed successfully\n")
		fmt.Printf("📝 Output:\n%s\n", output)
	}
	
	fmt.Println("\n🎉 Test completed - check console output for debug information")
}

// Mock client that simulates Agent 0 delegating to coder
type mockClient struct{}

func (m *mockClient) Complete(ctx context.Context, msgs []interface{}, tools []interface{}) (interface{}, error) {
	// This is a simplified mock - in reality it would return proper model.Completion
	// but this test is just to verify our configuration and tool registry work
	return map[string]interface{}{
		"Content": "I'll delegate this web scraping task to a coder agent.",
		"ToolCalls": []map[string]interface{}{
			{
				"ID":   "call_123",
				"Name": "agent",
				"Arguments": `{"agent": "coder", "input": "create Python web scraper"}`,
			},
		},
	}, nil
}
