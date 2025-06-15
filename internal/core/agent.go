package core

import (
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

func New(sel router.Selector, reg tool.Registry, mem memory.Store, tr trace.Writer) *Agent {
	return &Agent{uuid.New(), reg, mem, sel, tr}
}

func (a *Agent) Spawn() *Agent {
	return New(a.Route, a.Tools, memory.NewInMemory(), a.Tracer)
}

func (a *Agent) Run(ctx context.Context, input string) (string, error) {
	client := a.Route.Select(input)
	msgs := buildMessages(a.Mem.History(), input)
	specs := buildToolSpecs(a.Tools)
	for i := 0; i < 8; i++ {
		res, err := client.Complete(ctx, msgs, specs)
		if err != nil {
			return "", err
		}
		a.Trace(ctx, trace.EventStepStart, res)
		msgs = append(msgs, model.ChatMessage{Role: "assistant", Content: res.Content, ToolCalls: res.ToolCalls})
		if len(res.ToolCalls) == 0 {
			a.Mem.AddStep(res.Content, "", "", "")
			a.Trace(ctx, trace.EventFinal, res.Content)
			return res.Content, nil
		}
		for _, tc := range res.ToolCalls {
			t, ok := a.Tools.Use(tc.Name)
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
			a.Trace(ctx, trace.EventToolEnd, r)
			a.Mem.AddStep(res.Content, tc.Name, r, tc.ID)
			msgs = append(msgs, model.ChatMessage{Role: "tool", ToolCallID: tc.ID, Content: r})
		}
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
		{Role: "system", Content: "You are an agent."},
	}
	for _, h := range hist {
		msgs = append(msgs, model.ChatMessage{Role: "assistant", Content: h.Output})
		if h.ToolName != "" {
			msgs = append(msgs, model.ChatMessage{Role: "tool", ToolCallID: h.CallID, Content: h.ToolResult})
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
