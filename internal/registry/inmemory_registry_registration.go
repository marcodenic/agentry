package registry

import (
	"context"
	"fmt"
	"time"
)

// RegisterAgent registers a new agent with the registry
func (r *InMemoryRegistry) RegisterAgent(ctx context.Context, info *AgentInfo) error {
	if info.ID == "" {
		return fmt.Errorf("agent ID cannot be empty")
	}
	
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Set registration time and last seen
	now := time.Now()
	info.RegisteredAt = now
	info.LastSeen = now
	
	// Set default status if not provided
	if info.Status == "" {
		info.Status = StatusIdle
	}
	
	// Store agent info
	r.agents[info.ID] = info
	
	// Initialize health metrics
	r.health[info.ID] = &HealthMetrics{
		CPUUsage:       0,
		MemoryUsage:    0,
		TasksCompleted: 0,
		TasksActive:    0,
		ErrorCount:     0,
		Uptime:         0,
	}
	
	// Notify subscribers
	r.notifyEvent(&RegistryEvent{
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
func (r *InMemoryRegistry) DeregisterAgent(ctx context.Context, agentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.agents[agentID]; !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}
	
	delete(r.agents, agentID)
	delete(r.health, agentID)
	
	// Notify subscribers
	r.notifyEvent(&RegistryEvent{
		Type:      EventAgentDeregistered,
		AgentID:   agentID,
		Timestamp: time.Now(),
	})
	
	return nil
}

// UpdateAgent updates agent information
func (r *InMemoryRegistry) UpdateAgent(ctx context.Context, agentID string, info *AgentInfo) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	existingAgent, exists := r.agents[agentID]
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}
	
	// Preserve registration time
	info.RegisteredAt = existingAgent.RegisteredAt
	info.LastSeen = time.Now()
	info.ID = agentID // Ensure ID consistency
	
	r.agents[agentID] = info
	
	return nil
}
