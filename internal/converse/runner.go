package converse

import (
	"context"
	"encoding/json"
	"errors"
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
	msgs := BuildMessages(ag.Mem.History(), input, name, peers)
	specs := tool.BuildSpecs(ag.Tools)
	for i := 0; i < 8; i++ {
		res, err := client.Complete(ctx, msgs, specs)
		if err != nil {
			return "", err
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
				return "", fmt.Errorf("unknown tool: %s", tc.Name)
			}
			var args map[string]any
			if err := json.Unmarshal(tc.Arguments, &args); err != nil {
				return "", err
			}
			r, err := t.Execute(ctx, args)
			if err != nil {
				return "", err
			}
			step.ToolResults[tc.ID] = r
			msgs = append(msgs, model.ChatMessage{Role: "tool", ToolCallID: tc.ID, Content: r})
		}
		ag.Mem.AddStep(step)
	}
	return "", errors.New("max iterations")
}
