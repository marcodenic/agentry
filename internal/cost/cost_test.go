package cost

import "testing"

func TestManagerTrackingAndBudgets(t *testing.T) {
	mgr := New(50, 0.001) // 50 tokens budget, $0.001 budget

	over := mgr.AddModelUsage("openai/gpt-4", 10, 5)
	if over {
		t.Fatalf("expected initial usage to be within budget")
	}

	if tok := mgr.TotalTokens(); tok != 15 {
		t.Fatalf("expected 15 total tokens, got %d", tok)
	}

	totalCost := mgr.TotalCost()
	if totalCost <= 0 || totalCost > 0.001 {
		t.Fatalf("unexpected total cost %.6f", totalCost)
	}

	modelCost := mgr.GetModelCost("openai/gpt-4")
	if modelCost != totalCost {
		t.Fatalf("expected per-model cost %.6f to equal total %.6f", modelCost, totalCost)
	}

	usage := mgr.GetModelUsage("openai/gpt-4")
	if usage.InputTokens != 10 || usage.OutputTokens != 5 {
		t.Fatalf("unexpected usage snapshot: %#v", usage)
	}

	over = mgr.AddModelUsage("openai/gpt-4", 40, 10)
	if !over {
		t.Fatalf("expected budget to be exceeded after additional usage")
	}
	if !mgr.OverBudget() {
		t.Fatalf("OverBudget() should report true once limits are crossed")
	}
}

func TestPricingFuzzyMatching(t *testing.T) {
	pt := NewPricingTable()

	baseCost := pt.CalculateCost("openai/gpt-4", 1000, 2000)
	aliasCost := pt.CalculateCost("openai/gpt-4-latest", 1000, 2000)

	if baseCost == 0 {
		t.Fatalf("expected default pricing for openai/gpt-4 to be non-zero")
	}
	if aliasCost != baseCost {
		t.Fatalf("expected fuzzy match to reuse pricing; base %.6f alias %.6f", baseCost, aliasCost)
	}
}
