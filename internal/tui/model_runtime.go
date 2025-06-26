package tui

import (
	"encoding/json"
	"fmt"
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
	args map[string]any
}

type thinkingMsg struct {
	id   uuid.UUID
	text string
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

type errMsg struct{ error }

type agentCompleteMsg struct {
	id     uuid.UUID
	result string
}

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
		case trace.EventStepStart:
			// Show thinking indicator when agent starts reasoning
			return thinkingMsg{id: id, text: "🤔 Thinking..."}
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

// formatToolAction creates user-friendly action messages
func (m *Model) formatToolAction(toolName string, args map[string]any) string {
	switch toolName {
	case "view", "read":
		if path, ok := args["path"].(string); ok {
			return fmt.Sprintf("🔍 Reading %s...", path)
		}
		return "🔍 Reading file..."
	case "write":
		if path, ok := args["path"].(string); ok {
			return fmt.Sprintf("✏️ Writing to %s...", path)
		}
		return "✏️ Writing file..."
	case "edit", "patch":
		if path, ok := args["path"].(string); ok {
			return fmt.Sprintf("✏️ Editing %s...", path)
		}
		return "✏️ Editing file..."
	case "ls", "list":
		if path, ok := args["path"].(string); ok {
			return fmt.Sprintf("📁 Listing %s...", path)
		}
		return "📁 Listing directory..."
	case "bash", "powershell", "cmd":
		return "⚡ Running command..."
	case "agent":
		if agent, ok := args["agent"].(string); ok {
			return fmt.Sprintf("🤖 Delegating to %s agent...", agent)
		}
		return "🤖 Delegating task..."
	case "grep", "search":
		if query, ok := args["query"].(string); ok {
			return fmt.Sprintf("🔍 Searching for '%s'...", query)
		}
		return "🔍 Searching..."
	case "fetch":
		if url, ok := args["url"].(string); ok {
			return fmt.Sprintf("🌐 Fetching %s...", url)
		}
		return "🌐 Fetching data..."
	default:
		return fmt.Sprintf("🔧 Using %s...", toolName)
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

func waitComplete(id uuid.UUID, ch <-chan string) tea.Cmd {
	return func() tea.Msg {
		result := <-ch
		return agentCompleteMsg{id: id, result: result}
	}
}

func (m Model) Init() tea.Cmd {
	// Start the activity chart ticker (update every second)
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return activityTickMsg{}
	})
}
