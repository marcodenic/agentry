package core

import (
	"context"
	"encoding/json"

	"github.com/marcodenic/agentry/internal/model"

	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/trace"
)

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

// Exported for use in team mode and other packages
func BuildMessages(prompt string, vars map[string]string, hist []memory.Step, input string) []model.ChatMessage {
	if prompt == "" {
		prompt = defaultPrompt()
	}

	// Inject platform-specific guidance
	prompt = InjectPlatformContextLegacy(prompt)

	prompt = applyVars(prompt, vars)
	msgs := []model.ChatMessage{{Role: "system", Content: prompt}}
	for _, h := range hist {
		msgs = append(msgs, model.ChatMessage{Role: "assistant", Content: h.Output, ToolCalls: h.ToolCalls})
		for id, res := range h.ToolResults {
			msgs = append(msgs, model.ChatMessage{Role: "tool", ToolCallID: id, Content: res})
		}
	}
	msgs = append(msgs, model.ChatMessage{Role: "user", Content: input})
	return msgs
}
