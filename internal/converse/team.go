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

type teamCtxKey struct{}

func contextWithTeam(ctx context.Context, t *Team) context.Context {
	return context.WithValue(ctx, teamCtxKey{}, t)
}

// TeamFromContext extracts a Team pointer if present.
func TeamFromContext(ctx context.Context) (*Team, bool) {
	t, ok := ctx.Value(teamCtxKey{}).(*Team)
	return t, ok
}

// Team manages a multi-agent conversation step by step.
type Team struct {
	agents       []*core.Agent
	names        []string
	agentsByName map[string]*core.Agent
	msg          string
	turn         int
	maxTurns     int
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
	return &Team{agents: agents, names: names, agentsByName: byName, msg: topic, maxTurns: maxTurns}, nil
}

// Step advances the conversation by one turn and returns the agent index and output.
func (t *Team) Step(ctx context.Context) (int, string, error) {
	if t.turn >= t.maxTurns {
		return -1, "", errors.New("max turns reached")
	}
	ctx = contextWithTeam(ctx, t)
	idx := t.turn % len(t.agents)
	out, err := runAgent(ctx, t.agents[idx], t.msg, t.names[idx], t.names)
	if err != nil {
		return idx, "", err
	}
	t.msg = out
	t.turn++
	return idx, out, nil
}

// ErrUnknownAgent is returned when Call is invoked with a name that doesn't exist.
var ErrUnknownAgent = errors.New("unknown agent")

// Call runs the named agent with the provided input once.
func (t *Team) Call(ctx context.Context, name, input string) (string, error) {
	ag, ok := t.agentsByName[name]
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrUnknownAgent, name)
	}
	ctx = contextWithTeam(ctx, t)
	return runAgent(ctx, ag, input, name, t.names)
}
