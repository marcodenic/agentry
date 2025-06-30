package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/marcodenic/agentry/internal/config"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/internal/team"
	"github.com/marcodenic/agentry/internal/tool"
	"github.com/marcodenic/agentry/internal/trace"
	"github.com/marcodenic/agentry/pkg/memstore"
)

func TestAgentSpawningWithToolRestriction(t *testing.T) {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Set up Agent 0 with tool registry
	reg := tool.Registry{
		"agent": tool.New("agent", "Delegate work to another agent", func(ctx context.Context, args map[string]any) (string, error) {
			agentName, ok := args["agent"].(string)
			if !ok {
				return "", fmt.Errorf("agent name is required")
			}
			input, ok := args["input"].(string)
			if !ok {
				return "", fmt.Errorf("input is required")
			}
			
			// Get the team from context
			teamInstance := team.TeamFromContext(ctx)
			if teamInstance == nil {
				return "", fmt.Errorf("no team found in context")
			}
			
			// Add the requested agent
			ag, name := teamInstance.AddAgent(agentName)
			
			// Execute the task
			msg, err := ag.Execute(ctx, input)
			return fmt.Sprintf("Agent %s: %s", name, msg), err
		}),
		"bash": tool.New("bash", "Execute bash commands", func(ctx context.Context, args map[string]any) (string, error) {
			return "bash command executed", nil
		}),
		"view": tool.New("view", "View files", func(ctx context.Context, args map[string]any) (string, error) {
			return "file viewed", nil
		}),
		"write": tool.New("write", "Write files", func(ctx context.Context, args map[string]any) (string, error) {
			return "file written", nil
		}),
		"edit": tool.New("edit", "Edit files", func(ctx context.Context, args map[string]any) (string, error) {
			return "file edited", nil
		}),
		"patch": tool.New("patch", "Apply patches", func(ctx context.Context, args map[string]any) (string, error) {
			return "patch applied", nil
		}),
		"fetch": tool.New("fetch", "Fetch data", func(ctx context.Context, args map[string]any) (string, error) {
			return "data fetched", nil
		}),
		"ls": tool.New("ls", "List files", func(ctx context.Context, args map[string]any) (string, error) {
			return "files listed", nil
		}),
	}

	// Create Agent 0
	route := router.NewOpenAI(model.FromConfig(cfg))
	mem := memory.NewInMemory()
	store := memstore.NewInMemory()
	vec := memory.NewInMemoryVector()
	tracer := trace.NewNoop()

	agent0 := core.New(route, reg, mem, store, vec, tracer)
	
	// Set Agent 0 prompt (this should be loaded from agent_0.yaml in real usage)
	agent0.Prompt = `You are Agent 0, the system orchestrator. You can delegate tasks using the "agent" tool.
Available agent types: coder, researcher, analyst.
When a user asks for coding help, delegate to the "coder" agent.`

	// Create team with Agent 0
	team := converse.New(agent0, 1, "Test spawning agents", 10)

	// Create context with team
	ctx := context.Background()

	fmt.Printf("🧪 TEST: Starting agent spawning test\n")
	fmt.Printf("🧪 TEST: Agent 0 has %d tools: ", len(agent0.Tools))
	for name := range agent0.Tools {
		fmt.Printf("%s ", name)
	}
	fmt.Printf("\n")

	// Test message that should trigger agent spawning
	testMessage := "I need help writing a Python script to calculate fibonacci numbers"

	// Create a timeout context to prevent infinite loops
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	fmt.Printf("🧪 TEST: Executing Agent 0 with message: %s\n", testMessage)
	response, err := agent0.Execute(ctx, testMessage)
	
	if err != nil {
		if err == context.DeadlineExceeded {
			t.Fatalf("Test timed out - likely infinite recursion occurred")
		}
		t.Fatalf("Agent execution failed: %v", err)
	}

	fmt.Printf("🧪 TEST: Agent 0 response: %s\n", response)
	
	// Check if agent was spawned successfully
	if len(team.GetNames()) > 1 {
		fmt.Printf("🧪 TEST: SUCCESS - Team now has %d agents: %v\n", len(team.GetNames()), team.GetNames())
		
		// Check if the spawned agent has restricted tools
		coderAgent := team.GetAgent("coder")
		if coderAgent != nil {
			fmt.Printf("🧪 TEST: Coder agent has %d tools: ", len(coderAgent.Tools))
			for name := range coderAgent.Tools {
				fmt.Printf("%s ", name)
			}
			fmt.Printf("\n")
			
			// Verify that coder doesn't have the "agent" tool
			if _, hasAgentTool := coderAgent.Tools["agent"]; hasAgentTool {
				t.Errorf("ERROR: Coder agent should not have access to 'agent' tool")
			} else {
				fmt.Printf("🧪 TEST: SUCCESS - Coder agent correctly does not have 'agent' tool\n")
			}
		}
	} else {
		t.Errorf("ERROR: No agents were spawned")
	}
}

func main() {
	testing.Main(func(pat, str string) (bool, error) { return true, nil },
		[]testing.InternalTest{
			{"TestAgentSpawningWithToolRestriction", TestAgentSpawningWithToolRestriction},
		},
		nil, nil)
}
