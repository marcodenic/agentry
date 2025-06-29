package registry

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// InMemoryRegistry implements AgentRegistry using in-memory storage
type InMemoryRegistry struct {
	mu          sync.RWMutex
	agents      map[string]*AgentInfo
	health      map[string]*HealthMetrics
	subscribers []EventSubscriber
	
	// Configuration
	heartbeatTimeout time.Duration
	cleanupInterval  time.Duration
	
	// Internal state
	stopCh chan struct{}
	done   chan struct{}
}

// NewInMemoryRegistry creates a new in-memory agent registry
func NewInMemoryRegistry(heartbeatTimeout, cleanupInterval time.Duration) *InMemoryRegistry {
	if heartbeatTimeout == 0 {
		heartbeatTimeout = 30 * time.Second
	}
	if cleanupInterval == 0 {
		cleanupInterval = 60 * time.Second
	}
	
	registry := &InMemoryRegistry{
		agents:           make(map[string]*AgentInfo),
		health:           make(map[string]*HealthMetrics),
		subscribers:      make([]EventSubscriber, 0),
		heartbeatTimeout: heartbeatTimeout,
		cleanupInterval:  cleanupInterval,
		stopCh:           make(chan struct{}),
		done:             make(chan struct{}),
	}
	
	// Start cleanup goroutine
	go registry.cleanupLoop()
	
	return registry
}

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

// GetAgent retrieves information about a specific agent
func (r *InMemoryRegistry) GetAgent(ctx context.Context, agentID string) (*AgentInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	agent, exists := r.agents[agentID]
	if !exists {
		return nil, fmt.Errorf("agent %s not found", agentID)
	}
	
	// Return a copy to prevent external modification
	agentCopy := *agent
	return &agentCopy, nil
}

// ListAllAgents returns all registered agents
func (r *InMemoryRegistry) ListAllAgents(ctx context.Context) ([]*AgentInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	agents := make([]*AgentInfo, 0, len(r.agents))
	for _, agent := range r.agents {
		agentCopy := *agent
		agents = append(agents, &agentCopy)
	}
	
	return agents, nil
}

// FindAgents finds agents with specific capabilities
func (r *InMemoryRegistry) FindAgents(ctx context.Context, capabilities []string) ([]*AgentInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var matchingAgents []*AgentInfo
	
	for _, agent := range r.agents {
		if r.hasCapabilities(agent.Capabilities, capabilities) {
			agentCopy := *agent
			matchingAgents = append(matchingAgents, &agentCopy)
		}
	}
	
	return matchingAgents, nil
}

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

// Subscribe adds an event subscriber
func (r *InMemoryRegistry) Subscribe(subscriber EventSubscriber) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.subscribers = append(r.subscribers, subscriber)
}

// Close closes the registry and stops background processes
func (r *InMemoryRegistry) Close() error {
	close(r.stopCh)
	<-r.done
	return nil
}

// Helper methods

func (r *InMemoryRegistry) hasCapabilities(agentCaps, requiredCaps []string) bool {
	if len(requiredCaps) == 0 {
		return true // No specific capabilities required
	}
	
	capMap := make(map[string]bool)
	for _, cap := range agentCaps {
		capMap[cap] = true
	}
	
	for _, required := range requiredCaps {
		if !capMap[required] {
			return false
		}
	}
	
	return true
}

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
