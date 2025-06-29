package hubgrpc

import (
	"context"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/marcodenic/agentry/api"
	"github.com/marcodenic/agentry/internal/registry"
)

// Server implements the AgentHub gRPC API.
type Server struct {
	api.UnimplementedAgentHubServer
	nodes     []api.AgentNodeClient
	agentNode map[string]int
	mu        sync.Mutex
	next      int
	
	// Registry functionality
	registry  registry.AgentRegistry
	discovery *registry.DiscoveryService
}

// New returns a new Server dialing provided node addresses.
func New(addrs []string) *Server {
	nodes := []api.AgentNodeClient{}
	for _, a := range addrs {
		a = strings.TrimSpace(a)
		if a == "" {
			continue
		}
		conn, err := grpc.Dial(a, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("dial node %s: %v", a, err)
		}
		nodes = append(nodes, api.NewAgentNodeClient(conn))
	}
	
	// Initialize registry
	reg := registry.NewInMemoryRegistry(30*time.Second, 60*time.Second)
	discovery := registry.NewDiscoveryService(reg)
	
	return &Server{
		nodes:     nodes,
		agentNode: map[string]int{},
		registry:  reg,
		discovery: discovery,
	}
}

func (h *Server) pick() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	idx := h.next % len(h.nodes)
	h.next++
	return idx
}

func (h *Server) Spawn(ctx context.Context, req *api.SpawnRequest) (*api.SpawnResponse, error) {
	idx := h.pick()
	resp, err := h.nodes[idx].Spawn(ctx, req)
	if err != nil {
		return nil, err
	}
	h.mu.Lock()
	h.agentNode[resp.AgentId] = idx
	h.mu.Unlock()
	return resp, nil
}

func (h *Server) SendMessage(ctx context.Context, req *api.SendMessageRequest) (*api.Ack, error) {
	h.mu.Lock()
	idx, ok := h.agentNode[req.AgentId]
	h.mu.Unlock()
	if !ok {
		return &api.Ack{Ok: false}, nil
	}
	return h.nodes[idx].SendMessage(ctx, req)
}

func (h *Server) Trace(req *api.TraceRequest, stream api.AgentHub_TraceServer) error {
	h.mu.Lock()
	idx, ok := h.agentNode[req.AgentId]
	h.mu.Unlock()
	if !ok {
		return nil
	}
	ns, err := h.nodes[idx].Trace(stream.Context(), req)
	if err != nil {
		return err
	}
	for {
		ev, err := ns.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if err := stream.Send(ev); err != nil {
			return err
		}
	}
	return nil
}

// Register the server with a gRPC server.
func Register(s *grpc.Server, h *Server) { api.RegisterAgentHubServer(s, h) }

// Registry method implementations

// RegisterAgent registers a new agent with the registry
func (h *Server) RegisterAgent(ctx context.Context, req *api.RegisterAgentRequest) (*api.Ack, error) {
	agentInfo := &registry.AgentInfo{
		ID:           req.AgentId,
		Capabilities: req.Capabilities,
		Endpoint:     req.Endpoint,
		Metadata:     req.Metadata,
		SessionID:    req.SessionId,
		Role:         req.Role,
		Version:      req.Version,
		Status:       registry.StatusIdle,
	}
	
	err := h.registry.RegisterAgent(ctx, agentInfo)
	if err != nil {
		return &api.Ack{Ok: false}, err
	}
	
	return &api.Ack{Ok: true}, nil
}

// DeregisterAgent removes an agent from the registry
func (h *Server) DeregisterAgent(ctx context.Context, req *api.GetAgentRequest) (*api.Ack, error) {
	err := h.registry.DeregisterAgent(ctx, req.AgentId)
	if err != nil {
		return &api.Ack{Ok: false}, err
	}
	
	return &api.Ack{Ok: true}, nil
}

// GetAgent retrieves information about a specific agent
func (h *Server) GetAgent(ctx context.Context, req *api.GetAgentRequest) (*api.GetAgentResponse, error) {
	agentInfo, err := h.registry.GetAgent(ctx, req.AgentId)
	if err != nil {
		return nil, err
	}
	
	apiAgent := h.toAPIAgentInfo(agentInfo)
	return &api.GetAgentResponse{Agent: apiAgent}, nil
}

// ListAgents returns all registered agents
func (h *Server) ListAgents(ctx context.Context, req *api.ListAgentsRequest) (*api.ListAgentsResponse, error) {
	var agents []*registry.AgentInfo
	var err error
	
	if len(req.Capabilities) > 0 {
		agents, err = h.registry.FindAgents(ctx, req.Capabilities)
	} else {
		agents, err = h.registry.ListAllAgents(ctx)
	}
	
	if err != nil {
		return nil, err
	}
	
	// Filter by status and role if specified
	agents = h.filterAgents(agents, req.Statuses, req.Role)
	
	// Convert to API format
	apiAgents := make([]*api.AgentInfo, len(agents))
	for i, agent := range agents {
		apiAgents[i] = h.toAPIAgentInfo(agent)
	}
	
	return &api.ListAgentsResponse{Agents: apiAgents}, nil
}

