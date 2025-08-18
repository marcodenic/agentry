package model

import "context"

// StreamChunk represents an incremental model output segment.
type StreamChunk struct {
	ContentDelta string
	Done         bool
	Err          error
	// Populated only on final chunk when available
	InputTokens  int
	OutputTokens int
	ToolCalls    []ToolCall
}

// StreamingClient provides incremental output chunks.
// Channel must be closed after final chunk (Done=true) or on Err.
type StreamingClient interface {
	Stream(ctx context.Context, msgs []ChatMessage, tools []ToolSpec) (<-chan StreamChunk, error)
}
