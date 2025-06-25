package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/marcodenic/agentry/internal/converse"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/internal/team"
	"github.com/marcodenic/agentry/internal/tool"
)

// Mock client that simulates LLM responses for agent delegation
type delegationTestClient struct {
	callCount int
}

func (c *delegationTestClient) Complete(ctx context.Context, msgs []model.ChatMessage, tools []model.ToolSpec) (model.Completion, error) {
	c.callCount++
	
	if len(msgs) == 0 {
		return model.Completion{Content: "No messages provided"}, nil
	}
	
	lastMsg := msgs[len(msgs)-1].Content
	fmt.Printf("Mock client received (call %d): %s\n", c.callCount, lastMsg[:min(100, len(lastMsg))])
	
	// If the request is to read roadmap and delegate
	if c.callCount == 1 {
		return model.Completion{
			Content: "I'll read the roadmap file first to understand what development tasks are needed.",
			ToolCalls: []model.ToolCall{
				{
					ID:   "call_1",
					Name: "view",
					Arguments: []byte(`{"path": "ROADMAP.md"}`),
				},
			},
		}, nil
	}
	
	// After reading the roadmap, delegate to coder
	if c.callCount == 2 {
		return model.Completion{
			Content: "Based on the roadmap, I'll delegate some development tasks to the coder agent.",
			ToolCalls: []model.ToolCall{
				{
					ID:   "call_2",
					Name: "agent",
					Arguments: []byte(`{"agent": "coder", "input": "Based on the roadmap, please analyze the current codebase and suggest priority improvements for cross-platform compatibility."}`),
				},
			},
		}, nil
	}
	
	// Coder agent response
	if c.callCount >= 3 {
		return model.Completion{
			Content: "I've analyzed the roadmap and current codebase. The cross-platform compatibility improvements are already implemented. Key achievements include PowerShell integration for Windows and comprehensive test coverage.",
		}, nil
	}
	
	return model.Completion{Content: "Task completed successfully."}, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	fmt.Println("Testing agent delegation scenario: read roadmap and delegate tasks")
	
	// Set sandbox to disabled
	tool.SetSandboxEngine("disabled")
	
	ctx := context.Background()
	registry := tool.DefaultRegistry()
	
	// Create mock client
	client := &delegationTestClient{}
	
	// Create Agent 0 with delegation capabilities
	route := router.Rules{{Name: "test", IfContains: []string{""}, Client: client}}
	agent0 := core.New(route, registry, memory.NewInMemory(), nil, memory.NewInMemoryVector(), nil)
	agent0.Prompt = `You are Agent 0, a system orchestrator. You can read files and delegate tasks to specialized agents like 'coder' using the agent tool.`
	
	// Create coder agent
	coderAgent := core.New(route, registry, memory.NewInMemory(), nil, memory.NewInMemoryVector(), nil)
	coderAgent.Prompt = `You are a coder agent specialized in software development tasks.`
	
	// Create team
	tm, err := converse.NewTeam(agent0, 1, "test")
	if err != nil {
		log.Fatalf("Failed to create team: %v", err)
	}
	
	// Add coder to team
	tm.Add("coder", coderAgent)
	
	ctx = team.WithContext(ctx, tm)
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	
	// Test the exact scenario: read roadmap and delegate tasks
	prompt := "read the roadmap.md and delegate tasks for development"
	
	fmt.Printf("Sending prompt: %s\n", prompt)
	
	response, err := agent0.Run(ctx, prompt)
	if err != nil {
		log.Fatalf("Agent failed with error: %v", err)
	}
	
	fmt.Printf("Agent response: %s\n", response)
	fmt.Println("Test completed successfully!")
}
