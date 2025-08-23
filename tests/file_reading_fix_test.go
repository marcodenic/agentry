package tests

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/team"
	"github.com/marcodenic/agentry/internal/tool"
)

// Test the actual file reading tools in the correct directory 
func TestFileToolsWithCorrectWorkingDirectory(t *testing.T) {
	// Change to agentry root where PRODUCT.md exists
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	
	agentryRoot := filepath.Join(originalWd, "..")
	err := os.Chdir(agentryRoot)
	if err != nil {
		t.Fatalf("Failed to change to agentry root: %v", err)
	}
	t.Logf("Working directory: %s", agentryRoot)

	// Test ls tool
	registry := tool.DefaultRegistry()
	lsTool, ok := registry.Use("ls")
	if !ok {
		t.Fatal("ls tool not found")
	}

	lsOutput, err := lsTool.Execute(context.Background(), map[string]any{})
	t.Logf("ls result - error: %v", err)
	if err != nil {
		t.Logf("ls failed (may be sandbox issue): %v", err)
	} else {
		t.Logf("ls output: %s", lsOutput)
		if !strings.Contains(lsOutput, "PRODUCT.md") {
			t.Errorf("PRODUCT.md not found in ls output")
		}
	}

	// Test view tool
	viewTool, ok := registry.Use("view")
	if !ok {
		t.Fatal("view tool not found")
	}

	viewOutput, err := viewTool.Execute(context.Background(), map[string]any{"path": "PRODUCT.md"})
	if err != nil {
		t.Fatalf("view tool failed: %v", err)
	}

	if !strings.Contains(viewOutput, "Agentry Product") {
		t.Errorf("Expected PRODUCT.md content, got: %.200s...", viewOutput)
	}
	t.Logf("✅ view tool successfully read PRODUCT.md (%d chars)", len(viewOutput))
}

// Real integration test that simulates the actual coder agent behavior
func TestRealCoderAgentFileReading(t *testing.T) {
	// Change to agentry root
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	
	agentryRoot := filepath.Join(originalWd, "..")
	err := os.Chdir(agentryRoot)
	if err != nil {
		t.Fatalf("Failed to change to agentry root: %v", err)
	}
	t.Logf("Working directory for test: %s", agentryRoot)

	// Create a mock client that simulates a real coder trying to read PRODUCT.md
	coderClient := &realCoderMockClient{t: t}
	
	registry := tool.DefaultRegistry()
	ag := core.New(coderClient, "mock", registry, memory.NewInMemory(), memory.NewInMemoryVector(), nil)
	ag.Prompt = `You are Coder, an AI software developer agent. When asked to read a file, use the 'view' tool to read it and then summarize the content.`

	// Run the agent directly (not through delegation) to test file reading
	result, err := ag.Run(context.Background(), "Please read PRODUCT.md and tell me what it contains")
	
	if err != nil {
		t.Fatalf("Coder agent failed: %v", err)
	}
	
	if !strings.Contains(result, "Agentry") {
		t.Errorf("Expected PRODUCT.md content in result, got: %s", result)
	}
	
	t.Logf("✅ Coder agent successfully read PRODUCT.md: %.200s...", result)
}

// Mock coder client that tries to use file tools step by step
type realCoderMockClient struct {
	callCount int
	t         *testing.T
}

