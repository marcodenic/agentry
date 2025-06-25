package tests

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/marcodenic/agentry/internal/converse"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/internal/team"
	"github.com/marcodenic/agentry/internal/tool"
)

// TestUserPromptScenario tests the exact user prompt that's causing issues
func TestUserPromptScenario(t *testing.T) {
	t.Log("🧪 Testing user's exact prompt scenario...")
	
	// Set sandbox to disabled to prevent cri-shim issues
	tool.SetSandboxEngine("disabled")
	t.Log("🔧 Sandbox engine set to: disabled")
	
	// Create a test README.md in current directory for testing
	readmeContent := `# Test Project for Agentry
This is a test project to verify agent delegation functionality.

## Features
- Multi-agent coordination
- Task delegation
- File operations
- Development automation

## Development Tasks
- Implement core features
- Add comprehensive tests
- Improve documentation
- Set up CI/CD pipeline
`
	
	// Write test README.md (will be cleaned up after test)
	if err := os.WriteFile("README.md", []byte(readmeContent), 0644); err != nil {
		t.Fatalf("Failed to create test README.md: %v", err)
	}
	defer os.Remove("README.md") // Clean up after test
	
	// Create tool registry with all the tools needed
	reg := tool.DefaultRegistry()
	t.Logf("🛠️  Tool registry created with %d tools:", len(reg))
	for name := range reg {
		t.Logf("   - %s", name)
	}
	
	// Create Agent 0 with realistic AI client that will try to delegate
	route := router.Rules{{Name: "mock", IfContains: []string{""}, Client: &realisticMockClient{}}}
	
	agent0 := core.New(route, reg, memory.NewInMemory(), nil, memory.NewInMemoryVector(), nil)
	
	// Load Agent 0 prompt
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
- Shell tools – powershell/cmd (Windows), bash/sh (Unix/Linux/macOS)
- Network tools – fetch, ping
- Utility tools – echo, patch

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

	t.Logf("👤 Agent 0 created with prompt length: %d chars", len(agent0.Prompt))
	t.Logf("🛠️  Agent 0 has access to %d tools", len(agent0.Tools))
	
	// Create team
	tm, err := converse.NewTeam(agent0, 1, "test")
	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}
	ctx := team.WithContext(context.Background(), tm)
	
	// Add timeout to prevent hanging
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	
	t.Log("🚀 Testing with user's exact prompt...")
	
	// THE EXACT USER PROMPT THAT'S CAUSING ISSUES
	userPrompt := "read the readme.md and delegate tasks for development"
	
	t.Logf("💬 User prompt: %s", userPrompt)
	
	// Execute and see what happens
	output, err := agent0.Run(ctx, userPrompt)
	
	if err != nil {
		t.Logf("❌ Agent execution failed: %v", err)
		t.Logf("🔍 This is where we need to investigate the cri-shim error")
		
		// Check if it's a timeout (infinite recursion)
		if err == context.DeadlineExceeded {
			t.Error("⏰ Test timed out - likely infinite recursion")
		}
		
		// Check if it's cri-shim related
		if err.Error() != "" {
			t.Logf("🔍 Error details: %s", err.Error())
		}
	} else {
		t.Logf("✅ Agent execution completed")
		t.Logf("📝 Output: %s", output)
	}
	
	// Check team state
	t.Log("👥 Team state after execution completed")
}

// Realistic mock client that simulates what a real AI would do with the user's prompt
type realisticMockClient struct{}

func (m *realisticMockClient) Complete(ctx context.Context, msgs []model.ChatMessage, tools []model.ToolSpec) (model.Completion, error) {
	// Analyze the last user message to see what the AI would likely do
	lastMsg := msgs[len(msgs)-1]
	
	// If the message mentions "read the readme.md", the AI would likely:
	// 1. First try to read the file using 'view' tool
	// 2. Then delegate development tasks to a coder
	
	if lastMsg.Content == "read the readme.md and delegate tasks for development" {
		// First response: try to read README.md
		return model.Completion{
			Content: "I'll first read the README.md file to understand the project, then delegate development tasks accordingly.",
			ToolCalls: []model.ToolCall{				{
					ID:   "call_view_readme",
					Name: "view",
					Arguments: []byte(`{"path": "README.md"}`),
				},
			},
		}, nil
	}
	
	// If we get tool results about README, then delegate
	if len(msgs) > 2 && msgs[len(msgs)-2].Role == "tool" {
		return model.Completion{
			Content: "Based on the README, I'll delegate development tasks to a coder agent.",
			ToolCalls: []model.ToolCall{
				{
					ID:   "call_delegate_coder",
					Name: "agent",
					Arguments: []byte(`{"agent": "coder", "input": "review project structure and implement development tasks based on README requirements"}`),
				},
			},
		}, nil
	}
	
	// Default response
	return model.Completion{
		Content: "I understand your request and will process it accordingly.",
		ToolCalls: []model.ToolCall{},
	}, nil
}
