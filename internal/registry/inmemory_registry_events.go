package registry

import (
	"fmt"
	"time"
)

// Subscribe adds an event subscriber
func (r *InMemoryRegistry) Subscribe(subscriber EventSubscriber) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.subscribers = append(r.subscribers, subscriber)
}

// notifyEvent sends an event to all subscribers
func (r *InMemoryRegistry) notifyEvent(event *RegistryEvent) {
	for _, subscriber := range r.subscribers {
		go func(sub EventSubscriber) {
			if err := sub.OnEvent(event); err != nil {
				// Log error but don't fail the operation
				// In a real implementation, we'd use proper logging
				fmt.Printf("Error notifying subscriber: %v\n", err)
			}
		}(subscriber)
	}
}

// cleanupLoop runs periodic cleanup of unreachable agents
func (r *InMemoryRegistry) cleanupLoop() {
	defer close(r.done)
	
	ticker := time.NewTicker(r.cleanupInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-r.stopCh:
			return
		case <-ticker.C:
			r.cleanup()
		}
	}
}

// cleanup marks agents as unreachable if they haven't sent heartbeats
func (r *InMemoryRegistry) cleanup() {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	now := time.Now()
	var unreachableAgents []string
	
	for agentID, agent := range r.agents {
		if now.Sub(agent.LastSeen) > r.heartbeatTimeout {
			if agent.Status != StatusUnreachable {
				agent.Status = StatusUnreachable
				unreachableAgents = append(unreachableAgents, agentID)
			}
		}
	}
	
	// Notify about unreachable agents
	for _, agentID := range unreachableAgents {
		r.notifyEvent(&RegistryEvent{
			Type:      EventAgentUnreachable,
			AgentID:   agentID,
			Timestamp: now,
		})
	}
}
