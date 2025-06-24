package tests

import (
	"testing"

	"github.com/marcodenic/agentry/internal/cost"
)

func TestBudgetExceededTokens(t *testing.T) {
	m := cost.New(10, 0)
	if m.AddModel("gpt", 5) {
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
	m := cost.New(0, 0.000004)
	m.AddModel("gpt", 1)
	if m.OverBudget() {
		t.Fatal("should not exceed yet")
	}
	m.AddModel("gpt", 1)
	if !m.OverBudget() {
		t.Fatal("should exceed dollar budget")
	}
}
