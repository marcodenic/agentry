package team

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/marcodenic/agentry/internal/contracts"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memstore"
)

// Compile-time check to ensure Team implements contracts.TeamService
var _ contracts.TeamService = (*Team)(nil)

// Team manages a multi-agent conversation step by step.
// This is a simplified version that consolidates the functionality
// from converse.Team and maintains compatibility.
type Team struct {
    parent       *core.Agent
    agents       map[string]*Agent // Changed to use Agent type
    agentsByName map[string]*Agent // Changed to use Agent type
    tasks        map[string]*Task
    roles        map[string]*RoleConfig
    name         string
    maxTurns     int
    mutex        sync.RWMutex
	// ENHANCED: Shared memory and communication tracking
	sharedMemory map[string]interface{} // Shared data between agents
	store        memstore.SharedStore   // Durable-backed store (in-memory by default)
	coordination []CoordinationEvent    // Log of coordination events
}

// NewTeam creates a new team with the given parent agent.
func NewTeam(parent *core.Agent, maxTurns int, name string) (*Team, error) {
    team := &Team{
        parent:       parent,
        maxTurns:     maxTurns,
        name:         name,
        agents:       make(map[string]*Agent),
        agentsByName: make(map[string]*Agent),
        tasks:        make(map[string]*Task),
        roles:        make(map[string]*RoleConfig),
        sharedMemory: make(map[string]interface{}),
        store:        memstore.Get(),
        coordination: make([]CoordinationEvent, 0),
    }

	// Kick off default GC for the store (once-per-process)
	memstore.StartDefaultGC(60 * time.Second)

	// Best-effort: load persisted coordination events for this team
	team.loadCoordinationFromStore()

	return team, nil
}

// NewTeamWithRoles creates a new team with the given parent agent and loads role configurations.
func NewTeamWithRoles(parent *core.Agent, maxTurns int, name string, includePaths []string, configDir string) (*Team, error) {
	team, err := NewTeam(parent, maxTurns, name)
	if err != nil {
		return nil, err
	}

	// Load role configurations from include paths
	if len(includePaths) > 0 {
		roles, err := LoadRolesFromIncludePaths(includePaths, configDir)
		if err != nil {
			// Don't fail completely, just warn and continue with empty roles
			fmt.Printf("Warning: failed to load some roles: %v\n", err)
		}

		// Add loaded roles to team
		for name, role := range roles {
			team.roles[name] = role
			if !isTUI() {
				if isDebug() {
					fmt.Fprintf(os.Stderr, "📋 Team role loaded: %s\n", name)
				}
			}
		}
	}

	return team, nil
}

// GetRoles returns the loaded role configurations by name.
func (t *Team) GetRoles() map[string]*RoleConfig {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	out := make(map[string]*RoleConfig, len(t.roles))
	for k, v := range t.roles {
		out[k] = v
	}
	return out
}

// Tool curation helpers moved to registry.go

// Add and AddAgent moved to add.go

// Call implements the Caller interface for compatibility with existing code.
// It delegates work to the named agent with enhanced communication logging.
// Call moved to delegation.go

// CallParallel executes multiple agent tasks in parallel for improved efficiency
// CallParallel moved to delegation.go

// runAgent and minimal context builder moved to runtime.go

// Service wrappers moved to service.go

// Logging helpers moved to logging.go

// Shared memory helpers moved to shared.go

// Coordination event helpers moved to events.go

// Direct agent-to-agent communication methods moved to messaging.go

// Workspace events moved to events.go

// Help and completion heuristics moved to help.go
