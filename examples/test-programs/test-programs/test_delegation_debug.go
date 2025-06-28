package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/marcodenic/agentry/internal/converse"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/internal/tool"
)

// mockClient simulates an AI that tries to delegate
type mockClient struct{}

func (m mockClient) Complete(ctx context.Context, msgs []model.ChatMessage, tools []model.ToolSpec) (model.Completion, error) {
	fmt.Printf("🤖 AI called with %d messages, %d tools available\n", len(msgs), len(tools))
	
	if len(msgs) == 0 {
		return model.Completion{Content: "No messages provided"}, nil
	}
	
	lastMsg := msgs[len(msgs)-1].Content
	fmt.Printf("📝 Last message: %s\n", lastMsg)
	
	// Check if tools are available
	hasAgentTool := false
	for _, tool := range tools {
		fmt.Printf("🔧 Available tool: %s\n", tool.Name)
		if tool.Name == "agent" {
			hasAgentTool = true
		}
	}
	
	if !hasAgentTool {
		return model.Completion{Content: "No agent tool available for delegation"}, nil
	}
	
	// Try to delegate to coder
	return model.Completion{
		Content: "I'll delegate this task to a coder agent.",
		ToolCalls: []model.ToolCall{
			{
				ID:   "call_123",
				Name: "agent",
				Arguments: []byte(`{"agent": "coder", "input": "Help with this task"}`),
			},
		},
	}, nil
}

func main() {
	fmt.Println("🔍 Testing delegation debug scenario...")
	
	// Set sandbox to disabled to prevent issues
	tool.SetSandboxEngine("disabled")
	fmt.Println("🔧 Sandbox engine set to: disabled")
	
	// Set up workspace directory
	workspaceDir := filepath.Join(os.TempDir(), "agentry_delegation_test")
	os.MkdirAll(workspaceDir, 0755)
	defer os.RemoveAll(workspaceDir)
	
	// Change to the workspace directory
	origDir, _ := os.Getwd()
	os.Chdir(workspaceDir)
	defer os.Chdir(origDir)
	
	// Create mock client
	client := mockClient{}
	route := router.Rules{
		{Pattern: ".*", Client: client},
	}
	
	// Create agent with builtins
	agent := &core.Agent{
		Prompt:        "You are Agent 0, the system orchestrator.",
		Mem:           memory.NewInMemory(),
		Route:         route,
		Tools:         tool.NewBuiltins(),
		MaxIterations: 10,
	}
	
	fmt.Printf("🏗️ Agent created with %d tools\n", len(agent.Tools.List()))
	for _, toolName := range agent.Tools.List() {
		fmt.Printf("   - %s\n", toolName)
	}
	
	// Create team and add agent
	fmt.Println("👥 Creating team...")
	team := converse.NewTeamContext(agent)
	
	// Test delegation
	fmt.Println("🚀 Testing delegation: 'Create a Python script'")
	result, err := team.Call(context.Background(), "Agent0", "Create a Python script")
	
	if err != nil {
		fmt.Printf("❌ Delegation failed: %v\n", err)
		log.Printf("Error details: %+v", err)
	} else {
		fmt.Printf("✅ Delegation successful: %s\n", result)
	}
}
