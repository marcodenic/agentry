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
func (m *Mock) Complete(ctx context.Context, msgs []ChatMessage, tools []ToolSpec) (Completion, error) {
	m.call++
	if m.call == 1 {
		args, _ := json.Marshal(map[string]string{"text": "hello"})
		return Completion{
			ToolCalls: []ToolCall{{ID: "1", Name: "echo", Arguments: args}},
			ModelName: "mock/test",
		}, nil
	}
	return Completion{
		Content:   "hello",
		ModelName: "mock/test",
	}, nil
}
