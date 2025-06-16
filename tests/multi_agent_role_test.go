package tests

import (
	"testing"

	"github.com/marcodenic/agentry/internal/converse"
	"github.com/marcodenic/agentry/internal/memory"
)

func TestRolesStayAssistant(t *testing.T) {
	mem := memory.NewInMemory()
	mem.AddStep(memory.Step{Output: "hi"})
	msgs := converse.BuildMessages(mem.History(), "", "Agent1", []string{"Agent1", "Agent2"})
	if len(msgs) < 2 {
		t.Fatalf("expected at least two messages")
	}
	if msgs[1].Role != "assistant" {
		t.Fatalf("expected assistant got %s", msgs[1].Role)
	}
}
