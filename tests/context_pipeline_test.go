package tests

import (
	"os"
	"strings"
	"testing"

	agentctx "github.com/marcodenic/agentry/internal/context"
	"github.com/marcodenic/agentry/internal/memory"
)

func TestProviderIncludesHistoryAndInput(t *testing.T) {
	prov := agentctx.Provider{
		Prompt:  "system",
		History: []memory.Step{{Output: "hello"}},
	}
	msgs := prov.Provide("world")
	if len(msgs) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(msgs))
	}
	if msgs[0].Role != "system" || msgs[1].Role != "assistant" || msgs[2].Role != "user" {
		t.Fatalf("unexpected roles: %#v", msgs)
	}
	if msgs[1].Content != "hello" || msgs[2].Content != "world" {
		t.Fatalf("unexpected content: %#v", msgs)
	}
}

func TestBudgetTrimsOldMessages(t *testing.T) {
	os.Setenv("AGENTRY_CONTEXT_MAX_TOKENS", "300")
	os.Setenv("AGENTRY_CONTEXT_RESERVE_OUTPUT", "256")
	defer os.Unsetenv("AGENTRY_CONTEXT_MAX_TOKENS")
	defer os.Unsetenv("AGENTRY_CONTEXT_RESERVE_OUTPUT")

	hist := []memory.Step{
		{Output: strings.Repeat("a ", 200)},
		{Output: strings.Repeat("b ", 200)},
	}
	prov := agentctx.Provider{Prompt: "system", History: hist}
	budget := agentctx.Budget{ModelName: "gpt-4o"}
	asm := agentctx.Assembler{Provider: prov, Budget: budget}
	msgs := asm.Assemble("input")

	if len(msgs) != 2 {
		t.Fatalf("expected history trimmed to 2 messages, got %d", len(msgs))
	}
	if msgs[0].Role != "system" || msgs[1].Role != "user" {
		t.Fatalf("unexpected roles after trim: %#v", msgs)
	}
}
