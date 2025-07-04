package team

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/tool"
)

// AddExistingAgent adds an existing agent to the team
func (t *Team) AddExistingAgent(name string, agent *core.Agent) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	id := uuid.New().String()
	teamAgent := &Agent{
		ID:        id,
		Name:      name,
		Agent:     agent,
		Status:    "ready",
		StartedAt: time.Now(),
		LastSeen:  time.Now(),
		Metadata:  make(map[string]string),
	}

	t.agents[id] = teamAgent
	t.agentsByName[name] = teamAgent

	return nil
}

// SpawnAgent creates a new agent with the given configuration
func (t *Team) SpawnAgent(ctx context.Context, name, role string) (*Agent, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// Get role configuration
	roleConfig := t.roles[role]
	if roleConfig == nil {
		// Create default role if not found
		roleConfig = &RoleConfig{
			Name:         role,
			Prompt:       fmt.Sprintf("You are a %s agent", role),
			Tools:        []string{},
			Capabilities: []string{},
			Metadata:     make(map[string]string),
		}
	}

	// Create the core agent
	registry := tool.DefaultRegistry()

	// Create proper model client based on role configuration
	var client model.Client
	var modelName string

	if roleConfig.Model != nil {
		// Use role-specific model configuration
		if os.Getenv("AGENTRY_TUI_MODE") != "1" {
			fmt.Printf("🔧 SpawnAgent: Attempting to create model client for role %s with provider %s\n", role, roleConfig.Model.Provider)
		}
		c, err := model.FromManifest(*roleConfig.Model)
		if err != nil {
			if os.Getenv("AGENTRY_TUI_MODE") != "1" {
				fmt.Printf("❌ SpawnAgent: failed to create model client for role %s: %v, falling back to mock\n", role, err)
			}
			client = model.NewMock()
			modelName = "mock"
		} else {
			client = c
			modelName = fmt.Sprintf("%s-%s", roleConfig.Model.Provider, roleConfig.Model.Options["model"])
			if os.Getenv("AGENTRY_TUI_MODE") != "1" {
				fmt.Printf("✅ SpawnAgent: Successfully created %s model client for role %s\n", modelName, role)
			}
		}
	} else {
		// Fallback to parent's client when no role config is found
		client = t.parent.Client
		modelName = t.parent.ModelName
		if os.Getenv("AGENTRY_TUI_MODE") != "1" {
			fmt.Printf("⚠️  SpawnAgent: No model config for role %s, using parent's client (%s)\n", role, modelName)
		}
	}

	agent := core.New(client, modelName, registry, memory.NewInMemory(), nil, memory.NewInMemoryVector(), nil)
	agent.Prompt = roleConfig.Prompt

	// Find available port
	port, err := t.findAvailablePort()
	if err != nil {
		return nil, fmt.Errorf("no available ports: %w", err)
	}

	id := uuid.New().String()
	teamAgent := &Agent{
		ID:        id,
		Name:      name,
		Role:      role,
		Agent:     agent,
		Port:      port,
		Status:    "starting",
		StartedAt: time.Now(),
		LastSeen:  time.Now(),
		Metadata:  roleConfig.Metadata,
	}

	t.agents[id] = teamAgent
	t.agentsByName[name] = teamAgent

	// Start the agent
	go func() {
		// Agent is ready
		teamAgent.SetStatus("ready")
	}()

	return teamAgent, nil
}

// GetAgent returns an agent by ID or name
func (t *Team) GetAgent(id string) *Agent {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	// Try ID first
	if agent := t.agents[id]; agent != nil {
		return agent
	}

	// Try name
	if agent := t.agentsByName[id]; agent != nil {
		return agent
	}

	return nil
}

// ListAgents returns all agents in the team
func (t *Team) ListAgents() []*Agent {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	agents := make([]*Agent, 0, len(t.agents))
	for _, agent := range t.agents {
		agents = append(agents, agent)
	}

	return agents
}

// StopAgent stops and removes an agent from the team
func (t *Team) StopAgent(ctx context.Context, agentID string) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	agent := t.agents[agentID]
	if agent == nil {
		return fmt.Errorf("agent %s not found", agentID)
	}

	// Update status
	agent.SetStatus("stopping")

	// Remove from maps
	delete(t.agents, agentID)
	delete(t.agentsByName, agent.Name)

	return nil
}

// GetAgentCount returns the number of agents in the team
func (t *Team) GetAgentCount() int {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return len(t.agents)
}

// GetAgentNames returns the names of all agents
func (t *Team) GetAgentNames() []string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	names := make([]string, 0, len(t.agentsByName))
	for name := range t.agentsByName {
		names = append(names, name)
	}

	return names
}
