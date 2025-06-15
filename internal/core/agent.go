package core

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/internal/tool"
	"github.com/marcodenic/agentry/internal/trace"
)

type Agent struct {
	ID     uuid.UUID
	Name   string
	Tools  tool.Registry
	Mem    memory.Store
	Route  router.Selector
	Tracer trace.Writer
}

const maxSteps = 32

func cleanInput(s string) string {
	return strings.Map(func(r rune) rune {
		if r < 0x20 && r != '\n' && r != '\t' && r != '\r' {
			return -1
		}
		return r
	}, s)
}

func New(sel router.Selector, reg tool.Registry, mem memory.Store, tr trace.Writer) *Agent {
	return &Agent{ID: uuid.New(), Tools: reg, Mem: mem, Route: sel, Tracer: tr}
}

func NewNamed(name string, sel router.Selector, reg tool.Registry, mem memory.Store, tr trace.Writer) *Agent {
	a := New(sel, reg, mem, tr)
	a.Name = name
	return a
}

func (a *Agent) Spawn() *Agent {
	child := New(a.Route, a.Tools, memory.NewInMemory(), a.Tracer)
	child.Name = a.Name
	return child
}

func (a *Agent) Run(ctx context.Context, input string) (string, error) {
	if strings.TrimSpace(input) != "" {
		a.Mem.AddStep(memory.Step{Speaker: "user", Output: input})
	}

	client, name := a.Route.Select(input)
	a.Trace(ctx, trace.EventModelStart, name)
	input = ""
	speaker := a.ID.String()
	if a.Name != "" {
		speaker = a.Name
	}
	msgs := buildMessages(a.Mem.History(), input, speaker)
	specs := buildToolSpecs(a.Tools)
	for i := 0; i < maxSteps; i++ {
		res, err := client.Complete(ctx, msgs, specs)
		if err != nil {
			return "", err
		}
		a.Trace(ctx, trace.EventStepStart, res)
		name := a.ID.String()
		if a.Name != "" {
			name = a.Name
		}
		msgs = append(msgs, model.ChatMessage{Role: "assistant", Name: name, Content: res.Content, ToolCalls: res.ToolCalls})
		step := memory.Step{Speaker: name, Output: res.Content, ToolCalls: res.ToolCalls, ToolResults: map[string]string{}}
		if len(res.ToolCalls) == 0 {
			a.Mem.AddStep(step)
			a.Trace(ctx, trace.EventFinal, res.Content)
			return res.Content, nil
		}
		seen := map[string]bool{}
		for _, tc := range res.ToolCalls {
			key := tc.Name + string(tc.Arguments)
			if seen[key] {
				return "", fmt.Errorf("model is looping on tool %s", tc.Name)
			}
			seen[key] = true
			t, ok := a.Tools.Use(tc.Name)
			if !ok {
				return "", fmt.Errorf("unknown tool: %s", tc.Name)
			}
			var args map[string]any
			clean := bytes.Map(func(r rune) rune {
				if r < 0x20 {
					return -1
				}
				return r
			}, tc.Arguments)
			if len(bytes.TrimSpace(clean)) == 0 {
				args = map[string]any{}
			} else if err := json.Unmarshal(clean, &args); err != nil {
				args = map[string]any{"_raw": string(clean)}
			}
			r, err := t.Execute(ctx, args)
			if err != nil {
				return "", err
			}
			a.Trace(ctx, trace.EventToolEnd, map[string]any{"name": tc.Name, "result": r})
			step.ToolResults[tc.ID] = r
			msgs = append(msgs, model.ChatMessage{Role: "tool", ToolCallID: tc.ID, Content: r})
		}
		a.Mem.AddStep(step)
	}
	return "", errors.New("max iterations")
}

func (a *Agent) Trace(ctx context.Context, typ trace.EventType, data any) {
	if a.Tracer != nil {
		a.Tracer.Write(ctx, trace.Event{
			Type:      typ,
			AgentID:   a.ID.String(),
			Data:      data,
			Timestamp: trace.Now(),
		})
	}
}

func buildMessages(hist []memory.Step, input, speaker string) []model.ChatMessage {
	input = cleanInput(input)
	sys := fmt.Sprintf("You are %s, one of several agents in a discussion. Read the conversation so far and add something new. When you call a tool, `arguments` must be a valid JSON object (use {} if no parameters). Control characters are forbidden.", speaker)
	msgs := []model.ChatMessage{{Role: "system", Content: sys}}
	for _, h := range hist {
		// Treat messages from other agents as user turns so the
		// current model perceives them as prompts to continue.
		role := "assistant"
		if h.Speaker != speaker {
			role = "user"
		}
		msgs = append(msgs, model.ChatMessage{
			Role:      role,
			Name:      h.Speaker,
			Content:   h.Output,
			ToolCalls: h.ToolCalls,
		})
		for id, res := range h.ToolResults {
			msgs = append(msgs, model.ChatMessage{Role: "tool", ToolCallID: id, Content: res})
		}
	}
	if strings.TrimSpace(input) != "" {
		msgs = append(msgs, model.ChatMessage{Role: "user", Content: input})
	}
	return msgs
}

func buildToolSpecs(reg tool.Registry) []model.ToolSpec {
	specs := make([]model.ToolSpec, 0, len(reg))
	for _, t := range reg {
		specs = append(specs, model.ToolSpec{
			Name:        t.Name(),
			Description: t.Description(),
			Parameters:  t.JSONSchema(),
		})
	}
	return specs
}
