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
	m := cost.New(0, 0.00005)             // Budget of $0.00005
	m.AddModelUsage("openai/gpt-4", 0, 1) // Add 1 output token (~$0.00003)
	if m.OverBudget() {
		t.Fatal("should not exceed yet")
	}
	m.AddModelUsage("openai/gpt-4", 0, 1) // Add another output token (~$0.00006 total)
	if !m.OverBudget() {
		t.Fatal("should exceed dollar budget")
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
