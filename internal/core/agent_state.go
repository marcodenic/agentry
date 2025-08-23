package core

import (
	"context"
	"encoding/json"

	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/memstore"
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
