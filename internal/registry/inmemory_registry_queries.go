package registry

import (
	"context"
	"fmt"
)

// GetAgent retrieves information about a specific agent
func (r *InMemoryRegistry) GetAgent(ctx context.Context, agentID string) (*AgentInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	agent, exists := r.agents[agentID]
	if !exists {
		return nil, fmt.Errorf("agent %s not found", agentID)
	}
	
	// Return a copy to prevent external modification
	agentCopy := *agent
	return &agentCopy, nil
}

// ListAllAgents returns all registered agents
func (r *InMemoryRegistry) ListAllAgents(ctx context.Context) ([]*AgentInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	agents := make([]*AgentInfo, 0, len(r.agents))
	for _, agent := range r.agents {
		agentCopy := *agent
		agents = append(agents, &agentCopy)
	}
	
	return agents, nil
}

// FindAgents finds agents with specific capabilities
func (r *InMemoryRegistry) FindAgents(ctx context.Context, capabilities []string) ([]*AgentInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var matchingAgents []*AgentInfo
	
	for _, agent := range r.agents {
		if r.hasCapabilities(agent.Capabilities, capabilities) {
			agentCopy := *agent
			matchingAgents = append(matchingAgents, &agentCopy)
		}
	}
	
	return matchingAgents, nil
}
