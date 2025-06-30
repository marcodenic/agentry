package registry

import (
	"context"
	"fmt"
)

// GetAgent retrieves information about a specific agent
func (r *FileRegistry) GetAgent(ctx context.Context, agentID string) (*AgentInfo, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	agent, exists := r.agents[agentID]
	if !exists {
		return nil, fmt.Errorf("agent %s not found", agentID)
	}

	// Return a copy to prevent external modification
	agentCopy := *agent
	return &agentCopy, nil
}

// ListAllAgents returns all registered agents
func (r *FileRegistry) ListAllAgents(ctx context.Context) ([]*AgentInfo, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	agents := make([]*AgentInfo, 0, len(r.agents))
	for _, agent := range r.agents {
		agentCopy := *agent
		agents = append(agents, &agentCopy)
	}

	return agents, nil
}

// FindAgents finds agents with specific capabilities
func (r *FileRegistry) FindAgents(ctx context.Context, capabilities []string) ([]*AgentInfo, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var matchingAgents []*AgentInfo

	for _, agent := range r.agents {
		if r.hasCapabilities(agent, capabilities) {
			agentCopy := *agent
			matchingAgents = append(matchingAgents, &agentCopy)
		}
	}

	return matchingAgents, nil
}
