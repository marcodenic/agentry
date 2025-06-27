package converse

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/internal/tool"
)

const maxTurns = 10

// Run spawns n sub-agents from parent and lets them talk about the given topic.
// The returned slice contains each message in order.
func Run(ctx context.Context, parent *core.Agent, n int, topic string) ([]string, error) {
	return runLoop(ctx, parent, n, topic, nil)
}

func runLoop(ctx context.Context, parent *core.Agent, n int, topic string, cb func(turn int, out string)) ([]string, error) {
	if n <= 0 {
		return nil, fmt.Errorf("n must be > 0")
	}

	if topic == "" {
		topic = "Hello agents, let's chat!"
	} else if (strings.HasPrefix(topic, "\"") && strings.HasSuffix(topic, "\"")) ||
		(strings.HasPrefix(topic, "'") && strings.HasSuffix(topic, "'")) {
		topic = strings.Trim(topic, "'\"")
	}

	shared := memory.NewInMemory()

	// Copy router rules to bump temperature
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

	transcript := make([]string, 0, maxTurns)
	msg := topic
	for turn := 0; turn < maxTurns; turn++ {
		idx := turn % n
		out, err := runAgent(ctx, agents[idx], msg, names[idx], names)
		if err != nil {
			return transcript, err
		}
		if cb != nil {
			cb(turn, out)
		}
		transcript = append(transcript, out)
		msg = out
	}
	return transcript, nil
}

// Repl runs Run() and prints coloured output.
func Repl(parent *core.Agent, n int, topic string) {
	_, err := runLoop(context.Background(), parent, n, topic, func(turn int, out string) {
		idx := turn % n
		col := colourFor(idx)
		fmt.Printf("%s[Agent%d]%s: %s\n", col, idx+1, colourReset, out)
	})
	if err != nil {
		fmt.Println("ERR:", err)
	}
}

func runAgent(ctx context.Context, ag *core.Agent, input, name string, peers []string) (string, error) {
	client, _ := ag.Route.Select(input)
	msgs := core.BuildMessages(ag.Prompt, ag.Vars, ag.Mem.History(), input)
	specs := tool.BuildSpecs(ag.Tools)
	limit := ag.MaxIterations
	if limit <= 0 {
		limit = 8 // Default for agents without explicit limit
	}
	// Special case: if MaxIterations is set to -1, allow unlimited iterations
	unlimited := ag.MaxIterations == -1
	
	for i := 0; unlimited || i < limit; i++ {
		res, err := client.Complete(ctx, msgs, specs)
		if err != nil {
			return "", fmt.Errorf("agent '%s' completion failed on iteration %d: %w", name, i+1, err)
		}
		msgs = append(msgs, model.ChatMessage{Role: "assistant", Content: res.Content, ToolCalls: res.ToolCalls})
		step := memory.Step{Output: res.Content, ToolCalls: res.ToolCalls, ToolResults: map[string]string{}}
		if len(res.ToolCalls) == 0 {
			ag.Mem.AddStep(step)
			return res.Content, nil
		}
		for _, tc := range res.ToolCalls {
			t, ok := ag.Tools.Use(tc.Name)
			if !ok {
				return "", fmt.Errorf("agent '%s' tried to use unknown tool '%s' on iteration %d", name, tc.Name, i+1)
			}
			var args map[string]any
			if err := json.Unmarshal(tc.Arguments, &args); err != nil {
				return "", fmt.Errorf("agent '%s' tool '%s' has invalid arguments on iteration %d: %w", name, tc.Name, i+1, err)
			}
			r, err := t.Execute(ctx, args)
			if err != nil {
				return "", fmt.Errorf("agent '%s' tool '%s' execution failed on iteration %d with args %v: %w", name, tc.Name, i+1, args, err)
			}
			step.ToolResults[tc.ID] = r
			msgs = append(msgs, model.ChatMessage{Role: "tool", ToolCallID: tc.ID, Content: r})
		}
		ag.Mem.AddStep(step)
	}
	return "", fmt.Errorf("agent '%s' exceeded maximum iterations (%d)", name, limit)
}
