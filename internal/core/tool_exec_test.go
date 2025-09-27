package core

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/tool"
)

func TestToolExecutorTreatsErrorsAsResults(t *testing.T) {
	tool.SetPermissions(nil)

	boom := errors.New("boom")
	reg := tool.Registry{
		"fail": tool.New("fail", "", func(context.Context, map[string]any) (string, error) {
			return "failure", boom
		}),
	}

	ag := &Agent{
		Tools:         reg,
		JSONValidator: NewJSONValidator(),
		ErrorHandling: ErrorHandlingConfig{
			TreatErrorsAsResults: true,
			IncludeErrorContext:  false,
			MaxErrorRetries:      3,
		},
	}

	exec := newToolExecutor(ag)
	step := memory.Step{ToolResults: map[string]string{}}
	msgs, hadErrors, err := exec.execute(context.Background(), []model.ToolCall{{ID: "1", Name: "fail", Arguments: []byte("{}")}}, step)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hadErrors {
		t.Fatalf("expected hadErrors to be true to signal soft failure")
	}
	if len(msgs) != 1 {
		t.Fatalf("expected a single tool message, got %d", len(msgs))
	}
	expected := "Error executing tool 'fail': boom"
	if msgs[0].Content != expected {
		t.Fatalf("unexpected tool content: %q", msgs[0].Content)
	}
	if got := step.ToolResults["1"]; got != expected {
		t.Fatalf("expected tool result to be recorded, got %q", got)
	}
}

func TestToolExecutorPropagatesErrorsWhenDisabled(t *testing.T) {
	tool.SetPermissions(nil)

	boom := errors.New("boom")
	reg := tool.Registry{
		"fail": tool.New("fail", "", func(context.Context, map[string]any) (string, error) {
			return "", boom
		}),
	}

	ag := &Agent{
		Tools:         reg,
		JSONValidator: NewJSONValidator(),
		ErrorHandling: ErrorHandlingConfig{
			TreatErrorsAsResults: false,
			IncludeErrorContext:  false,
			MaxErrorRetries:      1,
		},
	}

	exec := newToolExecutor(ag)
	step := memory.Step{ToolResults: map[string]string{}}
	_, hadErrors, err := exec.execute(context.Background(), []model.ToolCall{{ID: "1", Name: "fail", Arguments: []byte("{}")}}, step)
	if !hadErrors {
		t.Fatalf("expected hadErrors to be true")
	}
	if err == nil {
		t.Fatalf("expected error to be returned when TreatErrorsAsResults is false")
	}
	if _, ok := step.ToolResults["1"]; ok {
		t.Fatalf("did not expect tool result to be recorded on hard failure")
	}
}

func TestNewToolExecutorRespectsTUIMode(t *testing.T) {
	ag := &Agent{JSONValidator: NewJSONValidator(), ErrorHandling: DefaultErrorHandling()}

	t.Run("noop notifier when TUI enabled", func(t *testing.T) {
		t.Setenv("AGENTRY_TUI_MODE", "1")
		exec := newToolExecutor(ag)
		if _, ok := exec.note.(noopToolNotifier); !ok {
			t.Fatalf("expected noopToolNotifier when AGENTRY_TUI_MODE=1, got %T", exec.note)
		}
	})

	t.Run("stderr notifier by default", func(t *testing.T) {
		t.Setenv("AGENTRY_TUI_MODE", "")
		exec := newToolExecutor(ag)
		if _, ok := exec.note.(stderrToolNotifier); !ok {
			t.Fatalf("expected stderrToolNotifier when AGENTRY_TUI_MODE unset, got %T", exec.note)
		}
	})
}

func TestToolExecutorReportsValidationErrors(t *testing.T) {
	tool.SetPermissions(nil)

	reg := tool.Registry{
		"echo": tool.New("echo", "", func(context.Context, map[string]any) (string, error) {
			return "should-not-run", nil
		}),
	}

	ag := &Agent{
		Tools:         reg,
		JSONValidator: NewJSONValidator(),
		ErrorHandling: DefaultErrorHandling(),
	}

	exec := newToolExecutor(ag)
	step := memory.Step{ToolResults: map[string]string{}}
	msgs, hadErrors, err := exec.execute(context.Background(), []model.ToolCall{{
		ID:        "call-1",
		Name:      "echo",
		Arguments: []byte(`{"__proto__":"hack"}`),
	}}, step)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hadErrors {
		t.Fatalf("expected hadErrors to be true for validation failure")
	}
	if len(msgs) != 1 {
		t.Fatalf("expected single tool error message, got %d", len(msgs))
	}
	if _, ok := step.ToolResults["call-1"]; !ok {
		t.Fatalf("expected step records for validation failure")
	}
	if got := msgs[0].Content; got == "" || !containsAll(got, "Invalid tool arguments", "disallowed key") {
		t.Fatalf("unexpected validation message: %q", got)
	}
}

func TestToolExecutorTerminalToolSuccess(t *testing.T) {
	tool.SetPermissions(nil)

	terminal := tool.MarkTerminal(tool.New("terminal", "", func(context.Context, map[string]any) (string, error) {
		return "done", nil
	}))
	reg := tool.Registry{"terminal": terminal}

	ag := &Agent{
		Tools:         reg,
		JSONValidator: NewJSONValidator(),
		ErrorHandling: DefaultErrorHandling(),
	}

	exec := newToolExecutor(ag)
	step := memory.Step{ToolResults: map[string]string{}}
	msgs, hadErrors, err := exec.execute(context.Background(), []model.ToolCall{{
		ID:        "1",
		Name:      "terminal",
		Arguments: []byte("{}"),
	}}, step)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hadErrors {
		t.Fatalf("did not expect errors for successful terminal tool")
	}
	if len(msgs) != 1 || msgs[0].Content != "done" {
		t.Fatalf("unexpected tool output: %#v", msgs)
	}
	if recorded := step.ToolResults["1"]; recorded != "done" {
		t.Fatalf("expected tool result recorded, got %q", recorded)
	}
}

func containsAll(s string, substrings ...string) bool {
	for _, sub := range substrings {
		if !strings.Contains(s, sub) {
			return false
		}
	}
	return true
}
