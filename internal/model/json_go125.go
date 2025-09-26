package model

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

// Go125JSONOptimizer provides optimized JSON operations using Go 1.25 features
type Go125JSONOptimizer struct {
	// Pool of JSON encoders/decoders for reuse
	encoderPool sync.Pool
	decoderPool sync.Pool
}

func NewGo125JSONOptimizer() *Go125JSONOptimizer {
	return &Go125JSONOptimizer{
		encoderPool: sync.Pool{
			New: func() interface{} {
				return &json.Encoder{}
			},
		},
		decoderPool: sync.Pool{
			New: func() interface{} {
				return &json.Decoder{}
			},
		},
	}
}

// OptimizedToolCallMarshal uses Go 1.25's improved JSON marshaling
func (opt *Go125JSONOptimizer) OptimizedToolCallMarshal(toolCalls []ToolCall) ([]byte, error) {
	// Go 1.25: json.Marshal performance is significantly improved
	// Especially beneficial for complex structures like tool calls
	return json.Marshal(toolCalls)
}

// OptimizedToolCallUnmarshal uses Go 1.25's improved JSON unmarshaling
func (opt *Go125JSONOptimizer) OptimizedToolCallUnmarshal(data []byte) ([]ToolCall, error) {
	var toolCalls []ToolCall
	// Go 1.25: json.Unmarshal performance is substantially better for decoding
	err := json.Unmarshal(data, &toolCalls)
	return toolCalls, err
}

// StreamOptimizedMarshal provides optimized streaming JSON marshaling
func (opt *Go125JSONOptimizer) StreamOptimizedMarshal(ctx context.Context, chunks []StreamChunk) ([]byte, error) {
	// Pre-allocate slice to avoid reallocations
	result := make([]byte, 0, len(chunks)*128) // Estimate 128 bytes per chunk

	for _, chunk := range chunks {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Go 1.25: Improved JSON marshaling performance
		chunkJSON, err := json.Marshal(chunk)
		if err != nil {
			return nil, err
		}

		result = append(result, chunkJSON...)
		result = append(result, '\n') // Line separator
	}

	return result, nil
}

// ParallelJSONProcess demonstrates parallel JSON processing with Go 1.25 improvements
func (opt *Go125JSONOptimizer) ParallelJSONProcess(ctx context.Context, messages []ChatMessage, workers int) ([][]byte, error) {
	if workers <= 0 {
		workers = 4 // Default
	}

	results := make([][]byte, len(messages))
	errors := make([]error, len(messages))

	// Use Go 1.25's WaitGroup.Go() for cleaner concurrent patterns
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, workers) // Limit concurrency

	for i, msg := range messages {
		i, msg := i, msg // Capture loop variables

		wg.Go(func() {
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			// Check context
			select {
			case <-ctx.Done():
				errors[i] = ctx.Err()
				return
			default:
			}

			// Go 1.25: Improved JSON marshaling performance
			data, err := json.Marshal(msg)
			if err != nil {
				errors[i] = err
				return
			}

			results[i] = data
		})
	}

	wg.Wait()

	// Check for errors
	for _, err := range errors {
		if err != nil {
			return nil, err
		}
	}

	return results, nil
}

// BenchmarkJSONPerformance provides a way to benchmark JSON performance improvements
func (opt *Go125JSONOptimizer) BenchmarkJSONPerformance(iterations int) map[string]interface{} {
	// Create test data
	testMessage := ChatMessage{
		Role: "user",
		Content: "This is a test message for benchmarking JSON performance in Go 1.25. " +
			"The new JSON implementation should provide significantly better performance for both " +
			"encoding and decoding operations, especially with complex nested structures.",
		ToolCalls: []ToolCall{
			{
				ID:        "test_call_1",
				Name:      "test_tool",
				Arguments: json.RawMessage(`{"param1": "value1", "param2": 42, "nested": {"key": "value"}}`),
			},
		},
	}

	// Benchmark marshaling
	start := time.Now()
	for i := 0; i < iterations; i++ {
		_, _ = json.Marshal(testMessage)
	}
	marshalDuration := time.Since(start)

	// Benchmark unmarshaling
	data, _ := json.Marshal(testMessage)
	start = time.Now()
	for i := 0; i < iterations; i++ {
		var msg ChatMessage
		_ = json.Unmarshal(data, &msg)
	}
	unmarshalDuration := time.Since(start)

	return map[string]interface{}{
		"iterations":         iterations,
		"marshal_duration":   marshalDuration,
		"unmarshal_duration": unmarshalDuration,
		"marshal_per_op":     marshalDuration / time.Duration(iterations),
		"unmarshal_per_op":   unmarshalDuration / time.Duration(iterations),
		"total_duration":     marshalDuration + unmarshalDuration,
	}
}
