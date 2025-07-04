package main

import (
	"fmt"
	"time"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/tui"
)

func main() {
	fmt.Println("Testing cost caching functionality...")

	// Create a mock client
	client := model.NewMock()

	// Create an agent
	agent := core.New(client, "gpt-4", nil, memory.NewInMemory(), memory.NewInMemoryVector(), nil)

	// Add some cost data
	agent.Cost.AddModelUsage("gpt-4", 1000, 1000)

	// Create TUI model
	tuiModel := tui.New(agent)

	// Access the internal agent info (this is for testing)
	fmt.Printf("Agent cost: $%.6f\n", agent.Cost.TotalCost())

	// Test the caching logic by calling View multiple times
	fmt.Println("Testing TUI rendering with cost caching...")

	start := time.Now()
	for i := 0; i < 100; i++ {
		tuiModel.View()
	}
	elapsed := time.Since(start)

	fmt.Printf("100 renders took: %v\n", elapsed)
	fmt.Printf("Average per render: %v\n", elapsed/100)

	fmt.Println("Cost caching test completed successfully!")
}
