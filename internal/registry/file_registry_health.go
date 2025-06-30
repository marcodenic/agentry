package registry

import (
	"context"
	"fmt"
	"time"
)

// UpdateAgentStatus updates the status of an agent
func (r *FileRegistry) UpdateAgentStatus(ctx context.Context, agentID string, status AgentStatus) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	agent, exists := r.agents[agentID]
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	oldStatus := agent.Status
	agent.Status = status
	agent.LastSeen = time.Now()

	if err := r.saveToFile(); err != nil {
		return err
	}

	// Emit status change event
	if oldStatus != status {
		r.emitEvent(&RegistryEvent{
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
func (r *FileRegistry) UpdateAgentHealth(ctx context.Context, agentID string, health *HealthMetrics) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.agents[agentID]; !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	r.health[agentID] = health
	return nil
}

// GetAgentHealth retrieves health metrics for an agent
func (r *FileRegistry) GetAgentHealth(ctx context.Context, agentID string) (*HealthMetrics, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	health, exists := r.health[agentID]
	if !exists {
		return nil, fmt.Errorf("health metrics for agent %s not found", agentID)
	}

	// Return a copy
	healthCopy := *health
	return &healthCopy, nil
}

// Heartbeat updates the last seen time for an agent
func (r *FileRegistry) Heartbeat(ctx context.Context, agentID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	agent, exists := r.agents[agentID]
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	agent.LastSeen = time.Now()
	
	// Update status to running if it was starting
	if agent.Status == StatusStarting {
		agent.Status = StatusIdle
	}

	return r.saveToFile()
}
