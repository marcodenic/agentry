package team

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPublishWorkspaceEventStoresAndTrims(t *testing.T) {
	t.Setenv("AGENTRY_TUI_MODE", "1")
	tm := &Team{
		name:         "squad",
		sharedMemory: make(map[string]interface{}),
		coordination: make([]CoordinationEvent, 0),
	}

	for i := 0; i < 55; i++ {
		tm.PublishWorkspaceEvent("agent", "event", fmt.Sprintf("event-%d", i), nil)
	}

	val, ok := tm.GetSharedData("workspace_events")
	if !ok {
		t.Fatalf("expected workspace events saved")
	}
	events := val.([]WorkspaceEvent)
	if len(events) != maxWorkspaceEvents {
		t.Fatalf("expected %d events, got %d", maxWorkspaceEvents, len(events))
	}
	if events[0].Description != "event-5" {
		t.Fatalf("expected oldest retained event to be event-5, got %s", events[0].Description)
	}
	workspaceEvents := 0
	for _, ev := range tm.coordination {
		if ev.Type == "workspace_event" {
			workspaceEvents++
		}
	}
	if workspaceEvents != 55 {
		t.Fatalf("expected 55 workspace event entries, got %d", workspaceEvents)
	}
}

func TestCoordinationHistoryStringsLimit(t *testing.T) {
	tm := &Team{coordination: []CoordinationEvent{
		{Timestamp: time.Unix(1, 0), From: "a", To: "b", Content: "first"},
		{Timestamp: time.Unix(2, 0), From: "c", To: "d", Content: "second"},
		{Timestamp: time.Unix(3, 0), From: "e", To: "f", Content: "third"},
	}}

	history := tm.CoordinationHistoryStrings(2)
	if len(history) != 2 {
		t.Fatalf("expected 2 history entries, got %d", len(history))
	}
	if !strings.Contains(history[0], "second") || !strings.Contains(history[1], "third") {
		t.Fatalf("expected to keep last two events, got %#v", history)
	}
}

func TestCheckWorkCompletedHeuristics(t *testing.T) {
	tm := &Team{}

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(originalWD) })

	if tm.checkWorkCompleted("agent", "Analyze reports") {
		t.Fatalf("expected false when task lacks file keywords")
	}

	filePath := filepath.Join(tempDir, "main.go")
	if err := os.WriteFile(filePath, []byte("package main"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if !tm.checkWorkCompleted("agent", "Please create a new script") {
		t.Fatalf("expected true once recent file work detected")
	}
}
