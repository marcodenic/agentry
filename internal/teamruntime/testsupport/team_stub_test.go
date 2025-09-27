package testsupport

import "testing"

func TestStubTeamRecordsCoordinationEvents(t *testing.T) {
	tm := &StubTeam{}
	tm.LogCoordinationEvent("delegation", "agent_a", "agent_b", "hand-off", nil)
	events := tm.EventsOfType("delegation")
	if len(events) != 1 {
		t.Fatalf("expected 1 delegation event, got %d", len(events))
	}
	if events[0].Content != "hand-off" {
		t.Fatalf("unexpected event content: %s", events[0].Content)
	}
}

func TestPublishWorkspaceEventAddsLogAndEvent(t *testing.T) {
	tm := &StubTeam{}
	tm.PublishWorkspaceEvent("builder", "task_completed", "shipped")

	if len(tm.WorkspaceLogs) != 1 {
		t.Fatalf("expected workspace log entry, got %d", len(tm.WorkspaceLogs))
	}
	if tm.WorkspaceLogs[0].Desc != "shipped" {
		t.Fatalf("unexpected workspace description: %s", tm.WorkspaceLogs[0].Desc)
	}
	if len(tm.EventsOfType("workspace_event")) != 1 {
		t.Fatalf("expected workspace_event coordination entry")
	}
}

func TestFakeWorkspaceEventsProvidesSampleData(t *testing.T) {
	events := FakeWorkspaceEvents("dev", "feature complete")
	if len(events) != 2 {
		t.Fatalf("expected two sample events, got %d", len(events))
	}
	if events[0].AgentID != "dev" {
		t.Fatalf("unexpected agent in sample events: %s", events[0].AgentID)
	}
}
