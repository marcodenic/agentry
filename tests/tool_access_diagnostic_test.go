package tests

import (
	"context"
	"testing"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/team"
	"github.com/marcodenic/agentry/internal/tool"
)

// Diagnostic test to understand why coder agent can't use ls properly
func TestCoderAgentToolAccess(t *testing.T) {
	// First, check what tools are available in the default registry
	defaultRegistry := tool.DefaultRegistry()
	t.Logf("=== DEFAULT REGISTRY ANALYSIS ===")
	t.Logf("Default registry has %d tools", len(defaultRegistry))

	lsTool, hasLs := defaultRegistry.Use("ls")
	if hasLs {
		t.Logf("✅ ls tool found in default registry")
		t.Logf("ls tool description: %s", lsTool.Description())
	} else {
		t.Logf("❌ ls tool NOT found in default registry")
	}

	// Test ls tool directly from default registry
	if hasLs {
		t.Logf("=== TESTING LS TOOL DIRECTLY ===")
		output, err := lsTool.Execute(context.Background(), map[string]any{})
		if err != nil {
			t.Logf("❌ ls tool execute failed: %v", err)
			t.Logf("Error type: %T", err)
		} else {
			t.Logf("✅ ls tool works directly: %s", output)
		}
	}

	// Now test what happens when we spawn a coder agent
	t.Logf("=== SPAWNING CODER AGENT ===")

	// Create Agent 0
	agent0Registry := tool.DefaultRegistry()
	mockClient := &toolDiagnosticClient{t: t}
	ag := core.New(mockClient, "mock", agent0Registry, memory.NewInMemory(), memory.NewInMemoryVector(), nil)

	// Create team
	tm, err := team.NewTeam(ag, 1, "test")
	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}
	tm.RegisterAgentTool(ag.Tools)
	ctx := team.WithContext(context.Background(), tm)

	t.Logf("Agent 0 has %d tools", len(ag.Tools))

	// Test delegation to coder
	tl, ok := agent0Registry.Use("agent")
	if !ok {
		t.Fatal("agent tool missing")
	}

	output, err := tl.Execute(ctx, map[string]any{
		"agent": "coder",
		"input": "diagnostic test - check your tools",
	})

	t.Logf("Delegation result: %s", output)
	if err != nil {
		t.Logf("Delegation error: %v", err)
	}
}

// Mock client that tries to list and use available tools
type toolDiagnosticClient struct {
	callCount int
	t         *testing.T
}

func (d *toolDiagnosticClient) Stream(ctx context.Context, msgs []model.ChatMessage, tools []model.ToolSpec) (<-chan model.StreamChunk, error) {
	out := make(chan model.StreamChunk, 1)
	go func() {
		defer close(out)
		d.callCount++
		d.t.Logf("=== MOCK CLIENT CALL #%d ===", d.callCount)
		d.t.Logf("Available tools count: %d", len(tools))

		// Log all available tools
		for i, tool := range tools {
			d.t.Logf("Tool %d: %s", i, tool.Name)
		}

		// Check specifically for ls
		hasLs := false
		for _, tool := range tools {
			if tool.Name == "ls" {
				hasLs = true
				break
			}
		}

		if hasLs {
			d.t.Logf("✅ Coder agent HAS ls tool available")

			// Try to use ls tool
			out <- model.StreamChunk{
				ContentDelta: "I have access to ls tool, let me try to use it.",
				ToolCalls: []model.ToolCall{
					{
						ID:        "test_ls",
						Name:      "ls",
						Arguments: []byte(`{}`),
					},
				},
				Done: true,
			}
		} else {
			d.t.Logf("❌ Coder agent does NOT have ls tool")
			out <- model.StreamChunk{
				ContentDelta: "I don't have access to ls tool",
				Done:         true,
			}
		}
	}()
	return out, nil
}

// Test the spawned agent's registry directly
func TestSpawnedAgentRegistry(t *testing.T) {
	t.Logf("=== SPAWNED AGENT REGISTRY TEST ===")

	// Create parent agent
	parentRegistry := tool.DefaultRegistry()
	mockClient := &simpleTestClient{response: "test", t: t}
	parentAgent := core.New(mockClient, "mock", parentRegistry, memory.NewInMemory(), memory.NewInMemoryVector(), nil)

	// Create team
	tm, err := team.NewTeam(parentAgent, 1, "test")
	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}

	// Manually spawn a coder agent to inspect its registry
	spawnedAgent, err := tm.SpawnAgent(context.Background(), "test-coder", "coder")
	if err != nil {
		t.Fatalf("Failed to spawn agent: %v", err)
	}

	t.Logf("Spawned agent has %d tools", len(spawnedAgent.Agent.Tools))

	// Check if spawned agent has ls
	lsTool, hasLs := spawnedAgent.Agent.Tools.Use("ls")
	if hasLs {
		t.Logf("✅ Spawned agent HAS ls tool")

		// Try to execute it
		output, err := lsTool.Execute(context.Background(), map[string]any{})
		if err != nil {
			t.Logf("❌ Spawned agent ls execution failed: %v", err)
		} else {
			t.Logf("✅ Spawned agent ls execution succeeded: %s", output)
		}
	} else {
		t.Logf("❌ Spawned agent does NOT have ls tool")
	}

	// Also check what tools it does have
	t.Logf("Spawned agent tools:")
	for name := range spawnedAgent.Agent.Tools {
		t.Logf("  - %s", name)
	}
}

type simpleTestClient struct {
	response string
	t        *testing.T
}

func (s *simpleTestClient) Stream(ctx context.Context, msgs []model.ChatMessage, tools []model.ToolSpec) (<-chan model.StreamChunk, error) {
	out := make(chan model.StreamChunk, 1)
	go func() {
		defer close(out)
		out <- model.StreamChunk{
			ContentDelta: s.response,
			Done:         true,
		}
	}()
	return out, nil
}

// Test the ls tool implementation directly
func TestLsToolImplementation(t *testing.T) {
	t.Logf("=== LS TOOL IMPLEMENTATION TEST ===")

	// Get ls tool from registry
	registry := tool.DefaultRegistry()
	lsTool, ok := registry.Use("ls")
	if !ok {
		t.Fatal("ls tool not in registry")
	}

	t.Logf("ls tool description: %s", lsTool.Description())

	// Try different argument patterns
	testCases := []map[string]any{
		{},            // no args
		{"path": "."}, // current directory
		{"path": "/home/marco/Documents/GitHub/agentry"}, // full path
	}

	for i, args := range testCases {
		t.Logf("Test case %d: %v", i, args)
		output, err := lsTool.Execute(context.Background(), args)
		if err != nil {
			t.Logf("❌ Failed: %v", err)
		} else {
			t.Logf("✅ Success: %s", output)
			break // Stop on first success
		}
	}
}
