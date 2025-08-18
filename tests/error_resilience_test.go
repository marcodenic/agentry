package tests

import (
	"context"
	"strings"
	"testing"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/tool"
)

// errorClient simulates a client that calls a tool that will fail
type errorClient struct{}

func (errorClient) Complete(ctx context.Context, msgs []model.ChatMessage, tools []model.ToolSpec) (model.Completion, error) {
	// Return a completion that tries to call a non-existent tool
	return model.Completion{
		Content: "I'll use a tool that doesn't exist to test error handling.",
		ToolCalls: []model.ToolCall{
			{
				ID:        "call_123",
				Name:      "nonexistent_tool",
				Arguments: []byte(`{"test": "value"}`),
			},
		},
	}, nil
}

// resilientClient simulates a client that can recover from errors
type resilientClient struct {
	callCount int
}

func (c *resilientClient) Complete(ctx context.Context, msgs []model.ChatMessage, tools []model.ToolSpec) (model.Completion, error) {
	c.callCount++

	// Check if we received error feedback from previous tool call
	hasError := false
	hasSuccess := false
	for _, msg := range msgs {
		if msg.Role == "tool" && strings.Contains(msg.Content, "Error") {
			hasError = true
		}
		if msg.Role == "tool" && strings.Contains(msg.Content, "Recovery successful") {
			hasSuccess = true
		}
	}

	if c.callCount == 1 {
		// First call - try a non-existent tool
		return model.Completion{
			Content: "I'll try to use a tool that doesn't exist.",
			ToolCalls: []model.ToolCall{
				{
					ID:        "call_123",
					Name:      "nonexistent_tool",
					Arguments: []byte(`{"test": "value"}`),
				},
			},
		}, nil
	} else if hasError && !hasSuccess {
		// Second call - we got error feedback, now use a working tool
		return model.Completion{
			Content: "I see the previous tool failed. Let me try a working tool instead.",
			ToolCalls: []model.ToolCall{
				{
					ID:        "call_456",
					Name:      "echo",
					Arguments: []byte(`{"text": "Recovery successful!"}`),
				},
			},
		}, nil
	} else if hasSuccess {
		// Third call - we got success feedback, now return final result
		return model.Completion{
			Content: "Task completed successfully after recovering from error.",
		}, nil
	} else {
		// Fallback - just return content
		return model.Completion{
			Content: "Task completed.",
		}, nil
	}
}

func TestErrorHandlingWithNonResilientAgent(t *testing.T) {
	// Create agent without error resilience
	registry := tool.DefaultRegistry()
	client := errorClient{}
	agent := core.New(client, "mock", registry, memory.NewInMemory(), memory.NewInMemoryVector(), nil)

	// Disable error handling to test old behavior
	agent.ErrorHandling.TreatErrorsAsResults = false

	_, err := agent.Run(context.Background(), "test task")

	// Should fail with error
	if err == nil {
		t.Fatal("Expected error when calling non-existent tool without error handling")
	}

	if !strings.Contains(err.Error(), "unknown tool") {
		t.Fatalf("Expected 'unknown tool' error, got: %v", err)
	}
}

func TestErrorHandlingWithResilientAgent(t *testing.T) {
	// Create agent with error resilience
	registry := tool.DefaultRegistry()
	client := &resilientClient{}
	agent := core.New(client, "mock", registry, memory.NewInMemory(), memory.NewInMemoryVector(), nil)

	// Enable error handling (should be default with our changes)
	agent.ErrorHandling.TreatErrorsAsResults = true
	agent.ErrorHandling.MaxErrorRetries = 3
	agent.ErrorHandling.IncludeErrorContext = true

	result, err := agent.Run(context.Background(), "test task")

	t.Logf("Client made %d calls", client.callCount)
	t.Logf("Result: %q", result)
	t.Logf("Error: %v", err)

	// Should succeed despite initial error
	if err != nil {
		t.Fatalf("Expected success with error handling, got error: %v", err)
	}

	if result == "" {
		t.Fatal("Expected non-empty result")
	}

	if !strings.Contains(result, "completed successfully") {
		t.Fatalf("Expected successful completion message, got: %s", result)
	}

	// Verify the agent made multiple attempts
	if client.callCount < 2 {
		t.Fatalf("Expected at least 2 calls (error + recovery), got %d", client.callCount)
	}
}

// errorOnlyClient always tries to call non-existent tools
type errorOnlyClient struct {
	callCount int
}

func (c *errorOnlyClient) Complete(ctx context.Context, msgs []model.ChatMessage, tools []model.ToolSpec) (model.Completion, error) {
	c.callCount++
	return model.Completion{
		Content: "I'll keep trying non-existent tools.",
		ToolCalls: []model.ToolCall{
			{
				ID:        "call_" + string(rune('0'+c.callCount)),
				Name:      "nonexistent_tool",
				Arguments: []byte(`{"test": "value"}`),
			},
		},
	}, nil
}

func TestErrorHandlingTooManyErrors(t *testing.T) {
	client := &errorOnlyClient{}
	registry := tool.DefaultRegistry()
	agent := core.New(client, "mock", registry, memory.NewInMemory(), memory.NewInMemoryVector(), nil)

	// Enable error handling with low retry limit
	agent.ErrorHandling.TreatErrorsAsResults = true
	agent.ErrorHandling.MaxErrorRetries = 2
	agent.ErrorHandling.IncludeErrorContext = true

	_, err := agent.Run(context.Background(), "test task")

	// Should fail after exceeding retry limit
	if err == nil {
		t.Fatal("Expected error after exceeding retry limit")
	}

	if !strings.Contains(err.Error(), "too many consecutive errors") {
		t.Fatalf("Expected 'too many consecutive errors' message, got: %v", err)
	}
}
