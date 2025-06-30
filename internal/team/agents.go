package team

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/router"
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
	routes := router.Rules{{
		Name:       role,
		IfContains: []string{""},
		Client:     model.NewMock(),
	}}
	
	agent := core.New(routes, registry, memory.NewInMemory(), nil, memory.NewInMemoryVector(), nil)
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
