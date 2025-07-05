package main

import (
"fmt"
"github.com/marcodenic/agentry/internal/cost"
)

func main() {
// Exactly replicate the test
m := cost.New(0, 0.00005)             // Budget of $0.00005
fmt.Printf("Initial budget: $%.9f\n", 0.00005)

m.AddModelUsage("openai/gpt-4", 0, 1) // Add 1 output token (~$0.00003)
fmt.Printf("After 1 token: Total cost = $%.9f, Over budget: %v\n", m.TotalCost(), m.OverBudget())
if m.OverBudget() {
fmt.Println("Test would fail at first check")
}

m.AddModelUsage("openai/gpt-4", 0, 1) // Add another output token (~$0.00006 total)
fmt.Printf("After 2 tokens: Total cost = $%.9f, Over budget: %v\n", m.TotalCost(), m.OverBudget())
if !m.OverBudget() {
fmt.Println("Test would fail at second check")
}

// Check exact comparison
budget := 0.00005
totalCost := m.TotalCost()
fmt.Printf("Budget: %.9f, Total cost: %.9f, Comparison: %v\n", budget, totalCost, totalCost > budget)
}
