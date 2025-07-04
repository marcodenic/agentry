package main

import (
	"fmt"

	"github.com/marcodenic/agentry/internal/cost"
)

func main() {
	fmt.Println("Testing new model pricing...")

	pricing := cost.NewPricingTable()

	// Test the new models provided by the user
	testModels := []string{
		"gpt-4.1",
		"gpt-4.1-2025-04-14",
		"gpt-4.1-mini",
		"gpt-4.1-mini-2025-04-14",
		"gpt-4.1-nano",
		"gpt-4.1-nano-2025-04-14",
		"claude-3-opus",
		"claude-3-sonnet",
		"claude-3-sonnet-3.7",
	}

	for _, model := range testModels {
		if modelPricing, found := pricing.GetPricing(model); found {
			fmt.Printf("%-25s: input=$%.2f/MTok, output=$%.2f/MTok\n",
				model, modelPricing.InputPrice, modelPricing.OutputPrice)
		} else {
			fmt.Printf("%-25s: NOT FOUND\n", model)
		}
	}

	fmt.Println("\nCost comparison for 10,000 tokens (input + output):")
	for _, model := range testModels {
		cost := pricing.CalculateCost(model, 10000, 10000)
		fmt.Printf("%-25s: $%.6f\n", model, cost)
	}

	fmt.Println("\nPricing verification completed!")
}
