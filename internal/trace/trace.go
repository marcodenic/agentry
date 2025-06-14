package trace

import (
	"context"
	"encoding/json"
	"io"
	"time"
)

type EventType string

const (
	EventStepStart EventType = "step_start"
	EventToolEnd   EventType = "tool_end"
	EventFinal     EventType = "final"
)

type Event struct {
	Timestamp time.Time   `json:"ts"`
	Type      EventType   `json:"type"`
	AgentID   string      `json:"agent_id"`
	Data      interface{} `json:"data,omitempty"`
}

type Writer interface {
	Write(ctx context.Context, e Event)
}

type JSONLWriter struct{ w io.Writer }

func NewJSONL(w io.Writer) *JSONLWriter { return &JSONLWriter{w} }
func (j *JSONLWriter) Write(_ context.Context, e Event) {
	_ = json.NewEncoder(j.w).Encode(e)
}

func Now() time.Time { return time.Now().UTC() }
