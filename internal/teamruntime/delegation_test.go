package teamruntime_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/marcodenic/agentry/internal/teamruntime"
	"github.com/marcodenic/agentry/internal/teamruntime/testsupport"
)

type stubTimer struct{ labels []string }

func (s *stubTimer) Checkpoint(label string) time.Duration {
	s.labels = append(s.labels, label)
	return 0
}

type stubNotifier struct {
	user []string
	file []string
}

func (n *stubNotifier) User(format string, args ...interface{}) {
	n.user = append(n.user, strings.TrimSpace(strings.ReplaceAll(fmt.Sprintf(format, args...), "\n", "")))
}

func (n *stubNotifier) File(format string, args ...interface{}) {
	n.file = append(n.file, fmt.Sprintf(format, args...))
}

func TestDelegationTelemetryFlow(t *testing.T) {
	timer := &stubTimer{}
	logger := &testsupport.StubTeam{}
	notifier := &stubNotifier{}

	tele := teamruntime.NewDelegationTelemetry("coder", "write docs", logger, notifier, timer)

	tele.Start()
	tele.WorkStart()
	tele.LogTaskFile()
	tele.RunAgentStart()
	tele.RunAgentComplete(42 * time.Millisecond)
	success := tele.TimeoutWithWork(time.Minute)
	tele.RecordSuccess("done")

	if len(timer.labels) == 0 {
		t.Fatalf("expected checkpoints recorded")
	}
	if len(logger.Events) == 0 {
		t.Fatalf("expected coordination events recorded")
	}
	if !strings.Contains(success, "completed the work") {
		t.Fatalf("unexpected timeout success message: %s", success)
	}
	if len(notifier.user) == 0 {
		t.Fatalf("expected user notifications")
	}
	if len(notifier.file) == 0 {
		t.Fatalf("expected file notifications")
	}
}

func TestDelegationTelemetryErrors(t *testing.T) {
	logger := &testsupport.StubTeam{}
	tele := teamruntime.NewDelegationTelemetry("ops", "restart", logger, nil, nil)

	timeout := tele.TimeoutWithoutWork(30 * time.Second)
	if !strings.Contains(timeout, "timed out") {
		t.Fatalf("expected timeout message, got %s", timeout)
	}

	tele.RecordFailure(errors.New("boom"))
	if got := logger.Events[len(logger.Events)-1].Type; got != "delegation_failed" {
		t.Fatalf("expected delegation_failed event, got %s", got)
	}
}

func TestBuildWorkspaceContext(t *testing.T) {
	ctx := teamruntime.BuildWorkspaceContext(testsupport.FakeWorkspaceEvents("coder", "Implemented feature"))
	if !strings.Contains(ctx, "coder | task_completed") {
		t.Fatalf("expected agent and type, got %s", ctx)
	}
	if !strings.Contains(ctx, "feedback") {
		t.Fatalf("expected feedback event, got %s", ctx)
	}
}

func TestIntegrationWithStubTeam(t *testing.T) {
	teamruntime.SetLogToFile(func(level, format string, args ...interface{}) {})
	t.Cleanup(func() { teamruntime.SetLogToFile(nil) })

	logger := &testsupport.StubTeam{}
	tele := teamruntime.NewDelegationTelemetry("analyst", "summarize", logger, nil, &stubTimer{})
	tele.Start()
	tele.RecordSuccess("ok")

	events := logger.EventsOfType("delegation_success")
	if len(events) != 1 {
		t.Fatalf("expected success event, got %d", len(events))
	}
}

func TestPublishWorkspaceEvent(t *testing.T) {
	teamruntime.SetLogToFile(func(level, format string, args ...interface{}) {})
	t.Cleanup(func() { teamruntime.SetLogToFile(nil) })

	stub := &testsupport.StubTeam{}
	stub.PublishWorkspaceEvent("agent_0", "delegation_started", "Delegated to coder")

	events := stub.EventsOfType("workspace_event")
	if len(events) != 1 {
		t.Fatalf("expected workspace event logged, got %d", len(events))
	}
	if !strings.Contains(events[0].Content, "delegation_started") {
		t.Fatalf("unexpected workspace event content: %s", events[0].Content)
	}
}
