package main

import (
	"fmt"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
)

func main() {
	fmt.Println("Debugging Agent 0 model name...")

	// Create a mock client
	client := model.NewMock()

	// Create an agent
	agent := core.New(client, "gpt-4o-mini", nil, memory.NewInMemory(), memory.NewInMemoryVector(), nil)

	fmt.Printf("Agent ModelName: '%s'\n", agent.ModelName)

	// Try with OpenAI client to see what happens
	openaiClient := model.NewOpenAI("gpt-4o-mini", "test-key")
	openaiAgent := core.New(openaiClient, "gpt-4o-mini", nil, memory.NewInMemory(), memory.NewInMemoryVector(), nil)
	fmt.Printf("OpenAI Agent ModelName: '%s'\n", openaiAgent.ModelName)
}
