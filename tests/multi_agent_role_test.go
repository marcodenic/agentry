package tests

import (
	"testing"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/router"
)

func TestRolesStayAssistant(t *testing.T) {
	mem := memory.NewInMemory()
	rules := router.Rules{{IfContains: []string{""}}}
	ag := core.NewNamed("Agent1", rules, nil, mem, nil)
	ag.PeerNames = []string{"Agent1", "Agent2"}
	mem.AddStep(memory.Step{Speaker: "Agent2", Output: "hi"})
	msgs := core.BuildMessages(mem.History(), "", "Agent1", ag.PeerNames, "")
	if len(msgs) < 2 {
		t.Fatalf("expected at least two messages")
	}
	if msgs[1].Role != "assistant" {
		t.Fatalf("expected assistant got %s", msgs[1].Role)
	}
}
