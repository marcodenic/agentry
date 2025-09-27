package tui

import (
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Update robot animation
	if m.view.Robot != nil {
		m.view.Robot.Update()
		m.updateRobotState()
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		var cmd tea.Cmd
		m, cmd = m.handleKeyMessages(msg)
		if cmd != nil {
			return m, cmd
		}
	case tokenMsg:
		return m.handleTokenMessages(msg)
	case startTokenStream:
		return m.handleTokenStream(msg)
	case tokenStreamTick:
		return m.handleTokenStreamTick(msg)
	case tokenUsageMsg:
		return m.handleTokenUsageMessage(msg)
	case finalMsg:
		return m.handleFinalMessage(msg)
	case toolUseMsg:
		return m.handleToolUseMessage(msg)
	case actionMsg:
		return m.handleActionMessage(msg)
	case modelMsg:
		return m.handleModelMessage(msg)
	case spinner.TickMsg:
		var spinnerCmds []tea.Cmd
		m, spinnerCmds = m.handleSpinnerTick(msg)
		cmds = append(cmds, spinnerCmds...)
	case progress.FrameMsg:
		// Update all progress bars for token usage
		var progressCmds []tea.Cmd
		for id, info := range m.infos {
			progressModel, cmd := info.TokenProgress.Update(msg)
			info.TokenProgress = progressModel.(progress.Model)
			m.infos[id] = info
			if cmd != nil {
				progressCmds = append(progressCmds, cmd)
			}
		}
		cmds = append(cmds, progressCmds...)
	case activityTickMsg:
		return m.handleActivityTick(msg)
	case refreshMsg:
		// This just causes a re-render to update the footer with live token/cost data
		// Do NOT schedule another refresh - let activityTickMsg handle all timing
	case errMsg:
		return m.handleErrorMessage(msg)
	case agentCompleteMsg:
		return m.handleAgentComplete(msg)
	case agentStartMsg:
		return m.handleAgentStart(msg)
	case thinkingAnimationMsg:
		return m.handleThinkingAnimation(msg)
	case tea.WindowSizeMsg:
		return m.handleWindowResize(msg)
	}

	// Handle viewport scrolling based on active tab
	if m.layout.activeTab == 0 {
		m.view.Chat.Main, _ = m.view.Chat.Main.Update(msg)
		// Only auto-scroll to bottom when new content is being added, not on every update
		// This allows users to scroll through chat history without being forced to bottom
	} else {
		m.view.Chat.Debug, _ = m.view.Chat.Debug.Update(msg)
		// Only auto-scroll to bottom when new content is being added, not on every update
		// This allows users to scroll through debug history without being forced to bottom
	}

	if cmd := m.view.Input.Update(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}
	m.view.Tools, _ = m.view.Tools.Update(msg)
	m.view.Status, _ = m.view.Status.Update(msg)

	return m, tea.Batch(cmds...)
}
