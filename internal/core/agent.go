package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/internal/tool"
	"github.com/marcodenic/agentry/internal/trace"
	"github.com/marcodenic/agentry/pkg/memstore"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Agent struct {
	ID     uuid.UUID
	Tools  tool.Registry
	Mem    memory.Store
	Route  router.Selector
	Tracer trace.Writer
	Store  memstore.KV
}

var (
	tokenCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "agentry_tokens_total",
		Help: "Total tokens processed by an agent",
	}, []string{"agent"})
	toolLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "agentry_tool_latency_seconds",
		Help:    "Latency of tool execution in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"agent", "tool"})
)

func New(sel router.Selector, reg tool.Registry, mem memory.Store, store memstore.KV, tr trace.Writer) *Agent {
	return &Agent{uuid.New(), reg, mem, sel, tr, store}
}

func (a *Agent) Spawn() *Agent {
	return New(a.Route, a.Tools, memory.NewInMemory(), a.Store, a.Tracer)
}

func (a *Agent) Run(ctx context.Context, input string) (string, error) {
	client, name := a.Route.Select(input)
	a.Trace(ctx, trace.EventModelStart, name)
	msgs := buildMessages(a.Mem.History(), input)
	specs := tool.BuildSpecs(a.Tools)
	tokenCounter.WithLabelValues(a.ID.String()).Add(float64(len(strings.Fields(input))))
	for i := 0; i < 8; i++ {
		res, err := client.Complete(ctx, msgs, specs)
		if err != nil {
			return "", err
		}
		a.Trace(ctx, trace.EventStepStart, res)
		tokenCounter.WithLabelValues(a.ID.String()).Add(float64(len(strings.Fields(res.Content))))
		msgs = append(msgs, model.ChatMessage{Role: "assistant", Content: res.Content, ToolCalls: res.ToolCalls})
		step := memory.Step{Output: res.Content, ToolCalls: res.ToolCalls, ToolResults: map[string]string{}}
		if len(res.ToolCalls) == 0 {
			a.Mem.AddStep(step)
			_ = a.Checkpoint(ctx)
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
			start := time.Now()
			r, err := t.Execute(ctx, args)
			toolLatency.WithLabelValues(a.ID.String(), tc.Name).Observe(time.Since(start).Seconds())
			if err != nil {
				return "", err
			}
			tokenCounter.WithLabelValues(a.ID.String()).Add(float64(len(strings.Fields(r))))
			a.Trace(ctx, trace.EventToolEnd, map[string]any{"name": tc.Name, "result": r})
			step.ToolResults[tc.ID] = r
			msgs = append(msgs, model.ChatMessage{Role: "tool", ToolCallID: tc.ID, Content: r})
		}
		a.Mem.AddStep(step)
		_ = a.Checkpoint(ctx)
	}
	return "", errors.New("max iterations")
}

// SaveState persists the agent's memory under the given ID.
func (a *Agent) SaveState(ctx context.Context, id string) error {
	if a.Store == nil {
		return nil
	}
	data, err := json.Marshal(a.Mem.History())
	if err != nil {
		return err
	}
	return a.Store.Set(ctx, "history", id, data)
}

// LoadState restores memory from the store.
func (a *Agent) LoadState(ctx context.Context, id string) error {
	if a.Store == nil {
		return nil
	}
	b, err := a.Store.Get(ctx, "history", id)
	if err != nil || b == nil {
		return err
	}
	var steps []memory.Step
	if err := json.Unmarshal(b, &steps); err != nil {
		return err
	}
	a.Mem.SetHistory(steps)
	return nil
}

// Checkpoint persists the agent's current loop state under its ID.
func (a *Agent) Checkpoint(ctx context.Context) error {
	if a.Store == nil {
		return nil
	}
	data, err := json.Marshal(a.Mem.History())
	if err != nil {
		return err
	}
	return a.Store.Set(ctx, "checkpoint", a.ID.String(), data)
}

// Resume restores the agent's loop state from the store.
func (a *Agent) Resume(ctx context.Context) error {
	if a.Store == nil {
		return nil
	}
	b, err := a.Store.Get(ctx, "checkpoint", a.ID.String())
	if err != nil || b == nil {
		return err
	}
	var steps []memory.Step
	if err := json.Unmarshal(b, &steps); err != nil {
		return err
	}
	a.Mem.SetHistory(steps)
	return nil
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
		{Role: "system", Content: "You are an agent. Use the tools provided to answer the user's question. When you call a tool, `arguments` must be a valid JSON object (use {} if no parameters). Control characters are forbidden."},
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
