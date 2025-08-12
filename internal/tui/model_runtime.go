package tui

import (
	"encoding/json"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/marcodenic/agentry/internal/trace"
)

type tokenMsg struct {
	id    uuid.UUID
	token string
}

type startTokenStream struct {
	id    uuid.UUID
	runes []rune
}

type tokenStreamTick struct {
	id       uuid.UUID
	runes    []rune
	position int
}

type toolUseMsg struct {
	id   uuid.UUID
	name string
	args map[string]any
}

type thinkingMsg struct {
	id   uuid.UUID
	text string
}

type thinkingAnimationMsg struct {
	id    uuid.UUID
	frame int
}

type actionMsg struct {
	id   uuid.UUID
	text string
}

type modelMsg struct {
	id   uuid.UUID
	name string
}

type activityTickMsg struct{}

type refreshMsg struct{}

type errMsg struct{ error }

type agentCompleteMsg struct {
	id     uuid.UUID
	result string
}

type agentStartMsg struct {
	id   uuid.UUID
	name string
	role string
}

type finalMsg struct {
	id   uuid.UUID
	text string
}

// ASCII spinner frames for thinking animation
var spinnerFrames = []string{"|", "/", "-", "\\"}

func (m *Model) readEvent(id uuid.UUID) tea.Msg {
	info := m.infos[id]
	if info == nil || info.Scanner == nil {
		return nil
	}
	for {
		if !info.Scanner.Scan() {
			if err := info.Scanner.Err(); err != nil {
				return errMsg{err}
			}
			return nil
		}
		var ev trace.Event
		if err := json.Unmarshal(info.Scanner.Bytes(), &ev); err != nil {
			return errMsg{err}
		}

		// Capture all events for debug trace
		m.addDebugTraceEvent(id, ev)

		switch ev.Type {
		case trace.EventToken:
			if s, ok := ev.Data.(string); ok {
				return tokenMsg{id: id, token: s}
			}
		case trace.EventFinal:
			if s, ok := ev.Data.(string); ok {
				return finalMsg{id: id, text: s}
			}
		case trace.EventModelStart:
			if name, ok := ev.Data.(string); ok {
				return modelMsg{id: id, name: name}
			}
		case trace.EventStepStart:
			// Skip - thinking animation is handled immediately when user sends input
			continue
		case trace.EventToolStart:
			if m2, ok := ev.Data.(map[string]any); ok {
				if name, ok := m2["name"].(string); ok {
					args, _ := m2["args"].(map[string]any)
					actionText := m.formatToolAction(name, args)
					return actionMsg{id: id, text: actionText}
				}
			}
		case trace.EventToolEnd:
			if m2, ok := ev.Data.(map[string]any); ok {
				if name, ok := m2["name"].(string); ok {
					args, _ := m2["args"].(map[string]any)
					return toolUseMsg{id: id, name: name, args: args}
				}
			}
		default:
			continue
		}
	}
}
