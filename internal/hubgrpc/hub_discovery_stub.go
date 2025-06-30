package hubgrpc

// NOTE: Discovery methods are commented out because the corresponding
// protobuf message types (FindAgentsRequest, FindAgentsResponse, etc.) 
// are not yet implemented in the API.

// TODO: Implement these when the required protobuf messages are added to api/agent.proto

/*
import (
	"context"

	"github.com/marcodenic/agentry/api"
	"github.com/marcodenic/agentry/internal/registry"
)

// FindAgents finds the best agents for a given set of requirements
func (h *Server) FindAgents(ctx context.Context, req *api.FindAgentsRequest) (*api.FindAgentsResponse, error) {
	opts := &registry.DiscoveryOptions{
		RequiredCapabilities:  req.RequiredCapabilities,
		PreferredCapabilities: req.PreferredCapabilities,
		ExcludeAgents:        req.ExcludeAgents,
		MaxResults:           int(req.MaxResults),
		SortBy:               req.SortBy,
	}
	
	// Convert status strings to AgentStatus
	if len(req.RequiredStatus) > 0 {
		statuses := make([]registry.AgentStatus, len(req.RequiredStatus))
		for i, status := range req.RequiredStatus {
			statuses[i] = registry.AgentStatus(status)
		}
		opts.RequiredStatus = statuses
	}

	scores, err := h.discovery.FindAgents(ctx, opts)
	if err != nil {
		return nil, err
	}

	// Convert scored agents to API format
	apiScores := make([]*api.ScoredAgent, len(scores))
	for i, score := range scores {
		apiScores[i] = convertScoredAgentToAPI(score)
	}

	return &api.FindAgentsResponse{Agents: apiScores}, nil
}

// GetClusterStatus returns overall cluster status information
func (h *Server) GetClusterStatus(ctx context.Context, req *api.GetClusterStatusRequest) (*api.GetClusterStatusResponse, error) {
	clusterStatus, err := h.discovery.GetClusterStatus(ctx)
	if err != nil {
		return nil, err
	}

	// Convert cluster status to API format
	apiStatus := &api.ClusterStatus{
		TotalAgents:     int32(clusterStatus.TotalAgents),
		ActiveAgents:    int32(clusterStatus.ActiveAgents),
		HealthyAgents:   int32(clusterStatus.HealthyAgents),
		UnhealthyAgents: int32(clusterStatus.UnhealthyAgents),
		Capabilities:    clusterStatus.Capabilities,
		LoadMetrics: &api.LoadMetrics{
			AverageLoad:    clusterStatus.LoadMetrics.AverageLoad,
			TotalLoad:      clusterStatus.LoadMetrics.TotalLoad,
			CapacityLeft:   clusterStatus.LoadMetrics.CapacityLeft,
		},
	}

	// Convert region status
	if clusterStatus.RegionStatus != nil {
		apiStatus.RegionStatus = make(map[string]*api.RegionMetrics)
		for region, metrics := range clusterStatus.RegionStatus {
			apiStatus.RegionStatus[region] = &api.RegionMetrics{
				AgentCount:     int32(metrics.AgentCount),
				AverageLatency: metrics.AverageLatency,
				ErrorRate:      metrics.ErrorRate,
			}
		}
	}

	if clusterStatus.LastUpdated != nil {
		apiStatus.LastUpdated = timestamppb.New(*clusterStatus.LastUpdated)
	}

	return &api.GetClusterStatusResponse{Status: apiStatus}, nil
}
*/
