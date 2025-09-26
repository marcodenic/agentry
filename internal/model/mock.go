package model

import (
	"context"
	"encoding/json"
)

// Mock model that cycles through one tool call then returns a final message.
type Mock struct {
	call int
}

func NewMock() *Mock { return &Mock{} }

func (m *Mock) Clone() Client { return NewMock() }

func (m *Mock) Stream(ctx context.Context, msgs []ChatMessage, tools []ToolSpec) (<-chan StreamChunk, error) {
	out := make(chan StreamChunk, 1)
	go func() {
		defer close(out)
		m.call++
		if m.call == 1 {
			args, _ := json.Marshal(map[string]string{"text": "hello"})
			out <- StreamChunk{
				ToolCalls: []ToolCall{{ID: "1", Name: "echo", Arguments: args}},
				Done:      true,
			}
		} else {
			out <- StreamChunk{
				ContentDelta: "hello",
				Done:         true,
			}
		}
	}()
	return out, nil
}
