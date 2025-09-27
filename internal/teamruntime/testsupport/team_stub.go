package testsupport

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/marcodenic/agentry/internal/teamruntime"
)

type CoordinationEvent struct {
	ID        string
	Type      string
	From      string
	To        string
	Content   string
	Timestamp time.Time
	Metadata  map[string]interface{}
}

type StubTeam struct {
	mu            sync.RWMutex
	Events        []CoordinationEvent
	WorkspaceLogs []struct {
		AgentID string
		Type    string
		Desc    string
	}
}

func (t *StubTeam) LogCoordinationEvent(eventType, from, to, content string, metadata map[string]interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Events = append(t.Events, CoordinationEvent{
		ID:        uuid.NewString(),
		Type:      eventType,
		From:      from,
		To:        to,
		Content:   content,
		Timestamp: time.Now(),
		Metadata:  metadata,
	})
}

func (t *StubTeam) PublishWorkspaceEvent(agentID, eventType, description string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.WorkspaceLogs = append(t.WorkspaceLogs, struct {
		AgentID string
		Type    string
		Desc    string
	}{AgentID: agentID, Type: eventType, Desc: description})
	t.Events = append(t.Events, CoordinationEvent{
		ID:        uuid.NewString(),
		Type:      "workspace_event",
		From:      agentID,
		To:        "*",
		Content:   fmt.Sprintf("%s: %s", eventType, description),
		Timestamp: time.Now(),
	})
}

func (t *StubTeam) EventsOfType(eventType string) []CoordinationEvent {
	t.mu.RLock()
	defer t.mu.RUnlock()
	out := make([]CoordinationEvent, 0)
	for _, ev := range t.Events {
		if ev.Type == eventType {
			out = append(out, ev)
		}
	}
	return out
}

func FakeWorkspaceEvents(agent, desc string) []teamruntime.WorkspaceEvent {
	return []teamruntime.WorkspaceEvent{
		{AgentID: agent, Type: "task_completed", Description: desc, Timestamp: time.Unix(1, 0)},
		{AgentID: "review", Type: "feedback", Description: "LGTM", Timestamp: time.Unix(2, 0)},
	}
}
