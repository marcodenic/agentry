package core

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/marcodenic/agentry/internal/cost"
	"github.com/marcodenic/agentry/internal/debug"
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
	ID            uuid.UUID
	Prompt        string
	Vars          map[string]string
	Tools         tool.Registry
	Mem           memory.Store
	Vector        memory.VectorStore
	Route         router.Selector
	Tracer        trace.Writer
	Store         memstore.KV
	Cost          *cost.Manager
	MaxIterations int
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

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// getToolNames extracts tool names from a registry for debugging
func getToolNames(reg tool.Registry) []string {
	var names []string
	for name := range reg {
		names = append(names, name)
	}
	return names
}

func New(sel router.Selector, reg tool.Registry, mem memory.Store, store memstore.KV, vec memory.VectorStore, tr trace.Writer) *Agent {
	return &Agent{ID: uuid.New(), Tools: reg, Mem: mem, Vector: vec, Route: sel, Tracer: tr, Store: store, Cost: nil, MaxIterations: 8}
}

func (a *Agent) Spawn() *Agent {
	spawned := &Agent{
		ID:            uuid.New(),
		Prompt:        a.Prompt, // Ensure prompt is inherited by sub-agents
		Vars:          a.Vars,
		Tools:         a.Tools,
		Mem:           memory.NewInMemory(),
		Vector:        a.Vector,
		Route:         a.Route,
		Tracer:        a.Tracer,
		Store:         a.Store,
		Cost:          a.Cost,
		MaxIterations: a.MaxIterations,
	}
	
	debug.Printf("Agent.Spawn: Parent ID=%s, Spawned ID=%s", a.ID.String()[:8], spawned.ID.String()[:8])
	debug.Printf("Agent.Spawn: Inherited prompt length=%d chars", len(spawned.Prompt))
	debug.Printf("Agent.Spawn: Inherited prompt preview: %s", spawned.Prompt[:min(100, len(spawned.Prompt))])
	
	return spawned
}

func (a *Agent) Run(ctx context.Context, input string) (string, error) {
	debug.Printf("Agent.Run: Agent ID=%s, Prompt length=%d chars", a.ID.String()[:8], len(a.Prompt))
	debug.Printf("Agent.Run: Available tools: %v", getToolNames(a.Tools))
	debug.Printf("Agent.Run: Input: %s", input[:min(100, len(input))])
	
	client, name := a.Route.Select(input)
	a.Trace(ctx, trace.EventModelStart, name)
	msgs := BuildMessages(a.Prompt, a.Vars, a.Mem.History(), input)
	specs := tool.BuildSpecs(a.Tools)
	inTok := len(strings.Fields(input))
	tokenCounter.WithLabelValues(a.ID.String()).Add(float64(inTok))
	if a.Cost != nil {
		if a.Cost.AddModel(name, inTok) {
			if a.Cost.OverBudget() {
				debug.Printf("budget exceeded")
			}
		}
	}
	limit := a.MaxIterations
	if limit <= 0 {
		limit = 8
	}
	for i := 0; i < limit; i++ {
		res, err := client.Complete(ctx, msgs, specs)
		if err != nil {
			return "", err
		}

		a.Trace(ctx, trace.EventStepStart, res)
		outTok := len(strings.Fields(res.Content))
		tokenCounter.WithLabelValues(a.ID.String()).Add(float64(outTok))
		if a.Cost != nil {
			if a.Cost.AddModel(name, outTok) && a.Cost.OverBudget() {
				debug.Printf("budget exceeded")
			}
		}
		msgs = append(msgs, model.ChatMessage{Role: "assistant", Content: res.Content, ToolCalls: res.ToolCalls})
		step := memory.Step{Output: res.Content, ToolCalls: res.ToolCalls, ToolResults: map[string]string{}}
		if len(res.ToolCalls) == 0 {
			a.Mem.AddStep(step)
			_ = a.Checkpoint(ctx)
			
			// Emit token events for streaming effect with proper formatting preservation
			// Process character by character to preserve newlines and formatting
			for _, r := range res.Content {
				a.Trace(ctx, trace.EventToken, string(r))
				// Small delay to simulate real streaming, but faster for better UX
				time.Sleep(20 * time.Millisecond)
				
				// Optional: Add slight extra delay after punctuation for more natural feel
				if r == '.' || r == '!' || r == '?' || r == '\n' {
					time.Sleep(80 * time.Millisecond)
				}
			}
			
			// FIXED: Only emit final message, don't duplicate the content
			a.Trace(ctx, trace.EventFinal, "")
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
			applyVarsMap(args, a.Vars)
			
			// Debug: Log tool execution details for coder agent
			debug.Printf("Agent '%s' executing tool '%s' with args: %v", a.ID, tc.Name, args)
			
			a.Trace(ctx, trace.EventToolStart, map[string]any{"name": tc.Name, "args": args})
			start := time.Now()
			r, err := t.Execute(ctx, args)
			toolLatency.WithLabelValues(a.ID.String(), tc.Name).Observe(time.Since(start).Seconds())
			if err != nil {
				debug.Printf("Agent '%s' tool '%s' failed: %v", a.ID, tc.Name, err)
				return "", err
			}
			
			debug.Printf("Agent '%s' tool '%s' succeeded, result length: %d", a.ID, tc.Name, len(r))

			tok := len(strings.Fields(r))
			tokenCounter.WithLabelValues(a.ID.String()).Add(float64(tok))
			if a.Cost != nil {
				if a.Cost.AddTool(tc.Name, tok) && a.Cost.OverBudget() {
					debug.Printf("budget exceeded")
				}
			}
			a.Trace(ctx, trace.EventToolEnd, map[string]any{"name": tc.Name, "result": r})
			step.ToolResults[tc.ID] = r
			msgs = append(msgs, model.ChatMessage{Role: "tool", ToolCallID: tc.ID, Content: r})
		}
		a.Mem.AddStep(step)
		_ = a.Checkpoint(ctx)
	}
	a.Trace(ctx, trace.EventYield, nil)
	return "", nil
}
