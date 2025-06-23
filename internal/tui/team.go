package tui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/marcodenic/agentry/internal/converse"
	"github.com/marcodenic/agentry/internal/core"
)

type teamMsg struct {
	idx  int
	text string
}

// TeamModel displays a multi-agent conversation.
type TeamModel struct {
	team    *converse.Team
	vps     []viewport.Model
	history []string
	focus   int
	paused  bool
	width   int
	height  int
	theme   Theme
	keys    Keybinds
	err     error
}

// NewTeam creates a TeamModel with n agents talking about topic.
func NewTeam(parent *core.Agent, n int, topic string) (TeamModel, error) {
	th := LoadTheme()
	t, err := converse.NewTeam(parent, n, topic)
	if err != nil {
		return TeamModel{}, err
	}
	vps := make([]viewport.Model, n)
	hist := make([]string, n)
	for i := range vps {
		vps[i] = viewport.New(0, 0)
	}
	return TeamModel{team: t, vps: vps, history: hist, theme: th, keys: th.Keybinds}, nil
}

func (m TeamModel) Init() tea.Cmd {
	return m.stepCmd()
}

func (m TeamModel) stepCmd() tea.Cmd {
	return func() tea.Msg {
		idx, out, err := m.team.Step(context.Background())
		if err != nil {
			return errMsg{err}
		}
		return teamMsg{idx: idx, text: out}
	}
}

func (m TeamModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case m.keys.Quit:
			return m, tea.Quit
		case m.keys.NextPane:
			m.focus = (m.focus + 1) % len(m.vps)
		case m.keys.PrevPane:
			m.focus = (m.focus - 1 + len(m.vps)) % len(m.vps)
		case m.keys.Pause:
			m.paused = !m.paused
			if !m.paused {
				return m, m.stepCmd()
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		paneWidth := msg.Width / len(m.vps)
		for i := range m.vps {
			m.vps[i].Width = paneWidth - 2
			m.vps[i].Height = msg.Height - 2
			m.vps[i].SetContent(lipgloss.NewStyle().Width(m.vps[i].Width).Render(m.history[i]))
		}
	case teamMsg:
		m.history[msg.idx] += msg.text + "\n"
		m.vps[msg.idx].SetContent(lipgloss.NewStyle().Width(m.vps[msg.idx].Width).Render(m.history[msg.idx]))
		m.vps[msg.idx].GotoBottom()
		if !m.paused {
			return m, m.stepCmd()
		}
	case errMsg:
		m.err = msg
	}
	return m, nil
}

func (m TeamModel) View() string {
	cols := make([]string, len(m.vps))
	paneWidth := int(float64(m.width) / float64(len(m.vps)))
	for i, vp := range m.vps {
		style := lipgloss.NewStyle().Width(paneWidth)
		if i == m.focus {
			style = style.Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color(m.theme.UserBarColor))
		}
		cols[i] = style.Render(vp.View())
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, cols...)
	footer := fmt.Sprintf("focus: %d | %s to pause", m.focus+1, m.keys.Pause)
	if m.err != nil {
		footer += " | ERR: " + m.err.Error()
	}
	footer = lipgloss.NewStyle().Width(m.width).Render(footer)
	return lipgloss.JoinVertical(lipgloss.Left, row, footer)
}

var _ tea.Model = TeamModel{}
