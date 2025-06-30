package hubgrpc

// NOTE: Health methods are commented out because the corresponding
// protobuf message types (UpdateAgentStatusRequest, UpdateHealthRequest, etc.) 
// are not yet implemented in the API.

// TODO: Implement these when the required protobuf messages are added to api/agent.proto

/*
import (
	"context"

	"github.com/marcodenic/agentry/api"
	"github.com/marcodenic/agentry/internal/registry"
)

// UpdateAgentStatus updates the status of an agent
func (h *Server) UpdateAgentStatus(ctx context.Context, req *api.UpdateAgentStatusRequest) (*api.Ack, error) {
	status := registry.AgentStatus(req.Status)
	err := h.registry.UpdateAgentStatus(ctx, req.AgentId, status)
	if err != nil {
		return nil, err
	}
	
	return &api.Ack{Success: true}, nil
}

// UpdateHealth updates health metrics for an agent
func (h *Server) UpdateHealth(ctx context.Context, req *api.UpdateHealthRequest) (*api.Ack, error) {
	health := &registry.AgentHealth{
		CPUUsage:      req.Health.CpuUsage,
		MemoryUsage:   req.Health.MemoryUsage,
		DiskUsage:     req.Health.DiskUsage,
		ActiveTasks:   int(req.Health.ActiveTasks),
		ErrorRate:     req.Health.ErrorRate,
		ResponseTime:  req.Health.ResponseTime,
		LastCheckTime: req.Health.LastCheckTime.AsTime(),
	}
	
	err := h.registry.UpdateHealth(ctx, req.AgentId, health)
	if err != nil {
		return nil, err
	}
	
	return &api.Ack{Success: true}, nil
}

// Heartbeat updates the last seen time for an agent
func (h *Server) Heartbeat(ctx context.Context, req *api.HeartbeatRequest) (*api.Ack, error) {
	err := h.registry.Heartbeat(ctx, req.AgentId)
	if err != nil {
		return nil, err
	}
	
	return &api.Ack{Success: true}, nil
}
*/
