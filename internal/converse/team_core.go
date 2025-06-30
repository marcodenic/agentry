package converse

import (
	"fmt"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/router"
)

// Team manages a multi-agent conversation step by step.
type Team struct {
	parent       *core.Agent
	agents       []*core.Agent
	names        []string
	agentsByName map[string]*core.Agent
	msg          string
	turn         int
	maxTurns     int
}

// Add registers ag under name so it can be addressed via Call.
func (t *Team) Add(name string, ag *core.Agent) {
	t.agents = append(t.agents, ag)
	t.names = append(t.names, name)
	if t.agentsByName == nil {
		t.agentsByName = map[string]*core.Agent{}
	}
	t.agentsByName[name] = ag
}

// Agents returns the current set of agents in the team.
func (t *Team) Agents() []*core.Agent { return t.agents }

// Names returns the display names of the agents.
func (t *Team) Names() []string { return t.names }

// NewTeamContext creates a Team context ready for agents to be added dynamically.
// Unlike NewTeam, this does not pre-spawn any agents - they can be added later
// via AddAgent or other team spawning mechanisms.
func NewTeamContext(parent *core.Agent) (*Team, error) {
	return &Team{
		parent:       parent,
		agents:       make([]*core.Agent, 0),       // Start with empty agent list
		names:        make([]string, 0),            // Start with empty names list
		agentsByName: make(map[string]*core.Agent), // Start with empty map
		msg:          "Hello agents, let's chat!",  // Default initial message
		maxTurns:     maxTurns,
	}, nil
}

// NewTeam spawns n sub-agents from parent ready to converse.
func NewTeam(parent *core.Agent, n int, topic string) (*Team, error) {
	if n <= 0 {
		return nil, fmt.Errorf("n must be > 0")
	}
	if topic == "" {
		topic = "Hello agents, let's chat!"
	}

	shared := memory.NewInMemory()

	convRoute := parent.Route
	if rules, ok := parent.Route.(router.Rules); ok {
		cpy := make(router.Rules, len(rules))
		for i, r := range rules {
			cpy[i] = r
			cpy[i].Client = model.WithTemperature(r.Client, 0.7)
		}
		convRoute = cpy
	}

	agents := make([]*core.Agent, n)
	names := make([]string, n)
	byName := make(map[string]*core.Agent, n)
	for i := 0; i < n; i++ {
		ag := parent.Spawn()
		ag.Tracer = nil
		ag.Mem = shared
		ag.Route = convRoute
		agents[i] = ag
		name := fmt.Sprintf("Agent%d", i+1)
		names[i] = name
		byName[name] = ag
	}
	return &Team{parent: parent, agents: agents, names: names, agentsByName: byName, msg: topic, maxTurns: maxTurns}, nil
}
