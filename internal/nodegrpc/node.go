package nodegrpc

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/google/uuid"
	"github.com/marcodenic/agentry/api"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/internal/tool"
	"github.com/marcodenic/agentry/internal/trace"
	"google.golang.org/grpc"
)

// Server implements the AgentNode gRPC API.
type Server struct {
	api.UnimplementedAgentNodeServer
	mu     sync.Mutex
	agents map[string]*core.Agent
	traces map[string]*chanWriter
}

type chanWriter struct{ ch chan *api.TraceEvent }

func newChanWriter() *chanWriter { return &chanWriter{ch: make(chan *api.TraceEvent, 16)} }

func (c *chanWriter) Write(_ context.Context, e trace.Event) {
	b, _ := json.Marshal(e.Data)
	c.ch <- &api.TraceEvent{
		Type:      string(e.Type),
		AgentId:   e.AgentID,
		Data:      string(b),
		Timestamp: e.Timestamp.UnixNano(),
	}
}

func defaultAgent() *core.Agent {
	reg := tool.DefaultRegistry()
	route := router.Rules{{Name: "mock", IfContains: []string{""}, Client: model.NewMock()}}
	return core.New(route, reg, memory.NewInMemory(), nil, nil)
}

// New returns a new Server instance.
func New() *Server {
	return &Server{agents: map[string]*core.Agent{}, traces: map[string]*chanWriter{}}
}

func (s *Server) Spawn(ctx context.Context, _ *api.SpawnRequest) (*api.SpawnResponse, error) {
	ag := defaultAgent()
	id := uuid.New().String()
	ag.ID = uuid.MustParse(id)
	s.mu.Lock()
	s.agents[id] = ag
	s.mu.Unlock()
	return &api.SpawnResponse{AgentId: id}, nil
}

func (s *Server) SendMessage(ctx context.Context, req *api.SendMessageRequest) (*api.Ack, error) {
	s.mu.Lock()
	ag := s.agents[req.AgentId]
	if ag == nil {
		s.mu.Unlock()
		return &api.Ack{Ok: false}, nil
	}
	cw := newChanWriter()
	ag.Tracer = cw
	s.traces[req.AgentId] = cw
	s.mu.Unlock()
	go func() {
		ag.Run(ctx, req.Input)
		close(cw.ch)
	}()
	return &api.Ack{Ok: true}, nil
}

func (s *Server) Trace(req *api.TraceRequest, stream api.AgentNode_TraceServer) error {
	s.mu.Lock()
	cw := s.traces[req.AgentId]
	s.mu.Unlock()
	if cw == nil {
		return nil
	}
	for ev := range cw.ch {
		if err := stream.Send(ev); err != nil {
			return err
		}
	}
	return nil
}

// Register the server with a gRPC server.
func Register(grpcSrv *grpc.Server, s *Server) {
	api.RegisterAgentNodeServer(grpcSrv, s)
}
