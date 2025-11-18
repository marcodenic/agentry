package team

import (
	"sync"
	"time"
)

// DelegationEventType represents a lifecycle event for delegated agents.
type DelegationEventType string

const (
	// DelegationEventSpawn indicates a new delegated agent instance was created.
	DelegationEventSpawn DelegationEventType = "spawn"
	// DelegationEventSessionStart indicates a delegation session has started work.
	DelegationEventSessionStart DelegationEventType = "session_start"
	// DelegationEventSessionComplete indicates a delegation session completed successfully.
	DelegationEventSessionComplete DelegationEventType = "session_complete"
	// DelegationEventSessionError indicates a delegation session ended with an error or timeout.
	DelegationEventSessionError DelegationEventType = "session_error"
)

// DelegationEvent captures lifecycle details about delegated agent activity for the TUI.
type DelegationEvent struct {
	Type        DelegationEventType
	TeamAgentID string
	CoreAgentID string
	AgentName   string
	Role        string
	SessionID   string
	Input       string
	Result      string
	Err         string
	Timestamp   time.Time
}

// SubscribeDelegationEvents allows observers (e.g., the TUI) to receive delegation lifecycle updates.
func (t *Team) SubscribeDelegationEvents() (<-chan DelegationEvent, func()) {
	ch := make(chan DelegationEvent, 32)

	t.mutex.Lock()
	id := t.nextDelegationSubID
	if t.delegationSubscribers == nil {
		t.delegationSubscribers = make(map[int]chan DelegationEvent)
	}
	t.delegationSubscribers[id] = ch
	t.nextDelegationSubID++
	t.mutex.Unlock()

	var once sync.Once

	return ch, func() {
		once.Do(func() {
			t.mutex.Lock()
			if subsCh, ok := t.delegationSubscribers[id]; ok {
				delete(t.delegationSubscribers, id)
				close(subsCh)
			}
			t.mutex.Unlock()
		})
	}
}

func (t *Team) emitDelegationEvent(ev DelegationEvent) {
	if ev.Timestamp.IsZero() {
		ev.Timestamp = time.Now()
	}

	t.mutex.RLock()
	subs := make([]chan DelegationEvent, 0, len(t.delegationSubscribers))
	for _, ch := range t.delegationSubscribers {
		subs = append(subs, ch)
	}
	t.mutex.RUnlock()

	for _, ch := range subs {
		safeSendDelegationEvent(ch, ev)
	}
}

func safeSendDelegationEvent(ch chan DelegationEvent, ev DelegationEvent) {
	defer func() {
		if recover() != nil {
			// Channel was closed between capture and send; ignore.
		}
	}()

	select {
	case ch <- ev:
	default:
		go func(out chan DelegationEvent, event DelegationEvent) {
			defer func() {
				if recover() != nil {
					// Channel closed; nothing to do.
				}
			}()
			select {
			case out <- event:
			case <-time.After(100 * time.Millisecond):
				// Drop event to avoid blocking; subsequent tick will refresh state.
			}
		}(ch, ev)
	}
}
