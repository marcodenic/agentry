package persistent

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/marcodenic/agentry/internal/converse"
	"github.com/marcodenic/agentry/internal/registry"
	"github.com/marcodenic/agentry/internal/sessions"
)

// SpawnAgent creates a new persistent agent with the given ID and role
func (pt *PersistentTeam) SpawnAgent(ctx context.Context, agentID, role string) (*PersistentAgent, error) {
	pt.mutex.Lock()
	defer pt.mutex.Unlock()

	// Check if agent already exists
	if existing, exists := pt.agents[agentID]; exists {
		if existing.Status == registry.StatusRunning {
			return existing, nil
		}
		// Clean up old agent
		pt.stopAgent(existing)
	}

	// Find available port
	port, err := pt.findAvailablePort()
	if err != nil {
		return nil, fmt.Errorf("no available ports: %w", err)
	}

	// Create team context from existing converse system
	team, err := converse.NewTeamContext(pt.parent)
	if err != nil {
		return nil, fmt.Errorf("failed to create team context: %w", err)
	}

	// Add agent to team using existing system
	agent, _ := team.AddAgent(agentID)

	// Create session-aware agent wrapper
	sessionAgent := sessions.NewSessionAgent(agent, pt.sessionManager)

	// Create persistent agent wrapper
	persistentAgent := &PersistentAgent{
		ID:           agentID,
		Agent:        agent,
		SessionAgent: sessionAgent,
		Port:         port,
		PID:          os.Getpid(), // For now, same process
		Status:       registry.StatusStarting,
		StartedAt:    time.Now(),
		LastSeen:     time.Now(),
		Role:         role,
	}

	// Start HTTP server for agent communication
	if err := pt.startAgentServer(persistentAgent); err != nil {
		return nil, fmt.Errorf("failed to start agent server: %w", err)
	}

	// Register with registry
	agentInfo := &registry.AgentInfo{
		ID:           agentID,
		Port:         port,
		PID:          os.Getpid(),
		Capabilities: []string{role}, // Use role as capability
		Endpoint:     fmt.Sprintf("localhost:%d", port),
		Status:       registry.StatusRunning,
		RegisteredAt: time.Now(),
		LastSeen:     time.Now(),
		Metadata:     map[string]string{"role": role, "spawned_by": "persistent_team"},
	}

	if err := pt.registry.RegisterAgent(ctx, agentInfo); err != nil {
		pt.stopAgent(persistentAgent)
		return nil, fmt.Errorf("failed to register agent: %w", err)
	}

	pt.agents[agentID] = persistentAgent
	persistentAgent.Status = registry.StatusRunning

	return persistentAgent, nil
}

// GetAgent returns an existing persistent agent by ID
func (pt *PersistentTeam) GetAgent(agentID string) (*PersistentAgent, bool) {
	pt.mutex.RLock()
	defer pt.mutex.RUnlock()
	
	agent, exists := pt.agents[agentID]
	return agent, exists
}

// ListAgents returns all currently managed persistent agents
func (pt *PersistentTeam) ListAgents() []*PersistentAgent {
	pt.mutex.RLock()
	defer pt.mutex.RUnlock()

	agents := make([]*PersistentAgent, 0, len(pt.agents))
	for _, agent := range pt.agents {
		agents = append(agents, agent)
	}
	return agents
}

// StopAgent stops a persistent agent and cleans up resources
func (pt *PersistentTeam) StopAgent(ctx context.Context, agentID string) error {
	pt.mutex.Lock()
	defer pt.mutex.Unlock()

	agent, exists := pt.agents[agentID]
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	pt.stopAgent(agent)
	delete(pt.agents, agentID)

	// Deregister from registry
	return pt.registry.DeregisterAgent(ctx, agentID)
}