// UpdateAgentStatus updates the status of an agent
func (h *Server) UpdateAgentStatus(ctx context.Context, req *api.UpdateAgentStatusRequest) (*api.Ack, error) {
	status := registry.AgentStatus(req.Status)
	err := h.registry.UpdateAgentStatus(ctx, req.AgentId, status)
	if err != nil {
		return &api.Ack{Ok: false}, err
	}
	
	return &api.Ack{Ok: true}, nil
}

// UpdateHealth updates health metrics for an agent
func (h *Server) UpdateHealth(ctx context.Context, req *api.UpdateHealthRequest) (*api.Ack, error) {
	health := &registry.HealthMetrics{
		CPUUsage:       req.Health.CpuUsage,
		MemoryUsage:    req.Health.MemoryUsage,
		TasksCompleted: req.Health.TasksCompleted,
		TasksActive:    int(req.Health.TasksActive),
		ErrorCount:     req.Health.ErrorCount,
		Uptime:         req.Health.UptimeSeconds,
	}
	
	err := h.registry.UpdateAgentHealth(ctx, req.AgentId, health)
	if err != nil {
		return &api.Ack{Ok: false}, err
	}
	
	return &api.Ack{Ok: true}, nil
}

// Heartbeat updates the last seen time for an agent
func (h *Server) Heartbeat(ctx context.Context, req *api.HeartbeatRequest) (*api.Ack, error) {
	err := h.registry.Heartbeat(ctx, req.AgentId)
	if err != nil {
		return &api.Ack{Ok: false}, err
	}
	
	return &api.Ack{Ok: true}, nil
}

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
		opts.RequiredStatus = make([]registry.AgentStatus, len(req.RequiredStatus))
		for i, status := range req.RequiredStatus {
			opts.RequiredStatus[i] = registry.AgentStatus(status)
		}
	}
	
	scores, err := h.discovery.FindBestAgents(ctx, opts)
	if err != nil {
		return nil, err
	}
	
	// Convert to API format
	apiScores := make([]*api.AgentScore, len(scores))
	for i, score := range scores {
		apiScores[i] = &api.AgentScore{
			Agent: h.toAPIAgentInfo(score.Agent),
			Score: score.Score,
		}
	}
	
	return &api.FindAgentsResponse{Agents: apiScores}, nil
}

// GetClusterStatus returns overall cluster status information
func (h *Server) GetClusterStatus(ctx context.Context, req *api.GetClusterStatusRequest) (*api.GetClusterStatusResponse, error) {
	clusterStatus, err := h.discovery.GetClusterStatus(ctx)
	if err != nil {
		return nil, err
	}
	
	// Convert to API format
	statusCounts := make(map[string]int32)
	for status, count := range clusterStatus.StatusCounts {
		statusCounts[string(status)] = int32(count)
	}
	
	capabilities := make(map[string]int32)
	for cap, count := range clusterStatus.Capabilities {
		capabilities[cap] = int32(count)
	}
	
	roles := make(map[string]int32)
	for role, count := range clusterStatus.Roles {
		roles[role] = int32(count)
	}
	
	apiStatus := &api.ClusterStatus{
		TotalAgents:         int32(clusterStatus.TotalAgents),
		StatusCounts:        statusCounts,
		Capabilities:        capabilities,
		Roles:              roles,
		AverageUptimeSeconds: int64(clusterStatus.AverageUptime.Seconds()),
		TotalTasksCompleted: clusterStatus.TotalTasksCompleted,
		TotalErrors:         clusterStatus.TotalErrors,
		LastUpdated:         clusterStatus.LastUpdated.Unix(),
	}
	
	return &api.GetClusterStatusResponse{Status: apiStatus}, nil
}

// Helper methods for registry functionality

func (h *Server) toAPIAgentInfo(agent *registry.AgentInfo) *api.AgentInfo {
	return &api.AgentInfo{
		Id:           agent.ID,
		Capabilities: agent.Capabilities,
		Endpoint:     agent.Endpoint,
		Status:       string(agent.Status),
		Metadata:     agent.Metadata,
		LastSeen:     agent.LastSeen.Unix(),
		RegisteredAt: agent.RegisteredAt.Unix(),
		SessionId:    agent.SessionID,
		Role:         agent.Role,
		Version:      agent.Version,
	}
}

func (h *Server) filterAgents(agents []*registry.AgentInfo, statuses []string, role string) []*registry.AgentInfo {
	var filtered []*registry.AgentInfo
	
	for _, agent := range agents {
		// Filter by status
		if len(statuses) > 0 {
			statusMatch := false
			for _, status := range statuses {
				if string(agent.Status) == status {
					statusMatch = true
					break
				}
			}
			if !statusMatch {
				continue
			}
		}
		
		// Filter by role
		if role != "" && agent.Role != role {
			continue
		}
		
		filtered = append(filtered, agent)
	}
	
	return filtered
}

// GetRegistry returns the underlying registry for testing
func (h *Server) GetRegistry() registry.AgentRegistry {
	return h.registry
}
