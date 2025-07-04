package main

import (
	"context"
	"fmt"
	"log"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/tool"
)

// Simple mock for testing
type simpleMock struct{}

func (s simpleMock) Complete(ctx context.Context, msgs []model.ChatMessage, tools []model.ToolSpec) (model.Completion, error) {
	return model.Completion{
		Content: "I'll use the view tool to read the README.",
		ToolCalls: []model.ToolCall{
			{
				ID:        "call_1",
				Name:      "view",
				Arguments: []byte(`{"path": "README.md"}`),
			},
		},
	}, nil
}

func main() {
	fmt.Println("🧪 Testing fixed view tool...")

	// Set sandbox to disabled
	tool.SetSandboxEngine("disabled")

	// Create agent with tool registry
	client := simpleMock{}
	reg := tool.DefaultRegistry()
	agent := core.New(client, "mock", reg, memory.NewInMemory(), nil, memory.NewInMemoryVector(), nil)

	// Test the view tool directly first
	fmt.Println("🔧 Testing view tool directly...")
	viewTool, ok := reg.Use("view")
	if !ok {
		log.Fatal("view tool not found")
	}

	result, err := viewTool.Execute(context.Background(), map[string]any{"path": "README.md"})
	if err != nil {
		fmt.Printf("❌ View tool failed: %v\n", err)
	} else {
		fmt.Printf("✅ View tool works! Read %d characters\n", len(result))
	}

	// Test through agent
	fmt.Println("🤖 Testing through agent...")
	output, err := agent.Run(context.Background(), "test")
	if err != nil {
		fmt.Printf("❌ Agent failed: %v\n", err)
	} else {
		fmt.Printf("✅ Agent works! Output: %s\n", output)
	}
}
