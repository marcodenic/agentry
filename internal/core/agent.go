package core

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/yourname/agentry/internal/memory"
	"github.com/yourname/agentry/internal/router"
	"github.com/yourname/agentry/internal/tool"
	"github.com/yourname/agentry/internal/trace"
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
	model := a.Route.Select(input)
	prompt := buildPrompt(a.Mem.History(), input, a.Tools)
	for i := 0; i < 8; i++ {
		out, err := model.Complete(ctx, prompt)
		if err != nil {
			return "", err
		}
		a.Trace(ctx, trace.EventStepStart, out)
		var call struct {
			Tool string         `json:"tool"`
			Args map[string]any `json:"args"`
		}
		if json.Unmarshal([]byte(out), &call) == nil && call.Tool != "" {
			t, ok := a.Tools.Use(call.Tool)
			if !ok {
				return "", errors.New("unknown tool")
			}
			r, err := t.Execute(ctx, call.Args)
			if err != nil {
				return "", err
			}
			a.Trace(ctx, trace.EventToolEnd, r)
			a.Mem.AddStep(out, call.Tool, r)
			prompt = buildPrompt(a.Mem.History(), input, a.Tools)
			continue
		}
		a.Mem.AddStep(out, "", "")
		a.Trace(ctx, trace.EventFinal, out)
		return out, nil
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

func buildPrompt(hist []memory.Step, input string, reg tool.Registry) string {
	var sb strings.Builder
	for _, h := range hist {
		sb.WriteString(h.Output)
		sb.WriteString("\n")
		if h.ToolName != "" {
			sb.WriteString(h.ToolResult)
			sb.WriteString("\n")
		}
	}
	sb.WriteString(input)
	return sb.String()
}
