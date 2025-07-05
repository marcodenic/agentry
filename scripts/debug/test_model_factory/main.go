package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/marcodenic/agentry/internal/config"
	"github.com/marcodenic/agentry/internal/model"
)

func main() {
	// Change to project root
	if err := os.Chdir(filepath.Join("..", "..")); err != nil {
		fmt.Printf("Error changing directory: %v\n", err)
		return
	}

	fmt.Println("Testing model factory with missing model names...")

	// Test OpenAI with missing model
	openAIManifest := config.ModelManifest{
		Provider: "openai",
		Options: map[string]string{
			"key": "test-key",
			// No model specified
		},
	}

	fmt.Println("\nTesting OpenAI with missing model:")
	client, err := model.FromManifest(openAIManifest)
	if err != nil {
		fmt.Printf("✓ Correctly rejected OpenAI with missing model: %v\n", err)
	} else {
		fmt.Printf("✗ OpenAI should have been rejected but got client: %v\n", client)
	}

	// Test Anthropic with missing model
	anthropicManifest := config.ModelManifest{
		Provider: "anthropic",
		Options: map[string]string{
			"key": "test-key",
			// No model specified
		},
	}

	fmt.Println("\nTesting Anthropic with missing model:")
	client, err = model.FromManifest(anthropicManifest)
	if err != nil {
		fmt.Printf("✓ Correctly rejected Anthropic with missing model: %v\n", err)
	} else {
		fmt.Printf("✗ Anthropic should have been rejected but got client: %v\n", client)
	}

	// Test OpenAI with model specified (should work)
	openAIWithModel := config.ModelManifest{
		Provider: "openai",
		Options: map[string]string{
			"key":   "test-key",
			"model": "gpt-4",
		},
	}

	fmt.Println("\nTesting OpenAI with model specified:")
	client, err = model.FromManifest(openAIWithModel)
	if err != nil {
		fmt.Printf("✗ OpenAI with model should have worked but got error: %v\n", err)
	} else {
		fmt.Printf("✓ OpenAI with model correctly created client\n")
	}

	fmt.Println("\nAll tests completed!")
}
