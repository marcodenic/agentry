package context

import (
	"os"
	"strings"
	"testing"

	"github.com/marcodenic/agentry/internal/model"
)

func TestBudgetCountsToolCallTokens(t *testing.T) {
	os.Setenv("AGENTRY_CONTEXT_MAX_TOKENS", "800")
	defer os.Unsetenv("AGENTRY_CONTEXT_MAX_TOKENS")

	budget := Budget{ModelName: "gpt-4o"}
	bigArgs := strings.Repeat("a ", 400)
	msgs := []model.ChatMessage{
		{Role: "system", Content: "s"},
		{Role: "assistant", ToolCalls: []model.ToolCall{{ID: "1", Name: "echo", Arguments: []byte(bigArgs)}}},
		{Role: "user", Content: "hi"},
	}
	trimmed := budget.Apply(msgs)
	if len(trimmed) != 2 || trimmed[0].Role != "system" || trimmed[1].Role != "user" {
		t.Fatalf("unexpected messages after trim: %#v", trimmed)
	}
}
