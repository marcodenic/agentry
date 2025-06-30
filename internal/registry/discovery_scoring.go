package registry

import (
	"context"
	"sort"
	"time"
)

// calculateAgentScore computes a suitability score for an agent
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

// sortCandidates sorts agent candidates by the specified criteria
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
