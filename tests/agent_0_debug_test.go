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

// Mock AI client that returns a tool call for "agent" tool with "coder" agent
type mockAgent0Client struct {
	callCount int
}

func (m *mockAgent0Client) Complete(ctx context.Context, msgs []model.ChatMessage, tools []model.ToolSpec) (model.Completion, error) {
	m.callCount++

	if m.callCount == 1 {
		// First call - return a tool call to spawn a coder agent
		return model.Completion{
			Content: "I'll spawn a coder agent to help you with this task.",
			ToolCalls: []model.ToolCall{
				{
					ID:        "call_123",
					Name:      "agent",
					Arguments: []byte(`{"agent": "coder", "input": "help with Python project"}`),
				},
			},
		}, nil
	} else {
		// After tool call, return final response
		return model.Completion{
			Content: "I've successfully delegated the Python project task to a coder agent. The coder agent will help you with your Python development needs.",
		}, nil
	}
}

func TestAgent0DebugOutput(t *testing.T) {
	// Create Agent 0 with the actual prompt from agent_0.yaml
	agent0Prompt := `You are Agent 0, the system orchestrator in a multi-agent environment. Your primary job is
  to analyse user requests and either act directly or delegate work to specialised agents.

  ## Core Responsibilities
  1. **Natural Language Analysis** – Understand each request and determine scope and complexity.
  2. **Agent Management** – Spawn and coordinate sub‑agents when tasks require special skills.
  3. **Team Orchestration** – Create teams for complex workflows and keep them aligned.
  4. **Task Delegation** – Use the ` + "`agent`" + ` tool to assign work to other agents.
  5. **Direct Assistance** – Handle simple questions yourself using available tools.

  ## Decision Framework
  - **Simple tasks** – Handle yourself.
  - **Specialised tasks** – Delegate to the appropriate agent (` + "`coder`" + `, ` + "`researcher`" + `, etc.).
  - **Complex projects** – Form a team and coordinate execution.
  - **Parallel work** – Spawn multiple agents to run at the same time when tasks are independent.

  ## Available Tools for Orchestration
  - ` + "`agent`" + ` – Delegate tasks. Usage example: {"agent": "agent_name", "input": "task"}
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

	// Create agent with mock client
	mockClient := &mockAgent0Client{}
	ag := core.New(mockClient, "mock", tool.DefaultRegistry(), memory.NewInMemory(), memory.NewInMemoryVector(), nil)
	ag.Prompt = agent0Prompt

	// Create a team context so the agent tool can work
	tm, err := team.NewTeam(ag, 1, "test")
	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}

	// Register the agent tool to enable delegation
	tm.RegisterAgentTool(ag.Tools)

	ctx := team.WithContext(context.Background(), tm)

	// Test request that should trigger Agent 0 to spawn a coder
	input := "I need help with a Python project"

	output, err := ag.Run(ctx, input)
	if err != nil {
		t.Fatalf("Agent run failed: %v", err)
	}

	t.Logf("Agent 0 output: %s", output)

	// The output should contain our debug information
	if output == "" {
		t.Fatal("No output received from Agent 0")
	}
}
