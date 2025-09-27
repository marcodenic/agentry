package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/marcodenic/agentry/internal/glyphs"
	"github.com/marcodenic/agentry/internal/trace"
)

func (m *Model) formatToolAction(toolName string, args map[string]any) string {
	// Format tool start messages using glyphs instead of emojis
	switch toolName {
	case "view", "read":
		if path, ok := args["path"].(string); ok {
			return fmt.Sprintf("%s Reading %s", glyphs.BlueCircle(), path)
		}
		return glyphs.BlueCircle() + " Reading file"
	case "write":
		if path, ok := args["path"].(string); ok {
			return fmt.Sprintf("%s Writing to %s", glyphs.GreenCheckmark(), path)
		}
		return glyphs.GreenCheckmark() + " Writing file"
	case "edit", "patch":
		// If we have a direct path (for edit_* tools), show it
		if path, ok := args["path"].(string); ok && path != "" {
			return fmt.Sprintf("%s Editing %s", glyphs.YellowStar(), path)
		}
		// For the patch tool, extract filenames from the unified diff
		if toolName == "patch" {
			if pstr, ok := args["patch"].(string); ok && pstr != "" {
				files := extractPatchFiles(pstr)
				if len(files) == 1 {
					return fmt.Sprintf("%s Patching %s", glyphs.YellowStar(), files[0])
				}
				if len(files) > 1 {
					// Limit to first few files for brevity
					display := files
					if len(display) > 3 {
						display = append(display[:3], "…")
					}
					return fmt.Sprintf("%s Patching %s", glyphs.YellowStar(), strings.Join(display, ", "))
				}
			}
		}
		return glyphs.YellowStar() + " Editing file"
	case "edit_range":
		path, _ := args["path"].(string)
		start := 0
		if v, ok := args["start_line"].(int); ok {
			start = v
		} else if v, ok := args["start_line"].(float64); ok {
			start = int(v)
		}
		end := 0
		if v, ok := args["end_line"].(int); ok {
			end = v
		} else if v, ok := args["end_line"].(float64); ok {
			end = int(v)
		}
		if path != "" && start > 0 && end > 0 {
			return fmt.Sprintf("%s Editing %s lines %d-%d", glyphs.YellowStar(), path, start, end)
		}
		if path != "" {
			return fmt.Sprintf("%s Editing %s (range)", glyphs.YellowStar(), path)
		}
		return glyphs.YellowStar() + " Editing range"
	case "insert_at":
		path, _ := args["path"].(string)
		lineNum := -1
		if v, ok := args["line"].(int); ok {
			lineNum = v
		} else if v, ok := args["line"].(float64); ok {
			lineNum = int(v)
		}
		if path != "" {
			if lineNum >= 0 {
				return fmt.Sprintf("%s Inserting into %s at line %d", glyphs.YellowStar(), path, lineNum)
			}
			return fmt.Sprintf("%s Inserting into %s", glyphs.YellowStar(), path)
		}
		return glyphs.YellowStar() + " Inserting into file"
	case "search_replace":
		path, _ := args["path"].(string)
		search, _ := args["search"].(string)
		replace, _ := args["replace"].(string)
		regex := false
		if v, ok := args["regex"].(bool); ok {
			regex = v
		}
		sr := truncateString(search, 60)
		rp := truncateString(replace, 60)
		if path != "" && search != "" {
			if regex {
				return fmt.Sprintf("%s Replacing /%s/ -> %q in %s", glyphs.YellowStar(), sr, rp, path)
			}
			return fmt.Sprintf("%s Replacing %q -> %q in %s", glyphs.YellowStar(), sr, rp, path)
		}
		return glyphs.YellowStar() + " Search/replace in file"
	case "create":
		if path, ok := args["path"].(string); ok && path != "" {
			return fmt.Sprintf("%s Creating %s", glyphs.GreenCheckmark(), path)
		}
		return glyphs.GreenCheckmark() + " Creating file"
	case "ls", "list":
		if path, ok := args["path"].(string); ok && path != "" {
			return fmt.Sprintf("%s Listing %s", glyphs.BlueCircle(), path)
		}
		return glyphs.BlueCircle() + " Listing directory"
	case "bash", "powershell", "cmd":
		if cmd, ok := args["command"].(string); ok && cmd != "" {
			return fmt.Sprintf("%s Running: %s", glyphs.OrangeTriangle(), truncateString(cmd, 80))
		}
		return glyphs.OrangeTriangle() + " Running command"
	case "agent":
		if agent, ok := args["agent"].(string); ok {
			if input, ok := args["input"].(string); ok && input != "" {
				return fmt.Sprintf("%s Delegating to %s: %s", glyphs.OrangeLightning(), agent, truncateString(input, 80))
			}
			return fmt.Sprintf("%s Delegating to %s agent", glyphs.OrangeLightning(), agent)
		}
		return glyphs.OrangeLightning() + " Delegating task"
	case "grep":
		pattern, _ := args["pattern"].(string)
		path := "."
		if p, ok := args["path"].(string); ok && p != "" {
			path = p
		}
		if pattern != "" {
			return fmt.Sprintf("%s Grep %q in %s", glyphs.YellowStar(), truncateString(pattern, 60), path)
		}
		return glyphs.YellowStar() + " Searching"
	case "find":
		name, _ := args["name"].(string)
		base := "."
		if p, ok := args["path"].(string); ok && p != "" {
			base = p
		}
		if name != "" {
			return fmt.Sprintf("%s Finding %q under %s", glyphs.BlueCircle(), name, base)
		}
		return glyphs.BlueCircle() + " Finding files"
	case "glob":
		pattern, _ := args["pattern"].(string)
		base := "."
		if p, ok := args["path"].(string); ok && p != "" {
			base = p
		}
		if pattern != "" {
			return fmt.Sprintf("%s Glob %q under %s", glyphs.BlueCircle(), pattern, base)
		}
		return glyphs.BlueCircle() + " Glob search"
	case "web_search":
		query, _ := args["query"].(string)
		provider := "duckduckgo"
		if v, ok := args["provider"].(string); ok && v != "" {
			provider = v
		}
		if query != "" {
			return fmt.Sprintf("%s Web search (%s): %s", glyphs.YellowStar(), provider, truncateString(query, 80))
		}
		return glyphs.YellowStar() + " Web search"
	case "lsp_diagnostics":
		// Show brief info about scope
		if paths, ok := args["paths"].([]any); ok && len(paths) > 0 {
			return fmt.Sprintf("%s Diagnostics on %d path(s)", glyphs.BlueCircle(), len(paths))
		}
		return glyphs.BlueCircle() + " Diagnostics: scanning workspace"
	case "project_tree":
		// Provide details about what path/depth/files are requested
		path := "."
		if p, ok := args["path"].(string); ok && p != "" {
			path = p
		}
		depth := 0
		if v, ok := args["depth"].(int); ok {
			depth = v
		} else if v, ok := args["depth"].(float64); ok {
			depth = int(v)
		}
		showFiles := true
		if v, ok := args["show_files"].(bool); ok {
			showFiles = v
		}
		desc := fmt.Sprintf("%s Using project_tree on %s", glyphs.YellowStar(), path)
		if depth > 0 {
			desc += fmt.Sprintf(" (depth=%d)", depth)
		}
		if !showFiles {
			desc += " (dirs only)"
		}
		return desc
	case "fetch":
		if url, ok := args["url"].(string); ok {
			return fmt.Sprintf("%s Fetching %s", glyphs.BlueCircle(), url)
		}
		return glyphs.BlueCircle() + " Fetching data"
	default:
		return fmt.Sprintf("%s Using %s...", glyphs.YellowStar(), toolName)
	}
}

func (m Model) Init() tea.Cmd {
	var cmds []tea.Cmd

	// Start activity tick
	cmds = append(cmds, tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return activityTickMsg{}
	}))

	// Start spinner ticks for all agents
	for _, info := range m.infos {
		cmds = append(cmds, info.Spinner.Tick)
	}

	return tea.Batch(cmds...)
}
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// extractPatchFiles returns a list of file paths mentioned in a unified diff.
// It mirrors the logic used in tool.parsePatchFiles but is duplicated here to avoid import cycles.
func extractPatchFiles(patchStr string) []string {
	var files []string
	for _, line := range strings.Split(patchStr, "\n") {
		if strings.HasPrefix(line, "+++ ") {
			f := strings.TrimPrefix(line, "+++ ")
			f = strings.TrimPrefix(f, "b/")
			if f != "/dev/null" && f != "" {
				files = append(files, f)
			}
		}
	}
	return files
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
		details = "Agent yielded"
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

	if id == m.active && m.layout.activeTab == 1 {
		debugContent := m.renderDetailedMemory(info.Agent)
		m.view.Chat.Debug.SetContent(debugContent)
		m.view.Chat.Debug.GotoBottom()
	}
}
