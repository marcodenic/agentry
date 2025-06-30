package registry

import (
	"context"
	"fmt"
	"time"
)

// RegisterAgent registers a new agent with the registry
func (r *FileRegistry) RegisterAgent(ctx context.Context, info *AgentInfo) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if info.ID == "" {
		return fmt.Errorf("agent ID cannot be empty")
	}

	// Set registration time and initial status
	now := time.Now()
	info.RegisteredAt = now
	info.LastSeen = now
	if info.Status == "" {
		info.Status = StatusStarting
	}

	r.agents[info.ID] = info

	// Initialize health metrics
	r.health[info.ID] = &HealthMetrics{
		TasksCompleted: 0,
		TasksActive:    0,
		ErrorCount:     0,
		Uptime:         0,
	}

	if err := r.saveToFile(); err != nil {
		return fmt.Errorf("failed to persist agent registration: %w", err)
	}

	// Emit registration event
	r.emitEvent(&RegistryEvent{
		Type:      EventAgentRegistered,
		AgentID:   info.ID,
		Timestamp: now,
		Data: map[string]string{
			"endpoint": info.Endpoint,
			"role":     info.Role,
		},
	})

	return nil
}

// DeregisterAgent removes an agent from the registry
func (r *FileRegistry) DeregisterAgent(ctx context.Context, agentID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.agents[agentID]; !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	delete(r.agents, agentID)
	delete(r.health, agentID)

	if err := r.saveToFile(); err != nil {
		return fmt.Errorf("failed to persist agent deregistration: %w", err)
	}

	// Emit deregistration event
	r.emitEvent(&RegistryEvent{
		Type:      EventAgentDeregistered,
		AgentID:   agentID,
		Timestamp: time.Now(),
	})

	return nil
}

// UpdateAgent updates agent information
func (r *FileRegistry) UpdateAgent(ctx context.Context, agentID string, info *AgentInfo) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.agents[agentID]; !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	// Preserve registration time
	info.RegisteredAt = r.agents[agentID].RegisteredAt
	info.LastSeen = time.Now()
	r.agents[agentID] = info

	return r.saveToFile()
}
