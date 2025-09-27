package memory

import (
	"testing"

	"github.com/marcodenic/agentry/internal/model"
)

func TestInMemoryAddStepCapsHistory(t *testing.T) {
	mem := NewInMemory()
	for i := 0; i < 30; i++ {
		mem.AddStep(Step{Input: "in", Output: "out"})
	}
	hist := mem.History()
	if len(hist) != 20 {
		t.Fatalf("expected history capped at 20, got %d", len(hist))
	}
	memSteps := mem.steps
	if len(memSteps) != 20 {
		t.Fatalf("expected internal steps capped, got %d", len(memSteps))
	}
}

func TestInMemoryHistoryReturnsCopy(t *testing.T) {
	mem := NewInMemory()
	mem.AddStep(Step{Input: "a", ToolCalls: []model.ToolCall{{Name: "tool"}}})

	hist := mem.History()
	if len(hist) != 1 {
		t.Fatalf("expected history length 1, got %d", len(hist))
	}

	hist[0].Input = "mutated"
	if mem.steps[0].Input == "mutated" {
		t.Fatalf("expected mutation of history copy to not affect original")
	}
}

func TestSetHistoryCopiesInput(t *testing.T) {
	mem := NewInMemory()
	input := []Step{{Input: "one"}, {Input: "two"}}
	mem.SetHistory(input)

	if len(mem.steps) != 2 {
		t.Fatalf("expected steps to match input length")
	}

	input[0].Input = "changed"
	if mem.steps[0].Input == "changed" {
		t.Fatalf("expected SetHistory to copy slice")
	}
}
