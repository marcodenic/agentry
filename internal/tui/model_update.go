package tui

import (
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Update robot animation
	if m.robot != nil {
		m.robot.Update()
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
	case thinkingAnimationMsg:
		return m.handleThinkingAnimation(msg)
	case tea.WindowSizeMsg:
		return m.handleWindowResize(msg)
	}

	// Handle viewport scrolling based on active tab
	if m.activeTab == 0 {
		m.vp, _ = m.vp.Update(msg)
		// Only auto-scroll to bottom when new content is being added, not on every update
		// This allows users to scroll through chat history without being forced to bottom
	} else {
		m.debugVp, _ = m.debugVp.Update(msg)
		// Only auto-scroll to bottom when new content is being added, not on every update
		// This allows users to scroll through debug history without being forced to bottom
	}

	m.input, _ = m.input.Update(msg)
	m.tools, _ = m.tools.Update(msg)
	m.statusBarModel, _ = m.statusBarModel.Update(msg)

	return m, tea.Batch(cmds...)
}
