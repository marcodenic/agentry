package trace

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

type SSEWriter struct {
	w  http.ResponseWriter
	fl http.Flusher
}

func NewSSE(w http.ResponseWriter) *SSEWriter {
	fl, _ := w.(http.Flusher)
	return &SSEWriter{w: w, fl: fl}
}

func (s *SSEWriter) Write(_ context.Context, e Event) {
	b, _ := json.Marshal(e)
	fmt.Fprintf(s.w, "data: %s\n\n", b)
	if s.fl != nil {
		defer func() { recover() }()
		s.fl.Flush()
	}
}

func Now() time.Time { return time.Now().UTC() }
