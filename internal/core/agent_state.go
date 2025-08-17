package core

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"strings"

	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/memstore"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/trace"
)

func stateKey(id string, a *Agent) string {
	if id != "" {
		return id
	}
	// default to agent UUID
	return a.ID.String()
}

func (a *Agent) SaveState(ctx context.Context, id string) error {
	key := stateKey(id, a)
	payload := struct {
		Prompt string            `json:"prompt"`
		Vars   map[string]string `json:"vars"`
		Hist   []memory.Step     `json:"hist"`
		Model  string            `json:"model"`
	}{Prompt: a.Prompt, Vars: a.Vars, Hist: a.Mem.History(), Model: a.ModelName}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return memstore.Get().Set("agent-state", key, b, 0)
}

// LoadState restores memory from the store.
func (a *Agent) LoadState(ctx context.Context, id string) error {
	key := stateKey(id, a)
	b, ok, err := memstore.Get().Get("agent-state", key)
	if err != nil || !ok {
		return err
	}
	var payload struct {
		Prompt string            `json:"prompt"`
		Vars   map[string]string `json:"vars"`
		Hist   []memory.Step     `json:"hist"`
		Model  string            `json:"model"`
	}
	if err := json.Unmarshal(b, &payload); err != nil {
		return err
	}
	a.Prompt = payload.Prompt
	a.Vars = payload.Vars
	a.ModelName = payload.Model
	a.Mem.SetHistory(payload.Hist)
	return nil
}

// Checkpoint persists the agent's current loop state under its ID.
func (a *Agent) Checkpoint(ctx context.Context) error {
	return a.SaveState(ctx, "")
}

// Resume restores the agent's loop state from the store.
func (a *Agent) Resume(ctx context.Context) error {
	return a.LoadState(ctx, "")
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

// Exported for use in team mode and other packages
func BuildMessages(prompt string, vars map[string]string, hist []memory.Step, input string) []model.ChatMessage {
	if prompt == "" {
		prompt = defaultPrompt()
	}

	// Inject platform-specific guidance
	prompt = InjectPlatformContextLegacy(prompt)

	prompt = applyVars(prompt, vars)
	msgs := []model.ChatMessage{{Role: "system", Content: prompt}}

	compactAfter := getenvIntFallback("AGENTRY_HISTORY_COMPACT_AFTER", 0)
	keepNewest := getenvIntFallback("AGENTRY_HISTORY_KEEP", 8)
	if compactAfter > 0 && len(hist) > compactAfter {
		if keepNewest < 1 {
			keepNewest = 1
		}
		newHist, summary := compactHistory(hist, keepNewest)
		if summary != "" {
			msgs = append(msgs, model.ChatMessage{Role: "system", Content: summary})
		}
		hist = newHist
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

func getenvIntFallback(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

func compactHistory(hist []memory.Step, keepNewest int) ([]memory.Step, string) {
	if keepNewest >= len(hist) {
		return hist, ""
	}
	cut := len(hist) - keepNewest
	older := hist[:cut]
	newer := hist[cut:]
	var totalToolCalls int
	toolFreq := map[string]int{}
	for _, step := range older {
		totalToolCalls += len(step.ToolCalls)
		for _, tc := range step.ToolCalls {
			toolFreq[tc.Name]++
		}
	}
	type kv struct {
		name string
		n    int
	}
	tops := make([]kv, 0, len(toolFreq))
	for k, v := range toolFreq {
		tops = append(tops, kv{k, v})
	}
	for i := 0; i < len(tops)-1; i++ {
		for j := i + 1; j < len(tops); j++ {
			if tops[j].n > tops[i].n {
				tops[i], tops[j] = tops[j], tops[i]
			}
		}
	}
	if len(tops) > 5 {
		tops = tops[:5]
	}
	var b strings.Builder
	b.WriteString("[HISTORY COMPACTED] Earlier ")
	b.WriteString(strconv.Itoa(cut))
	b.WriteString(" steps summarized (" + strconv.Itoa(totalToolCalls) + " tool calls). Top tools: ")
	if len(tops) == 0 {
		b.WriteString("none")
	} else {
		for i, t := range tops {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(t.name + "x" + strconv.Itoa(t.n))
		}
	}
	b.WriteString(". Recent context retained.")
	return newer, b.String()
}
