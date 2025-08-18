package main

import (
	"fmt"
	"github.com/marcodenic/agentry/internal/cost"
)

func main() {
	pricing := cost.NewPricingTable()

	models := []string{"gpt-4", "gpt-4o", "gpt-4o-mini", "gpt-4.1", "claude-3-opus", "claude-3-sonnet", "claude-3-haiku"}

	fmt.Println("Testing model pricing lookup:")
	for _, model := range models {
		if p, found := pricing.GetPricingByModelName(model); found {
			fmt.Printf("%-20s : input=$%.2f/MTok, output=$%.2f/MTok\n", model, p.InputPrice, p.OutputPrice)
		} else {
			fmt.Printf("%-20s : NOT FOUND\n", model)
		}
	}

	// Test some specific Azure/OpenAI variants
	fmt.Println("\nAzure vs OpenAI variants:")
	azureModels := []string{"azure/gpt-4", "openai/gpt-4", "gpt-4"}
	for _, model := range azureModels {
		if p, found := pricing.GetPricingByModelName(model); found {
			fmt.Printf("%-20s : input=$%.2f/MTok, output=$%.2f/MTok\n", model, p.InputPrice, p.OutputPrice)
		}
	}
}
