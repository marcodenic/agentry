package team

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/tool"
)

// Team manages a multi-agent conversation step by step.
// This is a simplified version that consolidates the functionality
// from converse.Team and maintains compatibility.
type Team struct {
	parent       *core.Agent
	agents       map[string]*Agent // Changed to use Agent type
	agentsByName map[string]*Agent // Changed to use Agent type
	names        []string
	tasks        map[string]*Task
	messages     []Message
	roles        map[string]*RoleConfig
	portRange    PortRange
	name         string
	msg          string
	turn         int
	maxTurns     int
	mutex        sync.RWMutex
}

// NewTeam creates a new team with the given parent agent.
func NewTeam(parent *core.Agent, maxTurns int, name string) (*Team, error) {
	return &Team{
		parent:       parent,
		maxTurns:     maxTurns,
		name:         name,
		agents:       make(map[string]*Agent),
		agentsByName: make(map[string]*Agent),
		tasks:        make(map[string]*Task),
		messages:     make([]Message, 0),
		roles:        make(map[string]*RoleConfig),
		portRange:    PortRange{Start: 9000, End: 9099},
	}, nil
}

// Add registers ag under name so it can be addressed via Call.
func (t *Team) Add(name string, ag *core.Agent) {
	// CRITICAL: Remove the "agent" tool from added agents to prevent delegation cascading
	if _, hasAgent := ag.Tools["agent"]; hasAgent {
		// Create a new registry without the agent tool
		newTools := make(tool.Registry)
		for toolName, toolInstance := range ag.Tools {
			if toolName != "agent" {
				newTools[toolName] = toolInstance
			}
		}
		ag.Tools = newTools

	}

	// Create wrapper
	agent := &Agent{
		ID:        name,
		Name:      name,
		Agent:     ag,
		Status:    "ready",
		StartedAt: time.Now(),
		LastSeen:  time.Now(),
		Metadata:  make(map[string]string),
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.agents[name] = agent
	t.agentsByName[name] = agent
	t.names = append(t.names, name)
}

// AddAgent creates a new agent and adds it to the team.
// Returns the core agent and its assigned name.
func (t *Team) AddAgent(name string) (*core.Agent, string) {
	// Create a new agent by spawning from the parent
	coreAgent := t.parent.Spawn()

	// CRITICAL: Remove the "agent" tool from spawned agents to prevent delegation cascading
	// but keep other essential tools so they can actually complete tasks
	if _, hasAgent := coreAgent.Tools["agent"]; hasAgent {
		// Create a new registry without the agent tool
		newTools := make(tool.Registry)
		for toolName, toolInstance := range coreAgent.Tools {
			if toolName != "agent" {
				newTools[toolName] = toolInstance
			}
		}
		coreAgent.Tools = newTools

	}

	// Create wrapper
	agent := &Agent{
		ID:        name, // Use name as ID for simplicity
		Name:      name,
		Agent:     coreAgent,
		Status:    "ready",
		StartedAt: time.Now(),
		LastSeen:  time.Now(),
		Metadata:  make(map[string]string),
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.agents[name] = agent
	t.agentsByName[name] = agent
	t.names = append(t.names, name)

	return coreAgent, name
}

// Call implements the Caller interface for compatibility with existing code.
// It delegates work to the named agent.
func (t *Team) Call(ctx context.Context, agentID, input string) (string, error) {
	t.mutex.RLock()
	agent, exists := t.agentsByName[agentID]
	t.mutex.RUnlock()

	if !exists {
		// If agent doesn't exist, create it
		_, _ = t.AddAgent(agentID) // Create agent and ignore return values

		t.mutex.RLock()
		agent = t.agentsByName[agentID] // Get the Agent wrapper
		t.mutex.RUnlock()
	}

	// Execute the input on the core agent using the same pattern as converse.runAgent
	return runAgent(ctx, agent.Agent, input, agentID, t.names)
}

// runAgent executes an agent with the given input, similar to converse.runAgent
func runAgent(ctx context.Context, ag *core.Agent, input, name string, peers []string) (string, error) {
	client, _ := ag.Route.Select(input)
	msgs := core.BuildMessages(ag.Prompt, ag.Vars, ag.Mem.History(), input)
	specs := tool.BuildSpecs(ag.Tools)
	limit := ag.MaxIterations
	if limit <= 0 {
		limit = 8 // Default for agents without explicit limit
	}
	// Special case: if MaxIterations is set to -1, allow unlimited iterations
	unlimited := ag.MaxIterations == -1

	for i := 0; unlimited || i < limit; i++ {
		res, err := client.Complete(ctx, msgs, specs)
		if err != nil {
			return "", fmt.Errorf("agent '%s' completion failed on iteration %d: %w", name, i+1, err)
		}
		msgs = append(msgs, model.ChatMessage{Role: "assistant", Content: res.Content, ToolCalls: res.ToolCalls})
		step := memory.Step{Output: res.Content, ToolCalls: res.ToolCalls, ToolResults: map[string]string{}}
		if len(res.ToolCalls) == 0 {
			ag.Mem.AddStep(step)
			return res.Content, nil
		}
		for _, tc := range res.ToolCalls {
			t, ok := ag.Tools.Use(tc.Name)
			if !ok {
				return "", fmt.Errorf("agent '%s' tried to use unknown tool '%s' on iteration %d", name, tc.Name, i+1)
			}
			var args map[string]any
			if err := json.Unmarshal(tc.Arguments, &args); err != nil {
				return "", fmt.Errorf("agent '%s' tool '%s' has invalid arguments on iteration %d: %w", name, tc.Name, i+1, err)
			}
			r, err := t.Execute(ctx, args)
			if err != nil {
				return "", fmt.Errorf("agent '%s' tool '%s' execution failed on iteration %d with args %v: %w", name, tc.Name, i+1, args, err)
			}
			step.ToolResults[tc.ID] = r
			msgs = append(msgs, model.ChatMessage{Role: "tool", ToolCallID: tc.ID, Content: r})
		}
		ag.Mem.AddStep(step)
	}
	return "", fmt.Errorf("agent '%s' exceeded maximum iterations (%d)", name, limit)
}

// GetAgents returns a list of all agent names in the team.
func (t *Team) GetAgents() []string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return append([]string(nil), t.names...)
}

// Agents returns a list of all core agents in the team.
func (t *Team) Agents() []*core.Agent {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	agents := make([]*core.Agent, 0, len(t.agents))
	for _, agent := range t.agents {
		agents = append(agents, agent.Agent)
	}
	return agents
}

// Names returns a list of all agent names in the team.
func (t *Team) Names() []string {
	return t.GetAgents()
}
