package model

import (
	"context"
	"strings"
)

// Mock model that echoes prompt or returns JSON for tool call.
type Mock struct{}

func NewMock() *Mock { return &Mock{} }

func (m *Mock) Complete(ctx context.Context, prompt string) (string, error) {
	if strings.Contains(prompt, "\"tool\":\"echo\"") {
		return "final output", nil
	}
	// call echo tool
	return `{"tool":"echo","args":{}}`, nil
}
