package core

import (
	"context"
	"errors"
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
	if hadErrors {
		t.Fatalf("did not expect hadErrors when TreatErrorsAsResults is true")
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
