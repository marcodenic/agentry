package main

import (
	"fmt"

	"github.com/marcodenic/agentry/internal/cost"
)

func main() {
	fmt.Println("=== Cost Stability Test ===")
	
	// Create a cost manager
	manager := cost.New(0, 0.0)
	
	// Add some model usage
	manager.AddModelUsage("gpt-4", 1000, 2000)
	manager.AddModelUsage("gpt-3.5-turbo", 500, 1000)
	
	fmt.Printf("Initial cost: $%.8f\n", manager.TotalCost())
	
	// Test stability by calling TotalCost() many times
	fmt.Println("Testing cost stability over 100 calls...")
	
	var costs []float64
	for i := 0; i < 100; i++ {
		cost := manager.TotalCost()
		costs = append(costs, cost)
		if i%10 == 0 {
			fmt.Printf("Call %d: $%.8f\n", i, cost)
		}
	}
	
	// Check if all costs are the same
	allSame := true
	firstCost := costs[0]
	for _, cost := range costs[1:] {
		if cost != firstCost {
			allSame = false
			break
		}
	}
	
	if allSame {
		fmt.Println("✅ SUCCESS: All cost calls returned the same value")
	} else {
		fmt.Println("❌ FAILURE: Cost values varied between calls")
	}
	
	// Test with concurrent access
	fmt.Println("\nTesting concurrent access...")
	done := make(chan float64, 10)
	
	for i := 0; i < 10; i++ {
		go func() {
			cost := manager.TotalCost()
			done <- cost
		}()
	}
	
	var concurrentCosts []float64
	for i := 0; i < 10; i++ {
		cost := <-done
		concurrentCosts = append(concurrentCosts, cost)
	}
	
	// Check if all concurrent costs are the same
	allConcurrentSame := true
	firstConcurrentCost := concurrentCosts[0]
	for _, cost := range concurrentCosts[1:] {
		if cost != firstConcurrentCost {
			allConcurrentSame = false
			break
		}
	}
	
	if allConcurrentSame {
		fmt.Println("✅ SUCCESS: All concurrent cost calls returned the same value")
	} else {
		fmt.Println("❌ FAILURE: Concurrent cost values varied")
	}
	
	// Test token count stability
	fmt.Println("\nTesting token count stability...")
	
	var tokenCounts []int
	for i := 0; i < 50; i++ {
		tokens := manager.TotalTokens()
		tokenCounts = append(tokenCounts, tokens)
	}
	
	allTokensSame := true
	firstTokens := tokenCounts[0]
	for _, tokens := range tokenCounts[1:] {
		if tokens != firstTokens {
			allTokensSame = false
			break
		}
	}
	
	if allTokensSame {
		fmt.Printf("✅ SUCCESS: All token counts returned the same value: %d\n", firstTokens)
	} else {
		fmt.Println("❌ FAILURE: Token counts varied between calls")
	}
	
	// Test tool cost tracking (should be no-op now)
	fmt.Println("\nTesting deprecated tool cost tracking...")
	
	initialCost := manager.TotalCost()
	manager.AddTool("test-tool", 100)
	afterToolCost := manager.TotalCost()
	
	if initialCost == afterToolCost {
		fmt.Println("✅ SUCCESS: AddTool() is properly deprecated and doesn't affect cost")
	} else {
		fmt.Printf("❌ FAILURE: AddTool() still affects cost: $%.8f -> $%.8f\n", initialCost, afterToolCost)
	}
	
	fmt.Println("\n=== Test Complete ===")
}
