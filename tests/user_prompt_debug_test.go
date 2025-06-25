package tests

import (
	"context"
	"os"
	"path/filepath"
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

// realisticUserPromptClient simulates an AI that tries to handle the user's request
type realisticUserPromptClient struct{}

func (r realisticUserPromptClient) Complete(ctx context.Context, msgs []model.ChatMessage, tools []model.ToolSpec) (model.Completion, error) {
	// Check the last message to understand the request
	if len(msgs) == 0 {
		return model.Completion{Content: "No messages provided"}, nil
	}
	
	lastMsg := msgs[len(msgs)-1].Content
	
	// If it's the user's request about reading README and delegating tasks
	if contains(lastMsg, "read the readme.md and delegate tasks for development") {
		return model.Completion{
			Content: "I'll read the README.md file first and then delegate development tasks to appropriate agents.",
			ToolCalls: []model.ToolCall{
				{
					ID:   "call_1",
					Name: "view",
					Arguments: []byte(`{"path": "README.md"}`),
				},
			},
		}, nil
	}
	
	// If it's a tool response, try to delegate to coder agent
	if contains(lastMsg, "Test Project") || contains(lastMsg, "Feature A") {
		return model.Completion{
			Content: "Based on the README, I can see this is a test project with features A, B, and C. I'll delegate the development tasks to a coder agent.",
			ToolCalls: []model.ToolCall{
				{
					ID:   "call_2",
					Name: "agent",
					Arguments: []byte(`{"agent": "coder", "input": "Implement feature A from the README.md"}`),
				},
			},
		}, nil
	}
	
	return model.Completion{Content: "Request processed"}, nil
}

func TestUserPromptDebug(t *testing.T) {
	// Your exact prompt
	userPrompt := "read the readme.md and delegate tasks for development"
	
	t.Logf("🔍 Testing user prompt: %s", userPrompt)
	
	// Set sandbox to disabled to prevent cri-shim issues
	tool.SetSandboxEngine("disabled")
	t.Log("🔧 Sandbox engine set to: disabled")
	
	// Set up workspace directory
	workspaceDir := filepath.Join(os.TempDir(), "agentry_test_workspace")
	os.MkdirAll(workspaceDir, 0755)
	defer os.RemoveAll(workspaceDir)
	
	// Create a test README.md
	readmeContent := `# Test Project
This is a test project for development.

## Features
- Feature A
- Feature B
- Feature C

## Development Tasks
- Implement feature A
- Add tests for feature B
- Update documentation
`
	readmePath := filepath.Join(workspaceDir, "README.md")
	os.WriteFile(readmePath, []byte(readmeContent), 0644)
		// Change to the workspace directory so tools can find the README
	origDir, _ := os.Getwd()
	os.Chdir(workspaceDir)
	defer os.Chdir(origDir)
		// Verify the README.md file exists
	if _, err := os.Stat("README.md"); err != nil {
		t.Fatalf("README.md file not found in workspace: %v", err)
	}
	
	// Debug: check current working directory and list files
	cwd, _ := os.Getwd()
	t.Logf("📄 Current working directory: %s", cwd)
	
	files, _ := os.ReadDir(".")
	t.Log("📄 Files in current directory:")
	for _, file := range files {
		t.Logf("  - %s", file.Name())
	}
	
	t.Log("📄 README.md file created and verified in workspace")
	
	// Get tool registry
	t.Log("📦 Getting tool registry...")
	registry := tool.DefaultRegistry()
	
	// Log available tools
	t.Log("🔧 Available tools in registry:")
	for toolName := range registry {
		t.Logf("  - %s", toolName)
	}
	
	// Create Agent 0 with realistic prompt and client
	agent0Prompt := `You are Agent 0, the system orchestrator in a multi-agent environment. Your primary job is
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

	route := router.Rules{{Name: "realistic", IfContains: []string{""}, Client: realisticUserPromptClient{}}}
	agent0 := core.New(route, registry, memory.NewInMemory(), nil, memory.NewInMemoryVector(), nil)
	agent0.Prompt = agent0Prompt
	
	t.Logf("🤖 Agent 0 created with %d tools available", len(registry))
	
	// Create team context so the agent tool can work
	tm, err := converse.NewTeam(agent0, 1, "test")
	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}
	ctx := team.WithContext(context.Background(), tm)
	
	// Add timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	
	// Process the user prompt
	t.Log("💬 Processing user prompt with Agent 0...")
	response, err := agent0.Run(ctx, userPrompt)
	
	if err != nil {
		t.Logf("❌ Agent 0 returned error: %v", err)
	} else {
		t.Logf("✅ Agent 0 response received")
	}
	
	t.Logf("📝 Agent 0 Raw Response:\n%s", response)
	
	// Check if response contains error indicators
	if response == "" {
		t.Error("❌ Empty response from Agent 0")
	}
	
	// Look for common error patterns
	errorPatterns := []string{
		"ERR: unknown agent",
		"unknown tool",
		"cri-shim",
		"sandbox error",
		"command not found",
		"not found",
		"error",
	}
	
	foundErrors := []string{}
	for _, pattern := range errorPatterns {
		if contains(response, pattern) {
			foundErrors = append(foundErrors, pattern)
			t.Logf("⚠️  Found error pattern '%s' in response", pattern)
		}
	}
	
	// Check for positive indicators
	positivePatterns := []string{
		"README",
		"delegate",
		"coder",
		"Feature A",
		"development",
	}
	
	foundPositive := []string{}
	for _, pattern := range positivePatterns {
		if contains(response, pattern) {
			foundPositive = append(foundPositive, pattern)
			t.Logf("✅ Found expected pattern '%s' in response", pattern)
		}
	}
	
	// Summary
	t.Logf("📊 Summary:")
	t.Logf("   Errors found: %d (%v)", len(foundErrors), foundErrors)
	t.Logf("   Expected patterns: %d (%v)", len(foundPositive), foundPositive)
	
	if len(foundErrors) > 0 {
		t.Logf("⚠️  Issues detected that need fixing")
	} else {
		t.Logf("✅ No obvious errors detected")
	}
	
	t.Log("🏁 User prompt debug test completed")
}

func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
