package trace

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"runtime"
	"sync"
	"time"
)

// PerformanceWriter demonstrates Go 1.25 performance improvements for trace writing
type PerformanceWriter struct {
	mu           sync.Mutex
	w            io.Writer
	encoder      *json.Encoder
	bufferPool   *sync.Pool
	writeMetrics map[string]int64
}

func NewPerformanceWriter(w io.Writer) *PerformanceWriter {
	pw := &PerformanceWriter{
		w:            w,
		encoder:      json.NewEncoder(w),
		writeMetrics: make(map[string]int64),
		bufferPool: &sync.Pool{
			New: func() interface{} {
				return make([]byte, 0, 512) // Pre-allocate 512 bytes
			},
		},
	}

	// In Go 1.25, JSON encoding is significantly faster
	// Pre-configure encoder for optimal performance
	pw.encoder.SetEscapeHTML(false) // Faster when HTML escaping not needed
	
	return pw
}

func (pw *PerformanceWriter) Write(ctx context.Context, e Event) {
	pw.mu.Lock()
	defer pw.mu.Unlock()

	// Track metrics
	pw.writeMetrics["total_writes"]++
	start := time.Now()

	// Go 1.25: Encoder.Encode() benefits from the new JSON implementation
	if err := pw.encoder.Encode(e); err != nil {
		log.Printf("trace encode error: %v", err)
		return
	}

	// Track encoding duration
	encodeDuration := time.Since(start)
	pw.writeMetrics["total_encode_time_ns"] += encodeDuration.Nanoseconds()

	if fl, ok := pw.w.(http.Flusher); ok {
		fl.Flush()
	}
}

func (pw *PerformanceWriter) GetMetrics() map[string]interface{} {
	pw.mu.Lock()
	defer pw.mu.Unlock()

	totalWrites := pw.writeMetrics["total_writes"]
	totalEncodeTime := pw.writeMetrics["total_encode_time_ns"]

	metrics := map[string]interface{}{
		"total_writes":        totalWrites,
		"total_encode_time":   time.Duration(totalEncodeTime),
		"go_version":          runtime.Version(),
		"go125_improvements":  "JSON encoding performance enhanced",
	}

	if totalWrites > 0 {
		avgEncodeTime := totalEncodeTime / totalWrites
		metrics["avg_encode_time_per_event"] = time.Duration(avgEncodeTime)
	}

	return metrics
}

// ConcurrentWriter demonstrates Go 1.25's improved concurrent patterns
type ConcurrentWriter struct {
	writers []Writer
	wg      sync.WaitGroup
}

func NewConcurrentWriter(writers ...Writer) *ConcurrentWriter {
	return &ConcurrentWriter{writers: writers}
}

func (cw *ConcurrentWriter) Write(ctx context.Context, e Event) {
	// Go 1.25: Use WaitGroup.Go() for cleaner concurrent patterns
	for _, writer := range cw.writers {
		writer := writer // Capture loop variable
		cw.wg.Go(func() {
			writer.Write(ctx, e)
		})
	}
	cw.wg.Wait()
}

// ContainerAwareWriter adapts to container environments using Go 1.25's container awareness
type ContainerAwareWriter struct {
	Writer
	maxProcs      int
	originalProcs int
}

func NewContainerAwareWriter(w Writer) *ContainerAwareWriter {
	// Go 1.25: GOMAXPROCS is now container-aware by default
	maxProcs := runtime.GOMAXPROCS(0)
	
	return &ContainerAwareWriter{
		Writer:        w,
		maxProcs:      maxProcs,
		originalProcs: maxProcs,
	}
}

func (caw *ContainerAwareWriter) Write(ctx context.Context, e Event) {
	// Add container awareness info to events
	enhancedEvent := e
	if enhancedEvent.Data == nil {
		enhancedEvent.Data = make(map[string]interface{})
	}
	
	// Add Go 1.25 container awareness metadata
	if dataMap, ok := enhancedEvent.Data.(map[string]interface{}); ok {
		dataMap["go125_gomaxprocs"] = caw.maxProcs
		dataMap["container_aware"] = true
	}
	
	caw.Writer.Write(ctx, enhancedEvent)
}

// BenchmarkResult represents the results of JSON performance benchmarking
type BenchmarkResult struct {
	GoVersion          string        `json:"go_version"`
	Operations         int           `json:"operations"`
	MarshalDuration    time.Duration `json:"marshal_duration"`
	UnmarshalDuration  time.Duration `json:"unmarshal_duration"`
	TotalDuration      time.Duration `json:"total_duration"`
	MarshalOpsPerSec   float64       `json:"marshal_ops_per_sec"`
	UnmarshalOpsPerSec float64       `json:"unmarshal_ops_per_sec"`
}

// BenchmarkJSON demonstrates the performance improvements in Go 1.25 JSON operations
func BenchmarkJSON(operations int) BenchmarkResult {
	// Create a representative Event for benchmarking
	testEvent := Event{
		Timestamp: time.Now(),
		Type:      EventStepStart,
		AgentID:   "benchmark_agent",
		Data: map[string]interface{}{
			"step":        "benchmark_test",
			"parameters":  []string{"param1", "param2", "param3"},
			"metadata":    map[string]string{"version": "1.0", "type": "test"},
			"performance": true,
			"timestamp":   time.Now().Unix(),
		},
	}

	// Benchmark Marshal operations
	start := time.Now()
	var lastMarshalData []byte
	for i := 0; i < operations; i++ {
		data, err := json.Marshal(testEvent)
		if err == nil {
			lastMarshalData = data
		}
	}
	marshalDuration := time.Since(start)

	// Benchmark Unmarshal operations
	start = time.Now()
	for i := 0; i < operations; i++ {
		var event Event
		_ = json.Unmarshal(lastMarshalData, &event)
	}
	unmarshalDuration := time.Since(start)

	totalDuration := marshalDuration + unmarshalDuration

	return BenchmarkResult{
		GoVersion:          runtime.Version(),
		Operations:         operations,
		MarshalDuration:    marshalDuration,
		UnmarshalDuration:  unmarshalDuration,
		TotalDuration:      totalDuration,
		MarshalOpsPerSec:   float64(operations) / marshalDuration.Seconds(),
		UnmarshalOpsPerSec: float64(operations) / unmarshalDuration.Seconds(),
	}
}
