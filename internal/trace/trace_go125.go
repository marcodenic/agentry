package trace

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// JSONLWriterV2 uses Go 1.25's improved JSON encoding when available
// Fallback to standard encoding/json when experimental features are disabled
type JSONLWriterV2 struct {
	mu sync.Mutex
	w  io.Writer
}

func NewJSONLV2(w io.Writer) *JSONLWriterV2 {
	return &JSONLWriterV2{w: w}
}

func (j *JSONLWriterV2) Write(_ context.Context, e Event) {
	j.mu.Lock()
	defer j.mu.Unlock()

	// Go 1.25: When GOEXPERIMENT=jsonv2 is enabled, encoding/json uses the new
	// implementation which provides substantially better performance
	enc := json.NewEncoder(j.w)

	// For high-frequency trace events, use the optimized JSON encoder
	if err := enc.Encode(e); err != nil {
		log.Printf("trace encode error: %v", err)
		return
	}

	if fl, ok := j.w.(http.Flusher); ok {
		fl.Flush()
	}
}

// BatchJSONLWriter demonstrates batched JSON writing for improved performance
type BatchJSONLWriter struct {
	mu     sync.Mutex
	w      io.Writer
	events []Event
	timer  *time.Timer
	size   int
}

func NewBatchJSONL(w io.Writer, batchSize int, flushInterval time.Duration) *BatchJSONLWriter {
	b := &BatchJSONLWriter{
		w:      w,
		events: make([]Event, 0, batchSize),
		size:   batchSize,
	}

	// Flush periodically
	b.timer = time.AfterFunc(flushInterval, b.flushBatch)

	return b
}

func (b *BatchJSONLWriter) Write(_ context.Context, e Event) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.events = append(b.events, e)

	if len(b.events) >= b.size {
		b.flushBatchLocked()
	}
}

func (b *BatchJSONLWriter) flushBatch() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.flushBatchLocked()
}

func (b *BatchJSONLWriter) flushBatchLocked() {
	if len(b.events) == 0 {
		return
	}

	// Go 1.25: Batch encoding benefits significantly from the new JSON implementation
	enc := json.NewEncoder(b.w)
	for _, event := range b.events {
		if err := enc.Encode(event); err != nil {
			log.Printf("batch trace encode error: %v", err)
		}
	}

	if fl, ok := b.w.(http.Flusher); ok {
		fl.Flush()
	}

	// Reset batch
	b.events = b.events[:0]

	// Reset timer
	b.timer.Reset(time.Minute)
}

func (b *BatchJSONLWriter) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.timer != nil {
		b.timer.Stop()
	}

	b.flushBatchLocked()
}

// HighPerformanceSSEWriter optimizes SSE writing with Go 1.25 features
type HighPerformanceSSEWriter struct {
	w      http.ResponseWriter
	fl     http.Flusher
	buffer []byte // Pre-allocated buffer for JSON marshaling
}

func NewHighPerformanceSSE(w http.ResponseWriter) *HighPerformanceSSEWriter {
	fl, _ := w.(http.Flusher)
	return &HighPerformanceSSEWriter{
		w:      w,
		fl:     fl,
		buffer: make([]byte, 0, 1024), // Pre-allocate 1KB buffer
	}
}

func (s *HighPerformanceSSEWriter) Write(_ context.Context, e Event) {
	// Go 1.25: json.Marshal benefits from the new JSON implementation
	b, err := json.Marshal(e)
	if err != nil {
		log.Printf("trace marshal error: %v", err)
		return
	}

	// Reuse buffer to avoid allocations
	s.buffer = s.buffer[:0]
	s.buffer = append(s.buffer, "data: "...)
	s.buffer = append(s.buffer, b...)
	s.buffer = append(s.buffer, "\n\n"...)

	if _, err := s.w.Write(s.buffer); err != nil {
		log.Printf("trace write error: %v", err)
		return
	}

	if s.fl != nil {
		s.fl.Flush()
	}
}
