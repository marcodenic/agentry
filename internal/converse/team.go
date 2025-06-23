package converse

import (
	"context"
	"errors"
	"fmt"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/router"
)

// Team manages a multi-agent conversation step by step.
type Team struct {
	agents   []*core.Agent
	names    []string
	msg      string
	turn     int
	maxTurns int
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
	for i := 0; i < n; i++ {
		ag := parent.Spawn()
		ag.Tracer = nil
		ag.Mem = shared
		ag.Route = convRoute
		agents[i] = ag
		names[i] = fmt.Sprintf("Agent%d", i+1)
	}
	return &Team{agents: agents, names: names, msg: topic, maxTurns: maxTurns}, nil
}

// Step advances the conversation by one turn and returns the agent index and output.
func (t *Team) Step(ctx context.Context) (int, string, error) {
	if t.turn >= t.maxTurns {
		return -1, "", errors.New("max turns reached")
	}
	idx := t.turn % len(t.agents)
	out, err := runAgent(ctx, t.agents[idx], t.msg, t.names[idx], t.names)
	if err != nil {
		return idx, "", err
	}
	t.msg = out
	t.turn++
	return idx, out, nil
}
