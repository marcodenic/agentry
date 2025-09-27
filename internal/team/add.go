package team

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	teamruntime "github.com/marcodenic/agentry/internal/teamruntime"
	"github.com/marcodenic/agentry/internal/tool"
)

// Add registers ag under name so it can be addressed via Call.
// AddAgent creates a new agent and adds it to the team. Returns the core agent and its assigned name.
func (t *Team) AddAgent(name string) (*core.Agent, string) {
	spawned, err := t.SpawnAgent(context.Background(), name, name)
	if err != nil {
		teamruntime.Debugf("AddAgent fallback: failed to SpawnAgent(%s): %v", name, err)
		registry := tool.DefaultRegistry()
		delete(registry, "agent")
		coreAgent := core.New(t.parent.Client, t.parent.ModelName, registry, memory.NewInMemory(), memory.NewInMemoryVector(), t.parent.Tracer)
		ag := t.Add(name, coreAgent)
		return ag.Agent, ag.Name
	}
	return spawned.Agent, spawned.Name
}

// Add registers an existing core.Agent instance with the team under the supplied name.
// It enforces name uniqueness, strips the recursive "agent" tool, and records metadata
// so downstream coordination helpers can reason about the new participant.
func (t *Team) Add(name string, ag *core.Agent) *Agent {
	if ag == nil {
		return nil
	}

	clean := strings.TrimSpace(name)
	if clean == "" {
		clean = fmt.Sprintf("agent_%d", len(t.agents)+1)
	}

	// Ensure we do not leave the agent tool enabled when manually wiring agents into the team.
	delete(ag.Tools, "agent")
	ag.InvalidateToolCache()

	t.mutex.Lock()
	defer t.mutex.Unlock()

	assigned := clean
	counter := 1
	for {
		if _, exists := t.agentsByName[assigned]; !exists {
			break
		}
		counter++
		assigned = fmt.Sprintf("%s_%d", clean, counter)
	}

	role := clean
	meta := map[string]string{}
	if cfg, ok := t.roles[clean]; ok && cfg != nil {
		if cfg.Name != "" {
			role = cfg.Name
		}
		if len(cfg.Metadata) > 0 {
			meta = make(map[string]string, len(cfg.Metadata))
			for k, v := range cfg.Metadata {
				meta[k] = v
			}
		}
	}

	teamAgent := &Agent{
		ID:        uuid.New().String(),
		Name:      assigned,
		Role:      role,
		Agent:     ag,
		Status:    "ready",
		StartedAt: time.Now(),
		LastSeen:  time.Now(),
		Metadata:  meta,
	}

	t.agents[teamAgent.ID] = teamAgent
	t.agentsByName[assigned] = teamAgent

	teamruntime.Debugf("👥 Added agent %s (%s) to team", assigned, role)

	return teamAgent
}
