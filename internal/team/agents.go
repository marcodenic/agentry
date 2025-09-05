package team

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/env"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/tool"
)

// AddExistingAgent adds an existing agent to the team
func (t *Team) AddExistingAgent(name string, agent *core.Agent) error {
	// Consolidate with Add to ensure consistent sanitization (e.g., removing the "agent" tool)
	t.Add(name, agent)
	return nil
}

// SpawnAgent creates a new agent with the given configuration
func (t *Team) SpawnAgent(ctx context.Context, name, role string) (*Agent, error) {
	timer := StartTimer(fmt.Sprintf("SpawnAgent(%s, %s)", name, role))
	defer timer.Stop()

	t.mutex.Lock()
	defer t.mutex.Unlock()

	timer.Checkpoint("mutex acquired")

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

	timer.Checkpoint("role config resolved")

	// Create the core agent
	registry := tool.DefaultRegistry()
	// Apply curated defaults and cap tool schemas to keep context small
	maxTools := env.Int("AGENTRY_MAX_TOOLS", 5)
	curated := curatedToolsForRole(role)
	if len(curated) > 0 {
		// Don't cap tools for roles that need comprehensive toolsets (like coder)
		roleName := strings.ToLower(strings.TrimSpace(role))
		if roleName == "coder" {
			// Give coder all the tools it needs - no artificial cap
			registry = filterRegistryByNames(registry, curated, 0) // 0 = no cap
		} else {
			registry = filterRegistryByNames(registry, curated, maxTools)
		}
	} else if maxTools > 0 {
		registry = capRegistry(registry, maxTools)
	}

	// Apply tool restrictions based on role configuration
	if len(roleConfig.RestrictedTools) > 0 {
		for _, restrictedTool := range roleConfig.RestrictedTools {
			delete(registry, restrictedTool)
		}
		if !isTUI() {
			fmt.Fprintf(os.Stderr, "🚫 SpawnAgent: Restricted %d tools for role %s: %v\n",
				len(roleConfig.RestrictedTools), role, roleConfig.RestrictedTools)
		}
	}

	// Create proper model client based on role configuration
	var client model.Client
	var modelName string

	if roleConfig.Model != nil {
		// Use role-specific model configuration
		if !isTUI() {
			fmt.Fprintf(os.Stderr, "🔧 SpawnAgent: Attempting to create model client for role %s with provider %s\n", role, roleConfig.Model.Provider)
		}
		c, err := model.FromManifest(*roleConfig.Model)
		if err != nil {
			if !isTUI() {
				fmt.Fprintf(os.Stderr, "❌ SpawnAgent: failed to create model client for role %s: %v, falling back to mock\n", role, err)
			}
			client = model.NewMock()
			modelName = "mock"
		} else {
			client = c
			modelName = fmt.Sprintf("%s/%s", roleConfig.Model.Provider, roleConfig.Model.Options["model"])
			if !isTUI() {
				fmt.Fprintf(os.Stderr, "✅ SpawnAgent: Successfully created %s model client for role %s\n", modelName, role)
			}
		}
	} else {
		// Fallback to parent's client when no role config is found
		client = t.parent.Client
		modelName = t.parent.ModelName
		if !isTUI() {
			fmt.Fprintf(os.Stderr, "⚠️  SpawnAgent: No model config for role %s, using parent's client (%s)\n", role, modelName)
		}
	}

	agent := core.New(client, modelName, registry, memory.NewInMemory(), memory.NewInMemoryVector(), nil)
	// Ensure we do not allow recursive delegation by default
	delete(agent.Tools, "agent")
	agent.InvalidateToolCache()
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
// (Removed) GetAgentCount and GetAgentNames: use GetAgents() and ListAgents() instead to avoid duplication
