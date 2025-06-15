package model

import "context"

// ChatMessage represents a message sent to or received from the model.
type ChatMessage struct {
	Role       string `json:"role"`
	Content    string `json:"content,omitempty"`
	Name       string `json:"name,omitempty"`
	ToolCallID string `json:"tool_call_id,omitempty"`
}

// ToolSpec describes a callable tool for the model.
type ToolSpec struct {
	Name        string
	Description string
	Parameters  map[string]any
}

// ToolCall returned by the model.
type ToolCall struct {
	ID        string
	Name      string
	Arguments []byte
}

// Completion holds either final content or tool calls.
type Completion struct {
	Content   string
	ToolCalls []ToolCall
}

// Client defines the interface for language model backends.
type Client interface {
	Complete(ctx context.Context, msgs []ChatMessage, tools []ToolSpec) (Completion, error)
}
