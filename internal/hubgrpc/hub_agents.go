package hubgrpc

import (
	"context"
	"io"

	"github.com/marcodenic/agentry/api"
)

// Spawn creates a new agent on a selected node
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

// SendMessage sends a message to an agent on its assigned node
func (h *Server) SendMessage(ctx context.Context, req *api.SendMessageRequest) (*api.Ack, error) {
	h.mu.Lock()
	idx, ok := h.agentNode[req.AgentId]
	h.mu.Unlock()
	if !ok {
		return &api.Ack{Ok: false}, nil
	}
	return h.nodes[idx].SendMessage(ctx, req)
}

// Trace streams trace events from an agent on its assigned node
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
