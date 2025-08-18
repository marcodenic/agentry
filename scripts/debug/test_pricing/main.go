package main

import (
"fmt"
"github.com/marcodenic/agentry/internal/cost"
)

func main() {
// Create pricing table
pt := cost.NewPricingTable()

// Check OpenAI GPT-4 pricing
pricing, found := pt.GetPricingByModelName("openai/gpt-4")
if found {
fmt.Printf("openai/gpt-4 pricing: input=%.6f, output=%.6f\n", pricing.InputPrice, pricing.OutputPrice)

// Calculate cost for 1 output token
cost1 := pt.CalculateCost("openai/gpt-4", 0, 1)
fmt.Printf("Cost for 1 output token: $%.9f\n", cost1)

// Calculate cost for 2 output tokens
cost2 := pt.CalculateCost("openai/gpt-4", 0, 2)
fmt.Printf("Cost for 2 output tokens: $%.9f\n", cost2)

// Calculate cost for 3 output tokens
cost3 := pt.CalculateCost("openai/gpt-4", 0, 3)
fmt.Printf("Cost for 3 output tokens: $%.9f\n", cost3)
} else {
fmt.Println("openai/gpt-4 not found")
}
}
