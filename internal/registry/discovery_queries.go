package registry

import (
	"context"
	"sort"
	"time"
)

// GetAgentCapabilities returns a list of all unique capabilities across all agents
func (d *DiscoveryService) GetAgentCapabilities(ctx context.Context) ([]string, error) {
	agents, err := d.registry.ListAllAgents(ctx)
	if err != nil {
		return nil, err
	}
	
	capabilitySet := make(map[string]bool)
	for _, agent := range agents {
		for _, cap := range agent.Capabilities {
			capabilitySet[cap] = true
		}
	}
	
	capabilities := make([]string, 0, len(capabilitySet))
	for cap := range capabilitySet {
		capabilities = append(capabilities, cap)
	}
	
	sort.Strings(capabilities)
	return capabilities, nil
}

// GetAgentsByRole returns all agents with a specific role
func (d *DiscoveryService) GetAgentsByRole(ctx context.Context, role string) ([]*AgentInfo, error) {
	allAgents, err := d.registry.ListAllAgents(ctx)
	if err != nil {
		return nil, err
	}
	
	var roleAgents []*AgentInfo
	for _, agent := range allAgents {
		if agent.Role == role {
			roleAgents = append(roleAgents, agent)
		}
	}
	
	return roleAgents, nil
}

// GetClusterStatus returns overall cluster status information
func (d *DiscoveryService) GetClusterStatus(ctx context.Context) (*ClusterStatus, error) {
	agents, err := d.registry.ListAllAgents(ctx)
	if err != nil {
		return nil, err
	}
	
	status := &ClusterStatus{
		TotalAgents:    len(agents),
		StatusCounts:   make(map[AgentStatus]int),
		Capabilities:   make(map[string]int),
		Roles:          make(map[string]int),
		LastUpdated:    time.Now(),
	}
	
	var totalUptime, totalTasks, totalErrors int64
	healthyAgents := 0
	
	for _, agent := range agents {
		// Count by status
		status.StatusCounts[agent.Status]++
		
		// Count capabilities
		for _, cap := range agent.Capabilities {
			status.Capabilities[cap]++
		}
		
		// Count roles
		if agent.Role != "" {
			status.Roles[agent.Role]++
		}
		
		// Aggregate health metrics
		if health, err := d.registry.GetAgentHealth(ctx, agent.ID); err == nil {
			totalUptime += health.Uptime
			totalTasks += health.TasksCompleted
			totalErrors += health.ErrorCount
			healthyAgents++
		}
	}
	
	// Calculate averages
	if healthyAgents > 0 {
		status.AverageUptime = time.Duration(totalUptime / int64(healthyAgents)) * time.Second
		status.TotalTasksCompleted = totalTasks
		status.TotalErrors = totalErrors
	}
	
	return status, nil
}
