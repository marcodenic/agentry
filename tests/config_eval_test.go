package tests

import (
	"context"
	"strings"
	"testing"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/internal/tool"
)

// MockClient that returns a tool call once, then a final output.
type cyclingMock struct {
	callCount int
}

func (m *cyclingMock) Complete(ctx context.Context, prompt string) (string, error) {
	m.callCount++
	if m.callCount == 1 {
		return `{"tool":"echo","args":{"text":"hello"}}`, nil
	}
	return "hello", nil
}

func TestConfigBootAndEval(t *testing.T) {
	// Inline echo tool
	tools := tool.Registry{
		"echo": tool.New("echo", "Repeats input text", func(ctx context.Context, args map[string]any) (string, error) {
			input, ok := args["text"].(string)
			if !ok {
				return "", nil
			}
			return input, nil
		}),
	}

	// Mock model
	mock := &cyclingMock{}

	route := router.Rules{{IfContains: []string{""}, Client: mock}}
	agent := core.New(route, tools, memory.NewInMemory(), nil)

	out, err := agent.Run(context.TODO(), "hello")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "hello") {
		t.Errorf("expected output to contain 'hello', got: %s", out)
	}
}
