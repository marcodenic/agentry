package tui

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"

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
	// NOTE: Token counts are now handled by the agent's cost manager
	info.TokensStarted = false  // Reset tokens started flag
	info.StreamingResponse = "" // Reset streaming response
	info.Spinner = spinner.New()
	info.Spinner.Spinner = spinner.Line
	info.Spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.AIBarColor))

	// Clear initial logo on first user input
	if m.showInitialLogo {
		info.History = ""                       // Clear the logo content
		info.LastContentType = ContentTypeEmpty // Reset content type
		m.showInitialLogo = false
	}

	// Add user input with proper spacing logic
	userMessage := m.formatUserInput(m.userBar(), input, m.vp.Width)
	info.addContentWithSpacing(userMessage, ContentTypeUserInput)

	m.vp.SetContent(info.History)
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
	// Increase scanner buffer to handle large JSONL trace events (e.g., big tool results)
	// Default is 64K which is too small for some tool outputs.
	info.Scanner.Buffer(make([]byte, 0, 256*1024), 4*1024*1024)
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

// handleDiagnostics triggers lsp_diagnostics tool and parses results
func (m Model) handleDiagnostics() (Model, tea.Cmd) {
	if m.diagRunning {
		return m, nil
	}
	info := m.infos[m.active]
	if info == nil || info.Agent == nil {
		return m, nil
	}
	// show status start
	info.startProgressiveStatusUpdate("Running diagnostics", m)
	m.infos[m.active] = info
	m.vp.SetContent(info.History)
	m.vp.GotoBottom()
	m.diagRunning = true

	return m, func() tea.Msg {
		// Execute tool directly on Agent 0
		tl, ok := info.Agent.Tools["lsp_diagnostics"]
		if !ok {
			return errMsg{error: fmt.Errorf("lsp_diagnostics tool not available")}
		}
		out, err := tl.Execute(context.Background(), map[string]any{})
		type res struct {
			Diagnostics []struct {
				File     string `json:"file"`
				Line     int    `json:"line"`
				Col      int    `json:"col"`
				Code     string `json:"code"`
				Severity string `json:"severity"`
				Message  string `json:"message"`
			} `json:"diagnostics"`
			Ok    bool   `json:"ok"`
			Error string `json:"error"`
		}
		var r res
		if err == nil {
			_ = json.Unmarshal([]byte(out), &r)
		}
		return toolUseMsg{id: m.active, name: "lsp_diagnostics", args: map[string]any{"result": r, "raw": out, "err": err}}
	}
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
		m.vp.SetContent(info.History)
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
		m.vp.SetContent(info.History)
	}
	return m
}
