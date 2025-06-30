package hubgrpc

import (
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

// Register the server with a gRPC server.
func Register(s *grpc.Server, h *Server) { api.RegisterAgentHubServer(s, h) }

// GetRegistry returns the underlying registry for testing
func (h *Server) GetRegistry() registry.AgentRegistry {
	return h.registry
}

// pick selects the next node using round-robin
func (h *Server) pick() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	idx := h.next % len(h.nodes)
	h.next++
	return idx
}
