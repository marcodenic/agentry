package registry

import (
	"context"
	"fmt"
	"time"
)

// UpdateAgentStatus updates the status of an agent
func (r *InMemoryRegistry) UpdateAgentStatus(ctx context.Context, agentID string, status AgentStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	agent, exists := r.agents[agentID]
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}
	
	oldStatus := agent.Status
	agent.Status = status
	agent.LastSeen = time.Now()
	
	// Notify subscribers if status changed
	if oldStatus != status {
		r.notifyEvent(&RegistryEvent{
			Type:      EventAgentStatusChange,
			AgentID:   agentID,
			Timestamp: time.Now(),
			Data: map[string]string{
				"old_status": string(oldStatus),
				"new_status": string(status),
			},
		})
	}
	
	return nil
}

// UpdateAgentHealth updates health metrics for an agent
func (r *InMemoryRegistry) UpdateAgentHealth(ctx context.Context, agentID string, health *HealthMetrics) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.agents[agentID]; !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}
	
	r.health[agentID] = health
	return nil
}

// GetAgentHealth retrieves health metrics for an agent
func (r *InMemoryRegistry) GetAgentHealth(ctx context.Context, agentID string) (*HealthMetrics, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	health, exists := r.health[agentID]
	if !exists {
		return nil, fmt.Errorf("health metrics for agent %s not found", agentID)
	}
	
	// Return a copy
	healthCopy := *health
	return &healthCopy, nil
}

// Heartbeat updates the last seen time for an agent
func (r *InMemoryRegistry) Heartbeat(ctx context.Context, agentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	agent, exists := r.agents[agentID]
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}
	
	agent.LastSeen = time.Now()
	
	// If agent was unreachable, mark as idle
	if agent.Status == StatusUnreachable {
		agent.Status = StatusIdle
		r.notifyEvent(&RegistryEvent{
			Type:      EventAgentStatusChange,
			AgentID:   agentID,
			Timestamp: time.Now(),
			Data: map[string]string{
				"old_status": string(StatusUnreachable),
				"new_status": string(StatusIdle),
			},
		})
	}
	
	return nil
}
