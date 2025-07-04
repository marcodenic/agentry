package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/marcodenic/agentry/internal/cost"
)

func main() {
	// Change to project root
	if err := os.Chdir(filepath.Join("..", "..")); err != nil {
		fmt.Printf("Error changing directory: %v\n", err)
		return
	}

	fmt.Println("Debugging pricing table contents...")

	pt := cost.NewPricingTable()
	
	// Try to find some models that are loaded
	testModels := []string{
		"openai/gpt-4.1-nano",
		"openai/gpt-4",
		"openai/gpt-3.5-turbo",
		"openai/gpt-4o",
		"openai/gpt-4o-mini",
		"anthropic/claude-3-opus",
		"anthropic/claude-3-sonnet",
	}
	
	fmt.Println("\nChecking test models:")
	foundAny := false
	for _, model := range testModels {
		_, found := pt.GetPricingByModelName(model)
		if found {
			fmt.Printf("✅ Found: %s\n", model)
			foundAny = true
		} else {
			fmt.Printf("❌ Not found: %s\n", model)
		}
	}
	
	if !foundAny {
		fmt.Println("\n❌ No models found at all - pricing table might not be loading")
		
		// Check if the cache file exists
		cacheFile := "internal/cost/data/models_pricing.json"
		if _, err := os.Stat(cacheFile); err != nil {
			fmt.Printf("❌ Cache file not found: %s\n", cacheFile)
		} else {
			fmt.Printf("✅ Cache file exists: %s\n", cacheFile)
		}
	}
	
	// Test the cost calculation directly
	fmt.Println("\nTesting cost calculation:")
	cost := pt.CalculateCost("openai/gpt-4.1-nano", 1053, 0)
	fmt.Printf("Cost for openai/gpt-4.1-nano (1053 input tokens): $%.9f\n", cost)
	
	// Try some variations
	variations := []string{
		"gpt-4.1-nano",
		"openai-gpt-4.1-nano",
		"openai/gpt-4.1-nano",
	}
	
	fmt.Println("\nTrying model name variations:")
	for _, variation := range variations {
		_, found := pt.GetPricingByModelName(variation)
		if found {
			fmt.Printf("✅ Found: %s\n", variation)
		} else {
			fmt.Printf("❌ Not found: %s\n", variation)
		}
	}
}
