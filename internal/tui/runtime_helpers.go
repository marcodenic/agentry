package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/marcodenic/agentry/internal/trace"
)

func (m *Model) formatToolAction(toolName string, args map[string]any) string {
	// Format tool start messages using glyphs instead of emojis
	switch toolName {
	case "view", "read":
		if path, ok := args["path"].(string); ok {
			return fmt.Sprintf("▧ Reading %s", path)
		}
		return "▧ Reading file"
	case "write":
		if path, ok := args["path"].(string); ok {
			return fmt.Sprintf("✎ Writing to %s", path)
		}
		return "✎ Writing file"
	case "edit", "patch":
		if path, ok := args["path"].(string); ok {
			return fmt.Sprintf("✏ Editing %s", path)
		}
		return "✏ Editing file"
	case "ls", "list":
		if path, ok := args["path"].(string); ok {
			return fmt.Sprintf("▨ Listing %s", path)
		}
		return "▨ Listing directory"
	case "bash", "powershell", "cmd":
		return "▶ Running command"
	case "agent":
		if agent, ok := args["agent"].(string); ok {
			return fmt.Sprintf("⚡ Delegating to %s agent", agent)
		}
		return "⚡ Delegating task"
	case "grep", "search":
		if query, ok := args["query"].(string); ok {
			return fmt.Sprintf("⌕ Searching for '%s'", query)
		}
		return "⌕ Searching"
	case "fetch":
		if url, ok := args["url"].(string); ok {
			return fmt.Sprintf("⊡ Fetching %s", url)
		}
		return "⊡ Fetching data"
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
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return activityTickMsg{}
	})
}

func startThinkingAnimation(id uuid.UUID) tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		frame := int(t.UnixMilli()/100) % len(spinnerFrames)
		return thinkingAnimationMsg{id: id, frame: frame}
	})
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func (m *Model) addDebugTraceEvent(id uuid.UUID, ev trace.Event) {
	info := m.infos[id]
	if info == nil {
		return
	}

	var dataMap map[string]interface{}
	if ev.Data != nil {
		if dm, ok := ev.Data.(map[string]interface{}); ok {
			dataMap = dm
		} else {
			dataMap = map[string]interface{}{"value": ev.Data}
		}
	}

	var details string
	switch ev.Type {
	case trace.EventModelStart:
		if name, ok := ev.Data.(string); ok {
			details = fmt.Sprintf("Started model: %s", name)
		} else {
			details = "Model started"
		}
	case trace.EventStepStart:
		info.CurrentStep++
		info.DebugStreamingResponse = ""
		if completion, ok := ev.Data.(map[string]interface{}); ok {
			if content, ok := completion["Content"].(string); ok && content != "" {
				details = fmt.Sprintf("New reasoning step with content: %s", truncateString(content, 100))
			} else {
				details = "Starting new reasoning step"
			}
		} else if res, ok := ev.Data.(string); ok && res != "" {
			details = fmt.Sprintf("New reasoning step: %s", truncateString(res, 100))
		} else {
			details = "Starting new reasoning step"
		}
	case trace.EventToken:
		if token, ok := ev.Data.(string); ok {
			if info.DebugStreamingResponse == "" {
				info.DebugStreamingResponse = token
			} else {
				info.DebugStreamingResponse += token
			}
			if len(token) > 1 || token == " " || token == "\n" || token == "." || token == "!" || token == "?" {
				details = fmt.Sprintf("Token: %q", token)
			} else {
				details = fmt.Sprintf("Character: %q", token)
			}
		}
	case trace.EventToolStart:
		if m2, ok := ev.Data.(map[string]any); ok {
			if name, ok := m2["name"].(string); ok {
				if argsRaw, ok := m2["args"]; ok {
					argsStr := fmt.Sprintf("%v", argsRaw)
					if len(argsStr) > 100 {
						argsStr = argsStr[:100] + "... [truncated]"
					}
					details = fmt.Sprintf("Tool called: %s with args: %s", name, argsStr)
				} else {
					details = fmt.Sprintf("Tool called: %s", name)
				}
			}
		}
	case trace.EventToolEnd:
		if m2, ok := ev.Data.(map[string]any); ok {
			if name, ok := m2["name"].(string); ok {
				if result, ok := m2["result"].(string); ok {
					displayResult := result
					if len(result) > 200 {
						displayResult = result[:200] + "... [truncated]"
					}
					details = fmt.Sprintf("Tool %s completed: %s", name, displayResult)
				} else {
					details = fmt.Sprintf("Tool %s completed", name)
				}
			}
		}
	case trace.EventFinal:
		if result, ok := ev.Data.(string); ok && result != "" {
			details = fmt.Sprintf("Final result: %s", truncateString(result, 150))
		} else {
			if info.DebugStreamingResponse != "" {
				details = fmt.Sprintf("Processing completed - Response: %s", truncateString(info.DebugStreamingResponse, 150))
				info.DebugStreamingResponse = ""
			} else {
				details = "Processing completed"
			}
		}
	case trace.EventYield:
		details = "Agent yielded (iteration limit reached)"
	case trace.EventSummary:
		details = "Summary with token and cost statistics"
	default:
		details = fmt.Sprintf("Event type: %s", string(ev.Type))
	}

	debugEvent := DebugTraceEvent{
		Timestamp: ev.Timestamp,
		Type:      string(ev.Type),
		Data:      dataMap,
		StepNum:   info.CurrentStep,
		Details:   details,
	}

	info.DebugTrace = append(info.DebugTrace, debugEvent)

	if id == m.active && m.activeTab == 1 {
		debugContent := m.renderDetailedMemory(info.Agent)
		m.debugVp.SetContent(debugContent)
		m.debugVp.GotoBottom()
	}
}
