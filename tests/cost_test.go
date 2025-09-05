package tests

import (
	"testing"

	"github.com/marcodenic/agentry/internal/cost"
)

func TestBudgetExceededTokens(t *testing.T) {
	m := cost.New(10, 0)
	if m.AddModelUsage("openai/gpt-4", 2, 3) {
		t.Fatal("unexpected over budget")
	}
	if m.OverBudget() {
		t.Fatal("should not be over budget")
	}
	// Add more model usage to exceed budget (tools are no longer tracked separately)
	if !m.AddModelUsage("openai/gpt-4", 3, 3) {
		t.Fatal("expected budget exceeded")
	}
	if !m.OverBudget() {
		t.Fatal("budget should be exceeded")
	}
}

func TestBudgetExceededDollars(t *testing.T) {
	// Use a budget that's more clearly between 1 and 2 token costs to avoid floating-point precision issues
	// With default pricing of $30/1M tokens, 1 token = $0.00003, 2 tokens = $0.00006
	m := cost.New(0, 0.000045)            // Budget between 1 and 2 token costs
	m.AddModelUsage("openai/gpt-4", 0, 1) // Add 1 output token (~$0.00003)
	cost1 := m.TotalCost()
	t.Logf("Cost after 1 token: %f, Budget: %f", cost1, 0.000045)
	if m.OverBudget() {
		t.Fatalf("should not exceed yet: cost=%f, budget=%f", cost1, 0.000045)
	}
	m.AddModelUsage("openai/gpt-4", 0, 1) // Add another output token (~$0.00006 total)
	cost2 := m.TotalCost()
	t.Logf("Cost after 2 tokens: %f, Budget: %f", cost2, 0.000045)
	if !m.OverBudget() {
		t.Fatalf("should exceed dollar budget: cost=%f, budget=%f", cost2, 0.000045)
	}
}

func TestModelSpecificPricing(t *testing.T) {
	m := cost.New(0, 0)

	// Add usage for expensive model
	m.AddModelUsage("openai/gpt-4", 1000, 1000)
	cost1 := m.GetModelCost("openai/gpt-4")

	// Add usage for cheaper model
	m.AddModelUsage("openai/gpt-4o-mini", 1000, 1000)
	cost2 := m.GetModelCost("openai/gpt-4o-mini")

	if cost1 <= cost2 {
		t.Fatal("openai/gpt-4 should be more expensive than openai/gpt-4o-mini")
	}
}

func TestTokenUsageTracking(t *testing.T) {
	m := cost.New(0, 0)

	// Add usage for a model
	m.AddModelUsage("openai/gpt-4", 100, 200)
	m.AddModelUsage("openai/gpt-4", 50, 75)

	usage := m.GetModelUsage("openai/gpt-4")
	if usage.InputTokens != 150 {
		t.Fatalf("expected 150 input tokens, got %d", usage.InputTokens)
	}
	if usage.OutputTokens != 275 {
		t.Fatalf("expected 275 output tokens, got %d", usage.OutputTokens)
	}
}
