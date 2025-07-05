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

	fmt.Println("Checking pricing for gpt-4.1-nano...")

	pt := cost.NewPricingTable()
	
	modelName := "openai/gpt-4.1-nano"
	
	// Calculate cost for 1053 tokens (assuming all input for system prompt)
	inputTokens := 1053
	outputTokens := 0
	
	calculatedCost := pt.CalculateCost(modelName, inputTokens, outputTokens)
	fmt.Printf("Model: %s\n", modelName)
	fmt.Printf("Input tokens: %d\n", inputTokens)
	fmt.Printf("Output tokens: %d\n", outputTokens)
	fmt.Printf("Calculated cost: $%.9f\n", calculatedCost)
	fmt.Printf("Displayed cost: $0.000108\n")
	
	if fmt.Sprintf("%.6f", calculatedCost) == "0.000108" {
		fmt.Println("✅ Cost calculation matches!")
	} else {
		fmt.Println("❌ Cost calculation doesn't match")
	}
}
