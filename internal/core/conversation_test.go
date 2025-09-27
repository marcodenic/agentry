package core

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/marcodenic/agentry/internal/cost"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/tool"
	"github.com/marcodenic/agentry/internal/trace"
)

func TestConversationRunCompletesWithToolCall(t *testing.T) {
	t.Setenv("AGENTRY_TUI_MODE", "1")
	t.Setenv("AGENTRY_PLAN_HEURISTIC", "0")
	t.Setenv("AGENTRY_STOP_ON_BUDGET", "0")
	t.Cleanup(func() { tool.SetPermissions(nil) })
	tool.SetPermissions(nil)

	mem := memory.NewInMemory()
	reg := tool.Registry{
		"echo": tool.New("echo", "echo tool", func(ctx context.Context, args map[string]any) (string, error) {
			if args == nil {
				return "", nil
			}
			if v, ok := args["text"].(string); ok {
				return v, nil
			}
			return "", nil
		}),
	}

	agent := New(model.NewMock(), "openai/gpt-4", reg, mem, nil, trace.NewJSONL(io.Discard))
	agent.MaxIter = 5

	out, err := agent.Run(context.Background(), "say hello")
	if err != nil {
		toLogs(t)
		t.Fatalf("Run returned error: %v", err)
	}
	if out != "hello" {
		toLogs(t)
		t.Fatalf("expected final output to be 'hello', got %q", out)
	}

	history := mem.History()
	if len(history) < 2 {
		toLogs(t)
		t.Fatalf("expected at least two history steps (tool + final), got %d", len(history))
	}
	if last := history[len(history)-1]; last.Output != "hello" {
		toLogs(t)
		t.Fatalf("expected last history output to be 'hello', got %q", last.Output)
	}

	hasToolCall := false
	for _, step := range history {
		if len(step.ToolCalls) > 0 {
			hasToolCall = true
			break
		}
	}
	if !hasToolCall {
		toLogs(t)
		t.Fatalf("expected at least one recorded tool call in history")
	}
}

func TestConversationSessionPlanFollowUpInjection(t *testing.T) {
	t.Setenv("AGENTRY_PLAN_HEURISTIC", "1")

	mem := memory.NewInMemory()
	agent := New(model.NewMock(), "openai/gpt-4", tool.Registry{}, mem, nil, trace.NewJSONL(io.Discard))

	session := newConversationSession(agent, context.Background(), "build plan")
	session.specs = []model.ToolSpec{{Name: "echo"}}
	session.msgs = []model.ChatMessage{{Role: "system", Content: "sys"}, {Role: "user", Content: "do work"}}

	completion := model.Completion{Content: "Here is the plan:\n1. Step\n2. Step"}

	out, done, err := session.handleCompletion(completion, "")
	if err != nil {
		t.Fatalf("handleCompletion returned error: %v", err)
	}
	if done {
		t.Fatalf("expected session to continue after plan injection")
	}
	if out != "" {
		t.Fatalf("expected empty output while plan follow-up injected, got %q", out)
	}

	history := mem.History()
	if len(history) != 1 {
		t.Fatalf("expected a single history step to be recorded, got %d", len(history))
	}

	if got := session.msgs[len(session.msgs)-1]; got.Role != "system" || got.Content != session.planFollowUpMessage() {
		t.Fatalf("expected last message to be plan follow-up system prompt, got role=%s content=%q", got.Role, got.Content)
	}
}

func TestTrackRecentToolCallsStopsAfterThreshold(t *testing.T) {
	session := &conversationSession{}
	var recorded bool
	msg, stop := session.trackRecentToolCalls([]model.ToolCall{
		{ID: "1", Name: "echo", Arguments: []byte("hi")},
		{ID: "2", Name: "echo", Arguments: []byte("hi")},
		{ID: "3", Name: "echo", Arguments: []byte("hi")},
	}, func() {
		recorded = true
	})

	if !stop {
		t.Fatalf("expected duplicate detection to stop execution")
	}
	if !strings.Contains(msg, "repeated tool execution (echo)") {
		t.Fatalf("unexpected loop guard message: %q", msg)
	}
	if !recorded {
		t.Fatalf("expected recordStep callback to run when guard trips")
	}
}

func TestHandleCompletionRespectsBudgetStop(t *testing.T) {
	t.Setenv("AGENTRY_STOP_ON_BUDGET", "true")

	mem := memory.NewInMemory()
	agent := New(model.NewMock(), "openai/gpt-4", tool.Registry{}, mem, nil, trace.NewJSONL(io.Discard))
	agent.Cost = cost.New(1, 0)

	session := newConversationSession(agent, context.Background(), "task")
	session.msgs = []model.ChatMessage{{Role: "user", Content: "work"}}

	completion := model.Completion{Content: "done", InputTokens: 2, OutputTokens: 0}

	if _, _, err := session.handleCompletion(completion, ""); err == nil {
		t.Fatalf("expected budget error when token usage exceeds limit")
	}
}

func TestAggregateToolOutputsFiltersAndConcatenates(t *testing.T) {
	msgs := []model.ChatMessage{
		{Role: "tool", Content: "first"},
		{Role: "assistant", Content: "ignore"},
		{Role: "tool", Content: "second"},
		{Role: "tool", Content: "   "},
	}

	got := aggregateToolOutputs(msgs)
	if got != "first\nsecond" {
		t.Fatalf("expected combined tool outputs, got %q", got)
	}
}

func toLogs(t *testing.T) {
	t.Helper()
	t.Log("See stdout/stderr for agent debug output")
}
