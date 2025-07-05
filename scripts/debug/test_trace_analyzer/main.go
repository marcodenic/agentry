package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/trace"
)

func main() {
	// Change to project root
	if err := os.Chdir(filepath.Join("..", "..")); err != nil {
		fmt.Printf("Error changing directory: %v\n", err)
		return
	}

	fmt.Println("Testing trace analyzer with model names...")

	// Create a mock completion with model name
	mockClient := model.NewMock()
	completion, err := mockClient.Complete(context.Background(), nil, nil)
	if err != nil {
		fmt.Printf("Error creating mock completion: %v\n", err)
		return
	}

	// Manually set some token counts for testing
	completion.InputTokens = 100
	completion.OutputTokens = 50
	fmt.Printf("Mock completion created with model: %s\n", completion.ModelName)

	// Create a trace event
	events := []trace.Event{
		{
			Type: trace.EventStepStart,
			Data: completion,
		},
	}

	// Analyze the trace
	summary := trace.Analyze("test input", events)

	fmt.Printf("\nTrace Analysis Results:\n")
	fmt.Printf("Total Input Tokens: %d\n", summary.InputTokens)
	fmt.Printf("Total Output Tokens: %d\n", summary.OutputTokens)
	fmt.Printf("Total Cost: $%.9f\n", summary.Cost)
	fmt.Printf("Model Usage: %+v\n", summary.ModelUsage)

	fmt.Println("\nTest completed!")
}
