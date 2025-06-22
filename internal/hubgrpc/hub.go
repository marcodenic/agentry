package hubgrpc

import (
	"context"
	"io"
	"log"
	"strings"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/marcodenic/agentry/api"
)

// Server implements the AgentHub gRPC API.
type Server struct {
	api.UnimplementedAgentHubServer
	nodes     []api.AgentNodeClient
	agentNode map[string]int
	mu        sync.Mutex
	next      int
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
	return &Server{nodes: nodes, agentNode: map[string]int{}}
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
