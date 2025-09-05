package tests

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/team"
	"github.com/marcodenic/agentry/internal/tool"
)

// Mock client that tries to use file tools step by step
type coderDebugClient struct {
	callCount int
	t         *testing.T
}

func (m *coderDebugClient) Stream(ctx context.Context, msgs []model.ChatMessage, tools []model.ToolSpec) (<-chan model.StreamChunk, error) {
	out := make(chan model.StreamChunk, 1)
	go func() {
		defer close(out)
		m.callCount++
		m.t.Logf("Mock client call #%d", m.callCount)

		// Log available tools
		m.t.Logf("Available tools: %d", len(tools))
		for i, tool := range tools {
			m.t.Logf("  Tool %d: %s - %s", i, tool.Name, tool.Description)
		}

		switch m.callCount {
		case 1:
			// First try: List directory to see if PRODUCT.md exists
			m.t.Logf("Attempting to list directory")
			out <- model.StreamChunk{
				ContentDelta: "Let me first check if PRODUCT.md exists in the current directory.",
				ToolCalls: []model.ToolCall{
					{
						ID:        "call_ls",
						Name:      "ls",
						Arguments: []byte(`{}`),
					},
				},
				Done: true,
			}
		case 2:
			// Check if we got tool results from ls
			hasLsResult := false
			for _, msg := range msgs {
				if msg.Role == "tool" {
					hasLsResult = true
					m.t.Logf("Got tool result: %s", msg.Content)
					break
				}
			}

			if hasLsResult {
				// Second try: Read the file directly
				m.t.Logf("Attempting to view PRODUCT.md")
				out <- model.StreamChunk{
					ContentDelta: "Now let me read the PRODUCT.md file.",
					ToolCalls: []model.ToolCall{
						{
							ID:        "call_view",
							Name:      "view",
							Arguments: []byte(`{"path": "PRODUCT.md"}`),
						},
					},
					Done: true,
				}
			} else {
				m.t.Logf("No ls result received, returning error message")
				out <- model.StreamChunk{
					ContentDelta: "Error: Could not list directory contents.",
					Done:         true,
				}
			}
		case 3:
			// Check if we got the file content
			hasFileContent := false
			for _, msg := range msgs {
				if msg.Role == "tool" && msg.ToolCallID == "call_view" {
					hasFileContent = true
					m.t.Logf("Got file content: %.200s...", msg.Content)
					break
				}
			}

			if hasFileContent {
				out <- model.StreamChunk{
					ContentDelta: "Successfully read PRODUCT.md file content.",
					Done:         true,
				}
			} else {
				m.t.Logf("No file content received")
				out <- model.StreamChunk{
					ContentDelta: "Error: Could not read PRODUCT.md file.",
					Done:         true,
				}
			}
		default:
			out <- model.StreamChunk{
				ContentDelta: "Debug session completed.",
				Done:         true,
			}
		}
	}()
	return out, nil
}

func TestCoderFileReadingDebug(t *testing.T) {
	// Create a team like the real application does
	registry := tool.DefaultRegistry()
	mockClient := &coderDebugClient{t: t}

	// Create parent agent (Agent 0)
	ag := core.New(mockClient, "mock", registry, memory.NewInMemory(), memory.NewInMemoryVector(), nil)
	ag.Prompt = "You are Agent 0"

	// Create team
	tm, err := team.NewTeam(ag, 1, "test")
	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}

	// Register the agent tool
	tm.RegisterAgentTool(ag.Tools)
	ctx := team.WithContext(context.Background(), tm)

	t.Logf("Testing coder agent creation and file reading...")

	// Try to spawn a coder agent and have it read PRODUCT.md
	tl, ok := registry.Use("agent")
	if !ok {
		t.Fatal("agent tool missing")
	}

	// This should create a coder agent and try to read PRODUCT.md
	output, err := tl.Execute(ctx, map[string]any{
		"agent": "coder",
		"input": "Please read the file PRODUCT.md and tell me what it contains",
	})

	t.Logf("Delegation output: %s", output)
	if err != nil {
		t.Logf("Delegation error: %v", err)
	}

	// Log team status
	agents := tm.GetTeamAgents()
	t.Logf("Team has %d agents after delegation", len(agents))
	for _, agent := range agents {
		t.Logf("  Agent: %s (status: %s)", agent.Name, agent.Status)
	}
}

func TestFileToolsDirectly(t *testing.T) {
	// Change to the agentry root directory so PRODUCT.md can be found
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	agentryRoot := filepath.Join(originalWd, "..")
	err := os.Chdir(agentryRoot)
	if err != nil {
		t.Fatalf("Failed to change to agentry root: %v", err)
	}

	// Test file tools directly to make sure they work
	registry := tool.DefaultRegistry()

	t.Logf("Testing ls tool directly...")
	lsTool, ok := registry.Use("ls")
	if !ok {
		t.Fatal("ls tool missing from registry")
	}

	lsOutput, err := lsTool.Execute(context.Background(), map[string]any{})
	if err != nil {
		t.Fatalf("ls tool failed: %v", err)
	}
	t.Logf("ls output: %s", lsOutput)

	t.Logf("Testing view tool directly...")
	viewTool, ok := registry.Use("view")
	if !ok {
		t.Fatal("view tool missing from registry")
	}

	viewOutput, err := viewTool.Execute(context.Background(), map[string]any{"path": "PRODUCT.md"})
	if err != nil {
		t.Fatalf("view tool failed: %v", err)
	}
	t.Logf("view output length: %d characters", len(viewOutput))
	t.Logf("view output preview: %.200s...", viewOutput)
}
