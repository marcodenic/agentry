package registry

import (
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

// Close closes the registry and stops background processes
func (r *InMemoryRegistry) Close() error {
	close(r.stopCh)
	<-r.done
	return nil
}

// hasCapabilities checks if an agent has all required capabilities
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
