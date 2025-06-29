package registry

import (
	"context"
	"fmt"
	"sort"
	"time"
)

// DiscoveryService provides intelligent agent discovery capabilities
type DiscoveryService struct {
	registry AgentRegistry
}

// NewDiscoveryService creates a new discovery service
func NewDiscoveryService(registry AgentRegistry) *DiscoveryService {
	return &DiscoveryService{
		registry: registry,
	}
}

// AgentScore represents an agent's suitability score for a task
type AgentScore struct {
	Agent *AgentInfo `json:"agent"`
	Score float64    `json:"score"`
}

// DiscoveryOptions contains options for agent discovery
type DiscoveryOptions struct {
	RequiredCapabilities []string  `json:"required_capabilities"`
	PreferredCapabilities []string  `json:"preferred_capabilities"`
	ExcludeAgents        []string  `json:"exclude_agents"`
	RequiredStatus       []AgentStatus `json:"required_status"`
	MaxResults           int       `json:"max_results"`
	SortBy               string    `json:"sort_by"` // "score", "load", "uptime"
}

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

// ClusterStatus represents the overall status of the agent cluster
type ClusterStatus struct {
	TotalAgents          int                    `json:"total_agents"`
	StatusCounts         map[AgentStatus]int    `json:"status_counts"`
	Capabilities         map[string]int         `json:"capabilities"`
	Roles               map[string]int         `json:"roles"`
	AverageUptime       time.Duration          `json:"average_uptime"`
	TotalTasksCompleted int64                  `json:"total_tasks_completed"`
	TotalErrors         int64                  `json:"total_errors"`
	LastUpdated         time.Time              `json:"last_updated"`
}

// Helper methods

func (d *DiscoveryService) isExcluded(agentID string, excludeList []string) bool {
	for _, excluded := range excludeList {
		if excluded == agentID {
			return true
		}
	}
	return false
}

func (d *DiscoveryService) hasRequiredStatus(agentStatus AgentStatus, requiredStatuses []AgentStatus) bool {
	for _, required := range requiredStatuses {
		if agentStatus == required {
			return true
		}
	}
	return false
}

func (d *DiscoveryService) hasAllCapabilities(agentCaps, requiredCaps []string) bool {
	if len(requiredCaps) == 0 {
		return true
	}
	
	capMap := make(map[string]bool)
	for _, cap := range agentCaps {
		capMap[cap] = true
	}
	
	for _, required := range requiredCaps {
		if !capMap[required] {
			return false
		}
	}
	
	return true
}

func (d *DiscoveryService) calculateAgentScore(ctx context.Context, agent *AgentInfo, opts *DiscoveryOptions) float64 {
	score := 0.0
	
	// Base score for having required capabilities
	score += 10.0
	
	// Bonus for preferred capabilities
	for _, preferred := range opts.PreferredCapabilities {
		for _, agentCap := range agent.Capabilities {
			if agentCap == preferred {
				score += 5.0
				break
			}
		}
	}
	
	// Bonus for idle status (prefer less busy agents)
	switch agent.Status {
	case StatusIdle:
		score += 15.0
	case StatusBusy:
		score += 5.0
	default:
		score -= 10.0
	}
	
	// Bonus for recent activity (prefer responsive agents)
	timeSinceLastSeen := time.Since(agent.LastSeen)
	if timeSinceLastSeen < time.Minute {
		score += 10.0
	} else if timeSinceLastSeen < 5*time.Minute {
		score += 5.0
	}
	
	// Factor in health metrics if available
	if health, err := d.registry.GetAgentHealth(ctx, agent.ID); err == nil {
		// Prefer agents with fewer active tasks
		score -= float64(health.TasksActive) * 2.0
		
		// Prefer agents with lower error rates
		if health.TasksCompleted > 0 {
			errorRate := float64(health.ErrorCount) / float64(health.TasksCompleted)
			score -= errorRate * 20.0
		}
		
		// Prefer agents with longer uptime (more stable)
		if health.Uptime > 3600 { // More than 1 hour
			score += 5.0
		}
	}
	
	return score
}

func (d *DiscoveryService) sortCandidates(candidates []*AgentScore, sortBy string) {
	switch sortBy {
	case "score":
		sort.Slice(candidates, func(i, j int) bool {
			return candidates[i].Score > candidates[j].Score
		})
	case "load":
		// Sort by number of active tasks (ascending)
		sort.Slice(candidates, func(i, j int) bool {
			healthI, errI := d.registry.GetAgentHealth(context.Background(), candidates[i].Agent.ID)
			healthJ, errJ := d.registry.GetAgentHealth(context.Background(), candidates[j].Agent.ID)
			
			if errI != nil || errJ != nil {
				return false
			}
			
			return healthI.TasksActive < healthJ.TasksActive
		})
	case "uptime":
		// Sort by uptime (descending)
		sort.Slice(candidates, func(i, j int) bool {
			healthI, errI := d.registry.GetAgentHealth(context.Background(), candidates[i].Agent.ID)
			healthJ, errJ := d.registry.GetAgentHealth(context.Background(), candidates[j].Agent.ID)
			
			if errI != nil || errJ != nil {
				return false
			}
			
			return healthI.Uptime > healthJ.Uptime
		})
	default:
		// Default to score-based sorting
		sort.Slice(candidates, func(i, j int) bool {
			return candidates[i].Score > candidates[j].Score
		})
	}
}
