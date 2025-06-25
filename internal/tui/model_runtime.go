package tui

import (
	"encoding/json"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/marcodenic/agentry/internal/trace"
)

type tokenMsg struct {
	id    uuid.UUID
	token string
}

type toolUseMsg struct {
	id   uuid.UUID
	name string
}

type modelMsg struct {
	id   uuid.UUID
	name string
}

type activityTickMsg struct{}

type errMsg struct{ error }

type finalMsg struct {
	id   uuid.UUID
	text string
}

func streamTokens(id uuid.UUID, out string) tea.Cmd {
	runes := []rune(out)
	cmds := make([]tea.Cmd, len(runes))
	for i, r := range runes {
		tok := string(r)
		delay := time.Duration(i*30) * time.Millisecond
		cmds[i] = tea.Tick(delay, func(t time.Time) tea.Msg { return tokenMsg{id: id, token: tok} })
	}
	return tea.Batch(cmds...)
}

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
		switch ev.Type {
		case trace.EventFinal:
			if s, ok := ev.Data.(string); ok {
				return finalMsg{id: id, text: s}
			}
		case trace.EventModelStart:
			if name, ok := ev.Data.(string); ok {
				return modelMsg{id: id, name: name}
			}
		case trace.EventToolEnd:
			if m2, ok := ev.Data.(map[string]any); ok {
				if name, ok := m2["name"].(string); ok {
					return toolUseMsg{id: id, name: name}
				}
			}
		default:
			continue
		}
	}
}

func (m *Model) readCmd(id uuid.UUID) tea.Cmd {
	return func() tea.Msg { return m.readEvent(id) }
}

func waitErr(ch <-chan error) tea.Cmd {
	return func() tea.Msg {
		if err := <-ch; err != nil {
			return errMsg{err}
		}
		return nil
	}
}

func (m Model) Init() tea.Cmd {
	// Start the activity chart ticker (update every second)
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return activityTickMsg{}
	})
}
