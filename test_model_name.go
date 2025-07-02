package main

import (
	"fmt"
	"strings"
)

// Copy of the specificity functions to test independently
func shouldUpdateModelName(currentName, newName string) bool {
	if currentName == "" {
		return true
	}

	currentScore := getModelNameSpecificity(currentName)
	newScore := getModelNameSpecificity(newName)

	return newScore > currentScore
}

func getModelNameSpecificity(name string) int {
	if name == "" {
		return 0
	}

	commonRuleNames := map[string]int{
		"openai":    1,
		"anthropic": 1,
		"mock":      1,
		"ollama":    1,
		"gemini":    1,
	}

	if score, exists := commonRuleNames[name]; exists {
		return score
	}

	specificity := 2

	if strings.Contains(name, "-") {
		specificity += 2
	}

	if strings.Contains(name, "mini") || strings.Contains(name, "max") ||
		strings.Contains(name, "ultra") || strings.Contains(name, "plus") {
		specificity += 1
	}

	for _, char := range name {
		if char >= '0' && char <= '9' {
			specificity += 1
			break
		}
	}

	return specificity
}

func main() {
	// Test cases
	testCases := []struct {
		current, new string
		shouldUpdate bool
	}{
		{"", "openai", true},             // Empty to rule name
		{"", "gpt-4o-mini", true},        // Empty to model name
		{"openai", "gpt-4o-mini", true},  // Rule name to model name
		{"gpt-4o-mini", "openai", false}, // Model name to rule name
		{"gpt-4", "gpt-4o-mini", true},   // Less specific to more specific
		{"gpt-4o-mini", "gpt-4", false},  // More specific to less specific
		{"claude-3", "anthropic", false}, // Model name to rule name
	}

	fmt.Println("Testing model name specificity logic:")
	fmt.Println("=====================================")

	for i, tc := range testCases {
		currentScore := getModelNameSpecificity(tc.current)
		newScore := getModelNameSpecificity(tc.new)
		result := shouldUpdateModelName(tc.current, tc.new)

		status := "✓"
		if result != tc.shouldUpdate {
			status = "✗"
		}

		fmt.Printf("%s Test %d: '%s' (score=%d) -> '%s' (score=%d) = %t (expected %t)\n",
			status, i+1, tc.current, currentScore, tc.new, newScore, result, tc.shouldUpdate)
	}
}