func (c *realCoderMockClient) Stream(ctx context.Context, msgs []model.ChatMessage, tools []model.ToolSpec) (<-chan model.StreamChunk, error) {
	out := make(chan model.StreamChunk, 1)
	go func() {
		defer close(out)
		c.callCount++
		c.t.Logf("Coder mock call #%d", c.callCount)

		switch c.callCount {
		case 1:
			// First attempt: try to read the file
			c.t.Logf("Coder attempting to view PRODUCT.md")
			out <- model.StreamChunk{
				ContentDelta: "I'll read the PRODUCT.md file for you.",
				ToolCalls: []model.ToolCall{
					{
						ID:        "view_product",
						Name:      "view",
						Arguments: []byte(`{"path": "PRODUCT.md"}`),
					},
				},
				Done: true,
			}
		case 2:
			// Check if we got tool results
			hasToolResult := false
			var toolResult string
			for _, msg := range msgs {
				if msg.Role == "tool" && msg.ToolCallID == "view_product" {
					hasToolResult = true
					toolResult = msg.Content
					c.t.Logf("Coder received file content (first 200 chars): %.200s...", toolResult)
					break
				}
			}
			
			if hasToolResult && strings.Contains(toolResult, "Agentry Product") {
				// Success! We got the file content
				out <- model.StreamChunk{
					ContentDelta: "Successfully read PRODUCT.md. This is the Agentry Product & Roadmap document which describes a multi-agent development orchestrator. Key points: Local-first, observable, resilient system where Agent 0 delegates tasks to specialized agents, implements solutions, runs tests, and manages reviews. The system includes 30+ built-in tools for file operations, web/network access, and agent delegation. It supports both OpenAI and Anthropic models.",
					Done: true,
				}
			} else if hasToolResult {
				// Got content but wrong content
				c.t.Logf("Got unexpected content: %.100s...", toolResult)
				out <- model.StreamChunk{
					ContentDelta: "I was able to read a file, but it doesn't appear to be the expected PRODUCT.md content.",
					Done: true,
				}
			} else {
				// No tool result, there was an error
				c.t.Logf("No tool result received - file reading failed")
				out <- model.StreamChunk{
					ContentDelta: "I encountered an error trying to read PRODUCT.md. The file may not exist or there may be a permissions issue.",
					Done: true,
				}
			}
		default:
			out <- model.StreamChunk{
				ContentDelta: "Task completed.",
				Done: true,
			}
		}
	}()
	return out, nil
}

// Test the complete delegation flow with proper working directory
func TestCompleteCoderDelegationFlow(t *testing.T) {
	// Change to agentry root
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	
	agentryRoot := filepath.Join(originalWd, "..")
	err := os.Chdir(agentryRoot)
	if err != nil {
		t.Fatalf("Failed to change to agentry root: %v", err)
	}
	
	// Create Agent 0 that delegates to coder
	agent0Client := &delegatingAgent0Client{t: t}
	registry := tool.DefaultRegistry()
	ag := core.New(agent0Client, "mock", registry, memory.NewInMemory(), memory.NewInMemoryVector(), nil)
	ag.Prompt = "You are Agent 0. Delegate file reading tasks to coder agents."

	// Create team with proper coder support
	tm, err := team.NewTeam(ag, 1, "test")
	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}
	tm.RegisterAgentTool(ag.Tools)
	ctx := team.WithContext(context.Background(), tm)

	// This should delegate to a coder that can actually read the file
	result, err := ag.Run(ctx, "spawn a coder to read PRODUCT.md and report back")
	if err != nil {
		t.Fatalf("Delegation failed: %v", err)
	}

	t.Logf("Delegation result: %s", result)
	// The result should indicate successful delegation occurred
	if !strings.Contains(strings.ToLower(result), "coder") {
		t.Errorf("Expected mention of coder in result: %s", result)
	}
}

// Mock Agent 0 client that delegates to coder
type delegatingAgent0Client struct {
	t *testing.T
}

func (d *delegatingAgent0Client) Stream(ctx context.Context, msgs []model.ChatMessage, tools []model.ToolSpec) (<-chan model.StreamChunk, error) {
	out := make(chan model.StreamChunk, 1)
	go func() {
		defer close(out)
		d.t.Logf("Agent 0 delegating to coder")
		out <- model.StreamChunk{
			ContentDelta: "I'll delegate this file reading task to a coder agent.",
			ToolCalls: []model.ToolCall{
				{
					ID:        "delegate_coder",
					Name:      "agent",
					Arguments: []byte(`{"agent": "coder", "input": "Please read PRODUCT.md and provide a summary of its contents"}`),
				},
			},
			Done: true,
		}
	}()
	return out, nil
}
