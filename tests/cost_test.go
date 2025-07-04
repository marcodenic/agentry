package tests

import (
	"testing"

	"github.com/marcodenic/agentry/internal/cost"
)

func TestBudgetExceededTokens(t *testing.T) {
	m := cost.New(10, 0)
	if m.AddModelUsage("gpt-4", 2, 3) {
		t.Fatal("unexpected over budget")
	}
	if m.OverBudget() {
		t.Fatal("should not be over budget")
	}
	if !m.AddTool("echo", 6) {
		t.Fatal("expected budget exceeded")
	}
	if !m.OverBudget() {
		t.Fatal("budget should be exceeded")
	}
}

func TestBudgetExceededDollars(t *testing.T) {
	m := cost.New(0, 0.0001)       // Budget of $0.0001
	m.AddModelUsage("gpt-4", 0, 1) // Add 1 output token (~$0.00006)
	if m.OverBudget() {
		t.Fatal("should not exceed yet")
	}
	m.AddModelUsage("gpt-4", 0, 1) // Add another output token (~$0.00012 total)
	if !m.OverBudget() {
		t.Fatal("should exceed dollar budget")
	}
}

func TestModelSpecificPricing(t *testing.T) {
	m := cost.New(0, 0)

	// Add usage for expensive model
	m.AddModelUsage("gpt-4", 1000, 1000)
	cost1 := m.GetModelCost("gpt-4")

	// Add usage for cheaper model
	m.AddModelUsage("gpt-4o-mini", 1000, 1000)
	cost2 := m.GetModelCost("gpt-4o-mini")

	if cost1 <= cost2 {
		t.Fatal("gpt-4 should be more expensive than gpt-4o-mini")
	}
}

func TestTokenUsageTracking(t *testing.T) {
	m := cost.New(0, 0)

	// Add usage for a model
	m.AddModelUsage("gpt-4", 100, 200)
	m.AddModelUsage("gpt-4", 50, 75)

	usage := m.GetModelUsage("gpt-4")
	if usage.InputTokens != 150 {
		t.Fatalf("expected 150 input tokens, got %d", usage.InputTokens)
	}
	if usage.OutputTokens != 275 {
		t.Fatalf("expected 275 output tokens, got %d", usage.OutputTokens)
	}
}
