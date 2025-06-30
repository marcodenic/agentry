package hubgrpc

// NOTE: Utility functions are commented out because the corresponding
// protobuf message types (AgentInfo, ScoredAgent, etc.) are not yet implemented in the API.

// TODO: Implement these when the required protobuf messages are added to api/agent.proto

/*
import (
	"github.com/marcodenic/agentry/api"
	"github.com/marcodenic/agentry/internal/registry"
)

// convertAgentToAPI converts a registry.Agent to an API Agent
func convertAgentToAPI(agent *registry.Agent) *api.AgentInfo {
	return &api.AgentInfo{
		Id:           agent.ID,
		Name:         agent.Name,
		Status:       string(agent.Status),
		Capabilities: agent.Capabilities,
		Endpoint:     agent.Endpoint,
		Region:       agent.Region,
		Metadata:     agent.Metadata,
		LastSeen:     timestamppb.New(agent.LastSeen),
		Health:       convertHealthToAPI(agent.Health),
	}
}

// convertHealthToAPI converts registry.AgentHealth to API format
func convertHealthToAPI(health *registry.AgentHealth) *api.HealthMetrics {
	if health == nil {
		return nil
	}
	
	return &api.HealthMetrics{
		CpuUsage:      health.CPUUsage,
		MemoryUsage:   health.MemoryUsage,
		DiskUsage:     health.DiskUsage,
		ActiveTasks:   int32(health.ActiveTasks),
		ErrorRate:     health.ErrorRate,
		ResponseTime:  health.ResponseTime,
		LastCheckTime: timestamppb.New(health.LastCheckTime),
	}
}

// convertScoredAgentToAPI converts a registry.ScoredAgent to API format
func convertScoredAgentToAPI(scored *registry.ScoredAgent) *api.ScoredAgent {
	return &api.ScoredAgent{
		Agent: convertAgentToAPI(scored.Agent),
		Score: scored.Score,
		Reasons: scored.Reasons,
	}
}

// filterAgentsByStatus filters agents by their status
func filterAgentsByStatus(agents []*registry.Agent, status registry.AgentStatus) []*registry.Agent {
	filtered := make([]*registry.Agent, 0, len(agents))
	for _, agent := range agents {
		if agent.Status == status {
			filtered = append(filtered, agent)
		}
	}
	return filtered
}

// filterAgentsByCapabilities filters agents that have all required capabilities
func filterAgentsByCapabilities(agents []*registry.Agent, capabilities []string) []*registry.Agent {
	if len(capabilities) == 0 {
		return agents
	}
	
	filtered := make([]*registry.Agent, 0, len(agents))
	for _, agent := range agents {
		hasAll := true
		for _, required := range capabilities {
			found := false
			for _, agentCap := range agent.Capabilities {
				if agentCap == required {
					found = true
					break
				}
			}
			if !found {
				hasAll = false
				break
			}
		}
		if hasAll {
			filtered = append(filtered, agent)
		}
	}
	return filtered
}

// filterAgentsByRegion filters agents by their region
func filterAgentsByRegion(agents []*registry.Agent, region string) []*registry.Agent {
	if region == "" {
		return agents
	}
	
	filtered := make([]*registry.Agent, 0, len(agents))
	for _, agent := range agents {
		if agent.Region == region {
			filtered = append(filtered, agent)
		}
	}
	return filtered
}
*/
