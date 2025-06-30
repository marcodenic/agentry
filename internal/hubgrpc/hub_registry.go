package hubgrpc

// NOTE: Registry methods are commented out because the corresponding
// protobuf message types (RegisterAgentRequest, GetAgentRequest, etc.) 
// are not yet implemented in the API.

// TODO: Implement these when the required protobuf messages are added to api/agent.proto

/*
import (
	"context"

	"github.com/marcodenic/agentry/api"
	"github.com/marcodenic/agentry/internal/registry"
)

// RegisterAgent registers a new agent with the hub
func (h *Server) RegisterAgent(ctx context.Context, req *api.RegisterAgentRequest) (*api.Ack, error) {
	agent := &registry.Agent{
		ID:           req.Agent.Id,
		Name:         req.Agent.Name,
		Capabilities: req.Agent.Capabilities,
		Status:       registry.AgentStatus(req.Agent.Status),
		Endpoint:     req.Agent.Endpoint,
		Region:       req.Agent.Region,
		Metadata:     req.Agent.Metadata,
	}
	
	err := h.registry.RegisterAgent(ctx, agent)
	if err != nil {
		return nil, err
	}
	
	return &api.Ack{Success: true}, nil
}

// DeregisterAgent removes an agent from the hub
func (h *Server) DeregisterAgent(ctx context.Context, req *api.GetAgentRequest) (*api.Ack, error) {
	err := h.registry.DeregisterAgent(ctx, req.AgentId)
	if err != nil {
		return nil, err
	}
	
	return &api.Ack{Success: true}, nil
}

// GetAgent returns information about a specific agent
func (h *Server) GetAgent(ctx context.Context, req *api.GetAgentRequest) (*api.GetAgentResponse, error) {
	agent, err := h.registry.GetAgent(ctx, req.AgentId)
	if err != nil {
		return nil, err
	}
	
	apiAgent := convertAgentToAPI(agent)
	return &api.GetAgentResponse{Agent: apiAgent}, nil
}

// ListAgents returns all agents that match the given criteria
func (h *Server) ListAgents(ctx context.Context, req *api.ListAgentsRequest) (*api.ListAgentsResponse, error) {
	var agents []*registry.Agent
	var err error
	
	if len(req.Capabilities) > 0 {
		agents, err = h.registry.FindAgents(ctx, req.Capabilities)
	} else {
		agents, err = h.registry.ListAgents(ctx)
	}
	
	if err != nil {
		return nil, err
	}
	
	// Filter by status if specified
	if len(req.Status) > 0 {
		filtered := make([]*registry.Agent, 0, len(agents))
		for _, agent := range agents {
			if agent.Status == registry.AgentStatus(req.Status) {
				filtered = append(filtered, agent)
			}
		}
		agents = filtered
	}
	
	// Convert to API format
	apiAgents := make([]*api.Agent, len(agents))
	for i, agent := range agents {
		apiAgents[i] = convertAgentToAPI(agent)
	}
	
	return &api.ListAgentsResponse{Agents: apiAgents}, nil
}
*/
