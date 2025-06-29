package trace

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

type EventType string

const (
	EventStepStart  EventType = "step_start"
	EventToolEnd    EventType = "tool_end"
	EventFinal      EventType = "final"
	EventModelStart EventType = "model_start"
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

type JSONLWriter struct {
	mu sync.Mutex
	w  io.Writer
}

func NewJSONL(w io.Writer) *JSONLWriter { return &JSONLWriter{w: w} }
func (j *JSONLWriter) Write(_ context.Context, e Event) {
	j.mu.Lock()
	defer j.mu.Unlock()
	enc := json.NewEncoder(j.w)
	_ = enc.Encode(e)
	if fl, ok := j.w.(http.Flusher); ok {
		fl.Flush()
	}
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
	b, err := json.Marshal(e)
	if err != nil {
		log.Printf("trace marshal error: %v", err)
		return
	}
	if _, err := fmt.Fprintf(s.w, "data: %s\n\n", b); err != nil {
		log.Printf("trace write error: %v", err)
		return
	}
	if s.fl != nil {
		s.fl.Flush()
	}
}

func Now() time.Time { return time.Now().UTC() }
