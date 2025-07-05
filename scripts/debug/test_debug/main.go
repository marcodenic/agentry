package main

import (
"fmt"
"github.com/marcodenic/agentry/internal/cost"
)

func main() {
m := cost.New(0, 0.00005)
fmt.Printf("Initial budget: $%.9f\n", 0.00005)

m.AddModelUsage("openai/gpt-4", 0, 1)
fmt.Printf("After 1 token: Total cost = $%.9f, Over budget: %v\n", m.TotalCost(), m.OverBudget())

m.AddModelUsage("openai/gpt-4", 0, 1)
fmt.Printf("After 2 tokens: Total cost = $%.9f, Over budget: %v\n", m.TotalCost(), m.OverBudget())
}
