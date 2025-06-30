package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/spinner"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	
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
	} else {
		m.debugVp, _ = m.debugVp.Update(msg)
	}

	m.input, _ = m.input.Update(msg)
	m.tools, _ = m.tools.Update(msg)

	return m, tea.Batch(cmds...)
}
