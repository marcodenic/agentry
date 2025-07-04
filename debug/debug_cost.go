package main

import (
"fmt"
"github.com/marcodenic/agentry/internal/cost"
)

func main() {
	m := cost.New(0, 0.00005)             // Budget of $0.00005
	fmt.Printf("Initial budget: $%.9f\n", 0.00005)
	
	// Add first token
	m.AddModelUsage("openai/gpt-4", 0, 1)
	totalCost1 := m.TotalCost()
	fmt.Printf("After 1 token: Total cost = $%.9f, Over budget: %v\n", totalCost1, m.OverBudget())
	
	// Add second token
	m.AddModelUsage("openai/gpt-4", 0, 1)
	totalCost2 := m.TotalCost()
	fmt.Printf("After 2 tokens: Total cost = $%.9f, Over budget: %v\n", totalCost2, m.OverBudget())
	
	// Check individual model cost
	modelCost := m.GetModelCost("openai/gpt-4")
	fmt.Printf("Model cost for openai/gpt-4: $%.9f\n", modelCost)
	
	// Also check the pricing table directly
	pt := cost.NewPricingTable()
	costPerToken := pt.CalculateCost("openai/gpt-4", 0, 2)
	fmt.Printf("Direct pricing calculation for 2 tokens: $%.9f\n", costPerToken)
}
