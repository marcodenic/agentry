package tui

import (
	"bufio"
	"context"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/marcodenic/agentry/internal/team"
	"github.com/marcodenic/agentry/internal/trace"
)

// startAgent runs an agent with the given input and streams its output.
func (m Model) startAgent(id uuid.UUID, input string) (Model, tea.Cmd) {
	info := m.infos[id]
	info.Status = StatusRunning
	info.TokenCount = 0 // Reset token count for new conversation
	info.TokensStarted = false // Reset tokens started flag
	info.StreamingResponse = "" // Reset streaming response
	info.Spinner = spinner.New()
	info.Spinner.Spinner = spinner.Line
	info.Spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.AIBarColor))
	
	// Clear initial logo on first user input
	if m.showInitialLogo {
		info.History = "" // Clear the logo content
		m.showInitialLogo = false
	}
	
	// Add user input with proper line wrapping and show thinking animation for responsive UX
	userMessage := m.formatWithBar(m.userBar(), input, m.vp.Width)
	info.History += userMessage + "\n\n"  // Add extra newline for spacing
	
	base := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.Palette.Foreground))
	m.vp.SetContent(base.Copy().Width(m.vp.Width).Render(info.History))
	m.vp.GotoBottom()

	pr, pw := io.Pipe()
	errCh := make(chan error, 1)
	completeCh := make(chan string, 1)
	tracer := trace.NewJSONL(pw)
	if info.Agent.Tracer != nil {
		info.Agent.Tracer = trace.NewMulti(info.Agent.Tracer, tracer)
	} else {
		info.Agent.Tracer = tracer
	}
	info.Scanner = bufio.NewScanner(pr)
	ctx := team.WithContext(context.Background(), m.team)
	ctx, cancel := context.WithCancel(ctx)
	info.Cancel = cancel
	m.infos[id] = info
	go func() {
		result, err := info.Agent.Run(ctx, input)
		pw.Close()
		if err != nil {
			errCh <- err
		} else {
			completeCh <- result
		}
	}()
	m.infos[id] = info
	return m, tea.Batch(m.readCmd(id), waitErr(errCh), waitComplete(id, completeCh), startThinkingAnimation(id))
}

// handleCommand parses a slash command and dispatches to the appropriate handler.
func (m Model) handleCommand(cmd string) (Model, tea.Cmd) {
	fields := strings.Fields(cmd)
	if len(fields) == 0 {
		return m, nil
	}
	switch fields[0] {
	case "/spawn":
		return m.handleSpawn(fields[1:])
	case "/switch":
		return m.handleSwitch(fields[1:])
	case "/stop":
		return m.handleStop(fields[1:])
	default:
		return m, nil
	}
}

// handleSpawn creates a new agent and adds it to the panel.
func (m Model) handleSpawn(args []string) (Model, tea.Cmd) {
	name := "agent"
	role := ""
	if len(args) > 0 {
		name = args[0]
	}
	if len(args) > 1 {
		role = args[1]
	}
	if len(m.agents) == 0 {
		return m, nil
	}
	ag := m.agents[0].Spawn()
	sp := spinner.New()
	sp.Spinner = spinner.Line
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.AIBarColor))
	info := &AgentInfo{
		Agent:           ag,
		Status:          StatusIdle,
		Spinner:         sp,
		Name:            name,
		Role:            role,
		ActivityData:    make([]float64, 0),
		ActivityTimes:   make([]time.Time, 0),
		CurrentActivity: 0,
		LastActivity:    time.Time{},
		TokenHistory:    []int{},
		TokensStarted:   false,
		StreamingResponse: "",
	}
	m.infos[ag.ID] = info
	m.order = append(m.order, ag.ID)
	if m.team != nil {
		m.team.Add(name, ag)
	}
	m.active = ag.ID
	m.vp.SetContent("")
	return m, nil
}

// handleSwitch focuses the agent whose ID prefix matches the argument.
func (m Model) handleSwitch(args []string) (Model, tea.Cmd) {
	if len(args) == 0 {
		return m, nil
	}
	prefix := args[0]
	for _, id := range m.order {
		if strings.HasPrefix(id.String(), prefix) {
			m.active = id
			if info, ok := m.infos[id]; ok {
				base := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.Palette.Foreground))
				m.vp.SetContent(base.Copy().Width(m.vp.Width).Render(info.History))
			}
			break
		}
	}
	return m, nil
}

// handleStop cancels a running agent.
func (m Model) handleStop(args []string) (Model, tea.Cmd) {
	id := m.active
	if len(args) > 0 {
		pref := args[0]
		for _, aid := range m.order {
			if strings.HasPrefix(aid.String(), pref) {
				id = aid
				break
			}
		}
	}
	if info, ok := m.infos[id]; ok {
		if info.Cancel != nil {
			info.Cancel()
		}
		info.Status = StatusStopped
		m.infos[id] = info
	}
	return m, nil
}

// cycleActive moves the focus to the next or previous agent.
func (m Model) cycleActive(delta int) Model {
	if len(m.order) == 0 {
		return m
	}
	idx := 0
	for i, id := range m.order {
		if id == m.active {
			idx = i
			break
		}
	}
	idx = (idx + delta + len(m.order)) % len(m.order)
	m.active = m.order[idx]
	if info, ok := m.infos[m.active]; ok {
		base := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.Palette.Foreground))
		m.vp.SetContent(base.Copy().Width(m.vp.Width).Render(info.History))
	}
	return m
}

// jumpToAgent sets the active agent by index.
func (m Model) jumpToAgent(index int) Model {
	if len(m.order) == 0 {
		return m
	}
	if index < 0 {
		index = 0
	}
	if index >= len(m.order) {
		index = len(m.order) - 1
	}
	m.active = m.order[index]
	if info, ok := m.infos[m.active]; ok {
		base := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.Palette.Foreground))
		m.vp.SetContent(base.Copy().Width(m.vp.Width).Render(info.History))
	}
	return m
}
