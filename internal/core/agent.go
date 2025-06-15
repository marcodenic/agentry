package core

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/internal/tool"
	"github.com/marcodenic/agentry/internal/trace"
)

type Agent struct {
	ID     uuid.UUID
	Tools  tool.Registry
	Mem    memory.Store
	Route  router.Selector
	Tracer trace.Writer
}

const maxSteps = 32

func New(sel router.Selector, reg tool.Registry, mem memory.Store, tr trace.Writer) *Agent {
	return &Agent{uuid.New(), reg, mem, sel, tr}
}

func (a *Agent) Spawn() *Agent {
	return New(a.Route, a.Tools, memory.NewInMemory(), a.Tracer)
}

func (a *Agent) Run(ctx context.Context, input string) (string, error) {
	client, name := a.Route.Select(input)
	a.Trace(ctx, trace.EventModelStart, name)
	msgs := buildMessages(a.Mem.History(), input)
	specs := buildToolSpecs(a.Tools)
	for i := 0; i < maxSteps; i++ {
		res, err := client.Complete(ctx, msgs, specs)
		if err != nil {
			return "", err
		}
		a.Trace(ctx, trace.EventStepStart, res)
		msgs = append(msgs, model.ChatMessage{Role: "assistant", Content: res.Content, ToolCalls: res.ToolCalls})
		step := memory.Step{Output: res.Content, ToolCalls: res.ToolCalls, ToolResults: map[string]string{}}
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

func buildMessages(hist []memory.Step, input string) []model.ChatMessage {
	msgs := []model.ChatMessage{
		{Role: "system", Content: "You are an agent. When you call a tool, `arguments` must be a valid JSON object (use {} if no parameters). Control characters are forbidden."},
	}
	for _, h := range hist {
		msgs = append(msgs, model.ChatMessage{Role: "assistant", Content: h.Output, ToolCalls: h.ToolCalls})
		for id, res := range h.ToolResults {
			msgs = append(msgs, model.ChatMessage{Role: "tool", ToolCallID: id, Content: res})
		}
	}
	msgs = append(msgs, model.ChatMessage{Role: "user", Content: input})
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
