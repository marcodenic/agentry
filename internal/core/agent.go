package core

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/marcodenic/agentry/internal/cost"
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

func New(sel router.Selector, reg tool.Registry, mem memory.Store, store memstore.KV, vec memory.VectorStore, tr trace.Writer) *Agent {
	return &Agent{ID: uuid.New(), Tools: reg, Mem: mem, Vector: vec, Route: sel, Tracer: tr, Store: store, Cost: nil, MaxIterations: 8}
}

func (a *Agent) Spawn() *Agent {
	return &Agent{
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
}

func (a *Agent) Run(ctx context.Context, input string) (string, error) {
	client, name := a.Route.Select(input)
	a.Trace(ctx, trace.EventModelStart, name)
	msgs := BuildMessages(a.Prompt, a.Vars, a.Mem.History(), input)
	specs := tool.BuildSpecs(a.Tools)
	inTok := len(strings.Fields(input))
	tokenCounter.WithLabelValues(a.ID.String()).Add(float64(inTok))
	if a.Cost != nil {
		if a.Cost.AddModel(name, inTok) {
			if a.Cost.OverBudget() {
				log.Printf("budget exceeded")
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
		// DEBUG: Show Agent 0 raw response for debugging in the TUI
		if a.Prompt != "" && strings.Contains(a.Prompt, "Agent 0") {
			debugInfo := fmt.Sprintf("🔍 [AGENT 0 DEBUG] Raw AI Response:\n   Content: %s\n   Tool Calls: %d\n", res.Content, len(res.ToolCalls))
			for j, tc := range res.ToolCalls {
				debugInfo += fmt.Sprintf("   Tool Call %d: %s(%s)\n", j+1, tc.Name, string(tc.Arguments))
			}
			debugInfo += "───────────────────────────────────────\n"
			// Prepend debug info to make it more visible
			res.Content = debugInfo + res.Content

			// Also log for console debugging
			log.Printf("🔍 AGENT_0_DEBUG: Raw AI Response:")
			log.Printf("  Content: %s", res.Content)
			log.Printf("  Tool Calls Count: %d", len(res.ToolCalls))
			for j, tc := range res.ToolCalls {
				log.Printf("  Tool Call %d: Name=%s, Args=%s", j+1, tc.Name, string(tc.Arguments))
			}
		}

		a.Trace(ctx, trace.EventStepStart, res)
		outTok := len(strings.Fields(res.Content))
		tokenCounter.WithLabelValues(a.ID.String()).Add(float64(outTok))
		if a.Cost != nil {
			if a.Cost.AddModel(name, outTok) && a.Cost.OverBudget() {
				log.Printf("budget exceeded")
			}
		}
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
				// DEBUG: Show available tools when agent tool lookup fails
				debugMsg := ""
				if a.Prompt != "" && strings.Contains(a.Prompt, "Agent 0") {
					debugMsg = fmt.Sprintf("\n🔍 AGENT_0_DEBUG: Tool lookup failed!\n  Requested tool: %s\n  Available tools:\n", tc.Name)
					for toolName := range a.Tools {
						debugMsg += fmt.Sprintf("    - %s\n", toolName)
					}

					// Also log for console debugging
					log.Printf("🔍 AGENT_0_DEBUG: Tool lookup failed!")
					log.Printf("  Requested tool: %s", tc.Name)
					log.Printf("  Available tools:")
					for toolName := range a.Tools {
						log.Printf("    - %s", toolName)
					}
				}
				return debugMsg + fmt.Sprintf("ERR: unknown tool: %s", tc.Name), fmt.Errorf("unknown tool: %s", tc.Name)
			}
			var args map[string]any
			if err := json.Unmarshal(tc.Arguments, &args); err != nil {
				return "", err
			}
			applyVarsMap(args, a.Vars)
			a.Trace(ctx, trace.EventToolStart, map[string]any{"name": tc.Name, "args": args})
			start := time.Now()
			r, err := t.Execute(ctx, args)
			toolLatency.WithLabelValues(a.ID.String(), tc.Name).Observe(time.Since(start).Seconds())
			if err != nil {
				return "", err
			}

			// DEBUG: Show tool execution results for Agent 0
			if a.Prompt != "" && strings.Contains(a.Prompt, "Agent 0") {
				log.Printf("🔍 AGENT_0_DEBUG: Tool '%s' executed successfully", tc.Name)
				log.Printf("  Result length: %d chars", len(r))
				if len(r) > 200 {
					log.Printf("  Result preview: %s...", r[:200])
				} else {
					log.Printf("  Result: %s", r)
				}
			}
			tok := len(strings.Fields(r))
			tokenCounter.WithLabelValues(a.ID.String()).Add(float64(tok))
			if a.Cost != nil {
				if a.Cost.AddTool(tc.Name, tok) && a.Cost.OverBudget() {
					log.Printf("budget exceeded")
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
