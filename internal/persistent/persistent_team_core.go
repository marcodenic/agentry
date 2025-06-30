package persistent

import (
	"fmt"
	"sync"

	"github.com/marcodenic/agentry/internal/config"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/registry"
	"github.com/marcodenic/agentry/internal/sessions"
)

// PersistentTeam manages a collection of persistent agents
type PersistentTeam struct {
	parent         *core.Agent
	agents         map[string]*PersistentAgent
	registry       registry.AgentRegistry
	sessionManager sessions.SessionManager
	portRange      registry.PortRange
	mutex          sync.RWMutex
}

// DefaultPortRange returns the default port range for agents
func DefaultPortRange() registry.PortRange {
	return registry.PortRange{Start: 9000, End: 9099}
}

// NewPersistentTeam creates a new persistent team with the given parent agent and port range
func NewPersistentTeam(parent *core.Agent, portRange registry.PortRange) (*PersistentTeam, error) {
	// Create file-based registry for agent discovery
	reg, err := registry.NewFileRegistry(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create registry: %w", err)
	}

	// Create session manager
	sessionManager, err := sessions.NewFileSessionManager("./sessions")
	if err != nil {
		return nil, fmt.Errorf("failed to create session manager: %w", err)
	}

	return &PersistentTeam{
		parent:         parent,
		agents:         make(map[string]*PersistentAgent),
		registry:       reg,
		sessionManager: sessionManager,
		portRange:      portRange,
	}, nil
}

// NewPersistentTeamFromConfig creates a PersistentTeam from configuration
func NewPersistentTeamFromConfig(parent *core.Agent, cfg *config.PersistentAgentsConfig) (*PersistentTeam, error) {
	if cfg == nil || !cfg.Enabled {
		return nil, fmt.Errorf("persistent agents not enabled")
	}

	portStart := cfg.PortStart
	portEnd := cfg.PortEnd
	
	// Set default port range if not specified
	if portStart == 0 {
		portStart = 9001
	}
	if portEnd == 0 {
		portEnd = 9010
	}

	return NewPersistentTeam(parent, registry.PortRange{
		Start: portStart,
		End:   portEnd,
	})
}

// Close shuts down all persistent agents and cleans up resources
func (pt *PersistentTeam) Close() error {
	pt.mutex.Lock()
	defer pt.mutex.Unlock()

	for _, agent := range pt.agents {
		pt.stopAgent(agent)
	}
	pt.agents = make(map[string]*PersistentAgent)

	return pt.registry.Close()
}
