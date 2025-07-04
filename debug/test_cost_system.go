package main

import (
	"fmt"

	"github.com/marcodenic/agentry/internal/cost"
)

func main() {
	fmt.Println("Testing cost system...")

	// Test basic functionality
	pricing := cost.NewPricingTable()

	// Test model pricing lookup
	if gpt4Pricing, found := pricing.GetPricing("gpt-4"); found {
		fmt.Printf("GPT-4 pricing: input=$%.2f/MTok, output=$%.2f/MTok\n", gpt4Pricing.InputPrice, gpt4Pricing.OutputPrice)
	} else {
		fmt.Println("GPT-4 pricing not found")
	}

	if claudePricing, found := pricing.GetPricing("claude-3-opus"); found {
		fmt.Printf("Claude-3-Opus pricing: input=$%.2f/MTok, output=$%.2f/MTok\n", claudePricing.InputPrice, claudePricing.OutputPrice)
	} else {
		fmt.Println("Claude-3-Opus pricing not found")
	}

	// Test cost calculation
	gpt4Cost := pricing.CalculateCost("gpt-4", 1000, 1000)
	fmt.Printf("Cost for 1000 input + 1000 output tokens with GPT-4: $%.6f\n", gpt4Cost)

	claudeCost := pricing.CalculateCost("claude-3-opus", 1000, 1000)
	fmt.Printf("Cost for 1000 input + 1000 output tokens with Claude-3-Opus: $%.6f\n", claudeCost)

	// Test cost manager
	manager := cost.New(0, 0)
	manager.AddModelUsage("gpt-4", 1000, 1000)
	manager.AddModelUsage("claude-3-opus", 500, 500)

	fmt.Printf("Total cost: $%.6f\n", manager.TotalCost())

	gpt4ManagerCost := manager.GetModelCost("gpt-4")
	fmt.Printf("GPT-4 cost: $%.6f\n", gpt4ManagerCost)

	claudeManagerCost := manager.GetModelCost("claude-3-opus")
	fmt.Printf("Claude cost: $%.6f\n", claudeManagerCost)

	fmt.Println("Cost system test completed successfully!")
}
