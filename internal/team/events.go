package team

import (
    "encoding/json"
    "fmt"
    "os"
    "time"
)

// LogCoordinationEvent adds a coordination event to the log and persists it.
func (t *Team) LogCoordinationEvent(eventType, from, to, content string, metadata map[string]interface{}) {
    t.mutex.Lock()
    defer t.mutex.Unlock()

    event := CoordinationEvent{
        ID:        fmt.Sprintf("%s_%d", eventType, time.Now().UnixNano()),
        Type:      eventType,
        From:      from,
        To:        to,
        Content:   content,
        Timestamp: time.Now(),
        Metadata:  metadata,
    }
    t.coordination = append(t.coordination, event)

    // Persist the event (best-effort)
    if t.store != nil {
        if b, err := json.Marshal(event); err == nil {
            _ = t.store.Set(t.name, "coord-"+event.ID, b, 0)
        }
    }

    // Enhanced console logging
    debugPrintf("📝 COORDINATION EVENT: %s -> %s | %s: %s\n", from, to, eventType, content)
    logToFile(fmt.Sprintf("COORDINATION: %s -> %s | %s: %s", from, to, eventType, content))
}

// loadCoordinationFromStore loads persisted coordination events at startup.
func (t *Team) loadCoordinationFromStore() {
    if t.store == nil {
        return
    }
    keys, err := t.store.Keys(t.name)
    if err != nil || len(keys) == 0 {
        return
    }
    // Collect coord-* keys
    events := make([]CoordinationEvent, 0)
    for _, k := range keys {
        if len(k) < 6 || k[:6] != "coord-" {
            continue
        }
        if b, ok, err := t.store.Get(t.name, k); err == nil && ok {
            var ev CoordinationEvent
            if err := json.Unmarshal(b, &ev); err == nil {
                events = append(events, ev)
            }
        }
    }
    if len(events) == 0 {
        return
    }
    // Append to in-memory log, keep order by timestamp
    t.mutex.Lock()
    t.coordination = append(t.coordination, events...)
    t.mutex.Unlock()
}

// GetCoordinationSummary returns a summary of recent coordination events
func (t *Team) GetCoordinationSummary() string {
    t.mutex.RLock()
    defer t.mutex.RUnlock()

    if len(t.coordination) == 0 {
        return "No coordination events recorded."
    }

    recent := t.coordination
    if len(recent) > 10 {
        recent = recent[len(recent)-10:] // Last 10 events
    }

    summary := fmt.Sprintf("Recent Coordination Events (%d total):\n", len(t.coordination))
    for _, event := range recent {
        summary += fmt.Sprintf("- %s: %s -> %s | %s\n",
            event.Timestamp.Format("15:04:05"), event.From, event.To, event.Content)
    }

    return summary
}

// CoordinationHistoryStrings returns formatted lines of coordination events.
// If limit <= 0, returns all events; otherwise returns the last 'limit' events.
func (t *Team) CoordinationHistoryStrings(limit int) []string {
    t.mutex.RLock()
    defer t.mutex.RUnlock()
    if len(t.coordination) == 0 {
        return nil
    }
    start := 0
    if limit > 0 && len(t.coordination) > limit {
        start = len(t.coordination) - limit
    }
    res := make([]string, 0, len(t.coordination)-start)
    for _, e := range t.coordination[start:] {
        res = append(res, fmt.Sprintf("%s %s -> %s | %s", e.Timestamp.Format("15:04:05"), e.From, e.To, e.Content))
    }
    return res
}

// WorkspaceEvent represents an event in the shared workspace
type WorkspaceEvent struct {
    ID          string                 `json:"id"`
    Type        string                 `json:"type"` // "file_created", "task_started", "task_completed", "question", "help_request"
    AgentID     string                 `json:"agent_id"`
    Description string                 `json:"description"`
    Timestamp   time.Time              `json:"timestamp"`
    Data        map[string]interface{} `json:"data"`
}

// PublishWorkspaceEvent publishes an event that all agents can see
func (t *Team) PublishWorkspaceEvent(agentID, eventType, description string, data map[string]interface{}) {
    event := WorkspaceEvent{
        ID:          fmt.Sprintf("%s_%s_%d", agentID, eventType, time.Now().Unix()),
        Type:        eventType,
        AgentID:     agentID,
        Description: description,
        Timestamp:   time.Now(),
        Data:        data,
    }

    // Store in shared memory
    eventsKey := "workspace_events"
    events, exists := t.GetSharedData(eventsKey)
    var eventList []WorkspaceEvent
    if exists {
        if oldList, ok := events.([]WorkspaceEvent); ok {
            eventList = oldList
        }
    }

    // Limit to last 50 events
    if len(eventList) > 50 {
        eventList = eventList[len(eventList)-50:]
    }
    eventList = append(eventList, event)
    t.SetSharedData(eventsKey, eventList)

    // Only log to stderr in non-TUI mode to avoid console interference
    if !isTUI() {
        fmt.Fprintf(os.Stderr, "📡 WORKSPACE EVENT: %s | %s: %s\n", agentID, eventType, description)
    }

    // Log coordination event
    t.LogCoordinationEvent("workspace_event", agentID, "*", fmt.Sprintf("%s: %s", eventType, description), map[string]interface{}{
        "event_type": eventType,
        "data":       data,
    })
}

// GetWorkspaceEvents returns recent workspace events
func (t *Team) GetWorkspaceEvents(limit int) []WorkspaceEvent {
    eventsKey := "workspace_events"
    events, exists := t.GetSharedData(eventsKey)
    if !exists {
        return []WorkspaceEvent{}
    }
    if eventList, ok := events.([]WorkspaceEvent); ok {
        if limit > 0 && len(eventList) > limit {
            return eventList[len(eventList)-limit:]
        }
        return eventList
    }
    return []WorkspaceEvent{}
}
