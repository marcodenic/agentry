package core

import (
	"context"

	"github.com/marcodenic/agentry/internal/trace"
)

// Trace reports agent events to the configured tracer.
func (a *Agent) Trace(ctx context.Context, typ trace.EventType, data any) {
	if a.Tracer == nil {
		return
	}

	a.Tracer.Write(ctx, trace.Event{
		Type:      typ,
		AgentID:   a.ID.String(),
		Data:      data,
		Timestamp: trace.Now(),
	})
}
