package registry

import (
	"context"
	"fmt"
)

// FindBestAgents finds the best agents for a given set of requirements
func (d *DiscoveryService) FindBestAgents(ctx context.Context, opts *DiscoveryOptions) ([]*AgentScore, error) {
	if opts == nil {
		opts = &DiscoveryOptions{}
	}
	
	// Set defaults
	if opts.MaxResults == 0 {
		opts.MaxResults = 10
	}
	if opts.SortBy == "" {
		opts.SortBy = "score"
	}
	if len(opts.RequiredStatus) == 0 {
		opts.RequiredStatus = []AgentStatus{StatusIdle, StatusBusy}
	}
	
	// Get all agents
	allAgents, err := d.registry.ListAllAgents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}
	
	var candidates []*AgentScore
	
	for _, agent := range allAgents {
		// Skip if agent is excluded
		if d.isExcluded(agent.ID, opts.ExcludeAgents) {
			continue
		}
		
		// Skip if agent doesn't have required status
		if !d.hasRequiredStatus(agent.Status, opts.RequiredStatus) {
			continue
		}
		
		// Skip if agent doesn't have required capabilities
		if !d.hasAllCapabilities(agent.Capabilities, opts.RequiredCapabilities) {
			continue
		}
		
		// Calculate score
		score := d.calculateAgentScore(ctx, agent, opts)
		candidates = append(candidates, &AgentScore{
			Agent: agent,
			Score: score,
		})
	}
	
	// Sort candidates
	d.sortCandidates(candidates, opts.SortBy)
	
	// Limit results
	if len(candidates) > opts.MaxResults {
		candidates = candidates[:opts.MaxResults]
	}
	
	return candidates, nil
}

// FindAvailableAgent finds a single available agent with required capabilities
func (d *DiscoveryService) FindAvailableAgent(ctx context.Context, capabilities []string) (*AgentInfo, error) {
	opts := &DiscoveryOptions{
		RequiredCapabilities: capabilities,
		RequiredStatus:       []AgentStatus{StatusIdle},
		MaxResults:           1,
		SortBy:               "score",
	}
	
	candidates, err := d.FindBestAgents(ctx, opts)
	if err != nil {
		return nil, err
	}
	
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no available agents found with capabilities: %v", capabilities)
	}
	
	return candidates[0].Agent, nil
}
