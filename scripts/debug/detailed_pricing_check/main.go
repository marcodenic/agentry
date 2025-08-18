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

	fmt.Println("Detailed pricing check for gpt-4.1-nano...")

	pt := cost.NewPricingTable()
	
	modelName := "openai/gpt-4.1-nano"
	
	// Test the raw pricing lookup
	pricing, found := pt.GetPricingByModelName(modelName)
	if !found {
		fmt.Printf("❌ Model %s not found in pricing table\n", modelName)
		return
	}
	
	fmt.Printf("✅ Found pricing for %s\n", modelName)
	fmt.Printf("Input cost: $%.3f per 1M tokens\n", pricing.InputPrice)
	fmt.Printf("Output cost: $%.3f per 1M tokens\n", pricing.OutputPrice)
	
	// Manual calculation for 1053 tokens
	inputTokens := 1053
	outputTokens := 0
	
	manualCost := (float64(inputTokens) * pricing.InputPrice / 1000000.0) + (float64(outputTokens) * pricing.OutputPrice / 1000000.0)
	fmt.Printf("\nManual calculation for %d input + %d output tokens:\n", inputTokens, outputTokens)
	fmt.Printf("(%d × %.3f / 1,000,000) + (%d × %.3f / 1,000,000) = $%.9f\n", inputTokens, pricing.InputPrice, outputTokens, pricing.OutputPrice, manualCost)
	
	// Using the pricing table method
	calculatedCost := pt.CalculateCost(modelName, inputTokens, outputTokens)
	fmt.Printf("PricingTable.CalculateCost result: $%.9f\n", calculatedCost)
	
	// Compare with screenshot
	fmt.Printf("\nComparison:\n")
	fmt.Printf("Screenshot shows: $0.000108\n")
	fmt.Printf("Manual calc:      $%.9f\n", manualCost)
	fmt.Printf("PricingTable:     $%.9f\n", calculatedCost)
	
	// Check if they're close (within 0.000001)
	screenshotCost := 0.000108
	if manualCost >= screenshotCost-0.000001 && manualCost <= screenshotCost+0.000001 {
		fmt.Println("✅ Manual calculation matches screenshot (within precision)")
	} else {
		fmt.Println("❌ Manual calculation doesn't match screenshot")
	}
	
	if calculatedCost >= screenshotCost-0.000001 && calculatedCost <= screenshotCost+0.000001 {
		fmt.Println("✅ PricingTable calculation matches screenshot (within precision)")
	} else {
		fmt.Println("❌ PricingTable calculation doesn't match screenshot")
	}
}
