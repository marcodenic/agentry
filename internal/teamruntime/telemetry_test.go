package teamruntime

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

type stubNotifier struct {
	userMsgs []string
	fileMsgs []string
}

func (s *stubNotifier) User(format string, args ...interface{}) {
	s.userMsgs = append(s.userMsgs, fmt.Sprintf(format, args...))
}

func (s *stubNotifier) File(format string, args ...interface{}) {
	s.fileMsgs = append(s.fileMsgs, fmt.Sprintf(format, args...))
}

type stubLogger struct {
	events []struct {
		eventType string
		from      string
		to        string
		content   string
	}
}

func (s *stubLogger) LogCoordinationEvent(eventType, from, to, content string, _ map[string]interface{}) {
	s.events = append(s.events, struct {
		eventType string
		from      string
		to        string
		content   string
	}{eventType: eventType, from: from, to: to, content: content})
}

type stubTimer struct{ checkpoints []string }

func (s *stubTimer) Checkpoint(label string) time.Duration {
	s.checkpoints = append(s.checkpoints, label)
	return 0
}

func TestBuildWorkspaceContext(t *testing.T) {
	events := []WorkspaceEvent{
		{AgentID: "coder", Type: "task_completed", Description: "Implemented feature", Timestamp: time.Date(2024, 1, 2, 15, 4, 5, 0, time.UTC)},
	}
	ctx := BuildWorkspaceContext(events)
	if !strings.Contains(ctx, "coder | task_completed") {
		t.Fatalf("expected agent and type in context, got %q", ctx)
	}
	if !strings.Contains(ctx, "Implemented feature") {
		t.Fatalf("expected description in context, got %q", ctx)
	}
}

func TestDelegationTelemetryStartRecordsEvent(t *testing.T) {
	notifier := &stubNotifier{}
	logger := &stubLogger{}
	timer := &stubTimer{}

	tele := NewDelegationTelemetry("coder", "write docs", logger, notifier, timer)
	tele.Start()

	if len(logger.events) != 1 {
		t.Fatalf("expected 1 coordination event, got %d", len(logger.events))
	}
	if logger.events[0].eventType != "delegation" {
		t.Fatalf("unexpected event type: %v", logger.events[0].eventType)
	}
	if len(notifier.userMsgs) == 0 {
		t.Fatalf("expected notifier to receive user message")
	}
	if len(timer.checkpoints) == 0 || timer.checkpoints[0] != "coordination logged" {
		t.Fatalf("missing coordination checkpoint: %+v", timer.checkpoints)
	}
}

func TestDelegationTelemetryTimeouts(t *testing.T) {
	notifier := &stubNotifier{}
	logger := &stubLogger{}
	tele := NewDelegationTelemetry("ops", "restart service", logger, notifier, nil)

	msg := tele.TimeoutWithoutWork(2 * time.Minute)
	if !strings.Contains(msg, "ops") {
		t.Fatalf("expected timeout message to mention agent, got %q", msg)
	}
	if logger.events[len(logger.events)-1].eventType != "delegation_timeout" {
		t.Fatalf("expected timeout event, got %v", logger.events)
	}

	successMsg := tele.TimeoutWithWork(time.Minute)
	if !strings.Contains(successMsg, "completed the work") {
		t.Fatalf("expected success timeout message, got %q", successMsg)
	}
	if logger.events[len(logger.events)-1].eventType != "delegation_success_timeout" {
		t.Fatalf("expected success timeout event, got %v", logger.events)
	}
}

func TestDelegationTelemetryFailure(t *testing.T) {
	notifier := &stubNotifier{}
	logger := &stubLogger{}
	tele := NewDelegationTelemetry("analyst", "parse report", logger, notifier, nil)

	tele.RecordFailure(fmt.Errorf("boom"))
	if logger.events[len(logger.events)-1].eventType != "delegation_failed" {
		t.Fatalf("expected failure event, got %v", logger.events)
	}
}

func TestDelegationTelemetrySuccess(t *testing.T) {
	notifier := &stubNotifier{}
	logger := &stubLogger{}
	tele := NewDelegationTelemetry("analyst", "parse report", logger, notifier, nil)

	tele.RecordSuccess("done")
	if logger.events[len(logger.events)-1].eventType != "delegation_success" {
		t.Fatalf("expected success event, got %v", logger.events)
	}
	if len(notifier.userMsgs) == 0 {
		t.Fatalf("expected success notifier message")
	}
}

func TestTeamNotifierHonorsTUI(t *testing.T) {
	t.Setenv("AGENTRY_TUI_MODE", "1")
	notifier := NewNotifier()
	notifier.User("message that should not print")
	t.Setenv("AGENTRY_TUI_MODE", "0")
	notifier.User("hello %s", "world")
}

func TestTeamNotifierFileWrites(t *testing.T) {
	original := logToFileFunc
	defer func() { logToFileFunc = original }()

	var wrote []string
	logToFileFunc = func(level, format string, args ...interface{}) {
		wrote = append(wrote, fmt.Sprintf(format, args...))
	}

	notifier := NewNotifier()
	notifier.File("stored %s", "value")

	if len(wrote) != 1 || wrote[0] != "stored value" {
		t.Fatalf("expected file log entry, got %+v", wrote)
	}
}
