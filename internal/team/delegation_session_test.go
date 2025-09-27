package team

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/tool"
)

func newTestParentAgent() *core.Agent {
	tool.SetPermissions(nil)
	return core.New(model.NewMock(), "mock", tool.Registry{}, memory.NewInMemory(), memory.NewInMemoryVector(), nil)
}

func TestCallSpawnsAgentAndAugmentsInput(t *testing.T) {
	t.Setenv("AGENTRY_TUI_MODE", "1")
	parent := newTestParentAgent()

	teamInstance, err := NewTeam(parent, 3, "squad")
	if err != nil {
		t.Fatalf("NewTeam: %v", err)
	}

	events := []WorkspaceEvent{{
		AgentID:     "reviewer",
		Type:        "task_completed",
		Description: "Document updated",
		Timestamp:   time.Now(),
	}}
	teamInstance.SetSharedData("workspace_events", events)

	originalRunAgent := runAgentFn
	defer func() { runAgentFn = originalRunAgent }()

	var capturedInput string
	runAgentFn = func(ctx context.Context, ag *core.Agent, input, name string, peers []string) (string, error) {
		capturedInput = input
		if name != "coder" {
			t.Fatalf("expected agent name 'coder', got %s", name)
		}
		return "all good", nil
	}

	result, err := teamInstance.Call(context.Background(), "coder", "Draft the plan ")
	if err != nil {
		t.Fatalf("Call returned error: %v", err)
	}
	if result != "all good" {
		t.Fatalf("unexpected result: %s", result)
	}

	if teamInstance.GetAgent("coder") == nil {
		t.Fatalf("expected spawned agent to be registered")
	}

	if !strings.Contains(capturedInput, "RECENT WORKSPACE EVENTS") {
		t.Fatalf("expected workspace context appended, got %q", capturedInput)
	}
	if !strings.Contains(capturedInput, "reviewer | task_completed") {
		t.Fatalf("expected workspace event details, got %q", capturedInput)
	}

	val, ok := teamInstance.GetSharedData("last_result_coder")
	if !ok || val.(string) != "all good" {
		t.Fatalf("expected shared last result, got %#v", val)
	}
	task, ok := teamInstance.GetSharedData("last_task_coder")
	if !ok || task.(string) != "Draft the plan " {
		t.Fatalf("expected shared last task, got %#v", task)
	}
}

func TestCallHonoursDelegationTimeout(t *testing.T) {
	t.Setenv("AGENTRY_TUI_MODE", "")
	t.Setenv("AGENTRY_DELEGATION_TIMEOUT", "50ms")

	parent := newTestParentAgent()
	teamInstance, err := NewTeam(parent, 3, "ops")
	if err != nil {
		t.Fatalf("NewTeam: %v", err)
	}

	originalRunAgent := runAgentFn
	defer func() { runAgentFn = originalRunAgent }()

	start := time.Now()
	var deadline time.Time
	runAgentFn = func(ctx context.Context, ag *core.Agent, input, name string, peers []string) (string, error) {
		var ok bool
		deadline, ok = ctx.Deadline()
		if !ok {
			t.Fatalf("expected context deadline")
		}
		return "", context.DeadlineExceeded
	}

	output, err := teamInstance.Call(context.Background(), "analyst", "Investigate incident")
	if err == nil {
		t.Fatalf("expected timeout error")
	}
	if output != "" {
		t.Fatalf("expected empty output, got %q", output)
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("expected timeout message, got %v", err)
	}

	if deadline.IsZero() {
		t.Fatalf("expected deadline to be captured")
	}
	elapsed := deadline.Sub(start)
	if elapsed < 40*time.Millisecond || elapsed > 200*time.Millisecond {
		t.Fatalf("expected deadline near 50ms, got %v", elapsed)
	}

	eventsVal, ok := teamInstance.GetSharedData("workspace_events")
	if !ok {
		t.Fatalf("expected workspace events recorded")
	}
	events := eventsVal.([]WorkspaceEvent)
	foundTimeout := false
	for _, ev := range events {
		if ev.Type == "delegation_timeout" {
			foundTimeout = true
			break
		}
	}
	if !foundTimeout {
		t.Fatalf("expected delegation_timeout workspace event, got %#v", events)
	}
}
