package team

import (
	"context"
	"time"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/tool"
)

// Add registers ag under name so it can be addressed via Call.
func (t *Team) Add(name string, ag *core.Agent) {
	if _, hasAgent := ag.Tools["agent"]; hasAgent {
		newTools := make(tool.Registry)
		for toolName, toolInstance := range ag.Tools {
			if toolName != "agent" {
				newTools[toolName] = toolInstance
			}
		}
		ag.Tools = newTools
		ag.InvalidateToolCache()
	}
	agent := &Agent{
		ID: name, Name: name, Agent: ag, Status: "ready", StartedAt: time.Now(), LastSeen: time.Now(), Metadata: make(map[string]string),
	}
	t.mutex.Lock()
	t.agents[name] = agent
	t.agentsByName[name] = agent
	t.mutex.Unlock()
}

// AddAgent creates a new agent and adds it to the team. Returns the core agent and its assigned name.
func (t *Team) AddAgent(name string) (*core.Agent, string) {
	spawned, err := t.SpawnAgent(context.Background(), name, name)
	if err != nil {
		debugPrintf("AddAgent fallback: failed to SpawnAgent(%s): %v", name, err)
		registry := tool.DefaultRegistry()
		delete(registry, "agent")
		coreAgent := core.New(t.parent.Client, t.parent.ModelName, registry, memory.NewInMemory(), memory.NewInMemoryVector(), t.parent.Tracer)
		t.Add(name, coreAgent)
		return coreAgent, name
	}
	return spawned.Agent, spawned.Name
}
