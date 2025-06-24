package tui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/marcodenic/agentry/internal/converse"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/team"
)

const maxTokens = 1000

type teamMsg struct {
	idx  int
	text string
}

type startMsg struct{ idx int }

// TeamModel displays a multi-agent conversation.
type TeamModel struct {
	team     *converse.Team
	vps      []viewport.Model
	spinners []spinner.Model
	bars     []progress.Model
	tokens   []int
	running  []bool
	turn     int
	history  []string
	statuses []AgentStatus
	roles    []string
	focus    int
	paused   bool
	width    int
	height   int
	theme    Theme
	keys     Keybinds
	err      error
}

// NewTeam creates a TeamModel with n agents talking about topic.
func NewTeam(parent *core.Agent, n int, topic string) (TeamModel, error) {
	th := LoadTheme()
	t, err := converse.NewTeam(parent, n, topic)
	if err != nil {
		return TeamModel{}, err
	}
	vps := make([]viewport.Model, n)
	sp := make([]spinner.Model, n)
	bars := make([]progress.Model, n)
	hist := make([]string, n)
	tokens := make([]int, n)
	running := make([]bool, n)
	status := make([]AgentStatus, n)
	roles := t.Names()
	for i := range vps {
		vps[i] = viewport.New(0, 0)
		sp[i] = spinner.New(spinner.WithSpinner(spinner.Line))
		sp[i].Style = lipgloss.NewStyle().Foreground(lipgloss.Color(th.AIBarColor))
		bars[i] = progress.New(progress.WithDefaultGradient())
	}
	return TeamModel{team: t, vps: vps, spinners: sp, bars: bars, history: hist, tokens: tokens, running: running, statuses: status, roles: roles, theme: th, keys: th.Keybinds}, nil
}

func (m TeamModel) Init() tea.Cmd {
	return m.stepCmd()
}

func (m TeamModel) stepCmd() tea.Cmd {
	idx := m.turn % len(m.vps)
	m.turn++
	return tea.Batch(func() tea.Msg { return startMsg{idx: idx} }, func() tea.Msg {
		ctx := team.WithContext(context.Background(), m.team)
		i, out, err := m.team.Step(ctx)
		if err != nil {
			return errMsg{err}
		}
		return teamMsg{idx: i, text: out}
	})
}

func (m TeamModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
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
			m.bars[i].Width = paneWidth - 4
		}
	case startMsg:
		m.running[msg.idx] = true
		m.statuses[msg.idx] = StatusRunning
		cmds = append(cmds, m.spinners[msg.idx].Tick)
	case spinner.TickMsg:
		for i := range m.spinners {
			var c tea.Cmd
			m.spinners[i], c = m.spinners[i].Update(msg)
			if m.running[i] {
				cmds = append(cmds, c)
			}		}
	case progress.FrameMsg:
		for i := range m.bars {
			var c tea.Cmd
			var newModel tea.Model
			newModel, c = m.bars[i].Update(msg)
			m.bars[i] = newModel.(progress.Model)
			if m.running[i] {
				cmds = append(cmds, c)
			}
		}
	case teamMsg:
		m.history[msg.idx] += msg.text + "\n"
		m.vps[msg.idx].SetContent(lipgloss.NewStyle().Width(m.vps[msg.idx].Width).Render(m.history[msg.idx]))
		m.vps[msg.idx].GotoBottom()
		m.running[msg.idx] = false
		m.tokens[msg.idx] += len([]rune(msg.text))
		m.statuses[msg.idx] = StatusIdle
		cmds = append(cmds, m.bars[msg.idx].SetPercent(float64(m.tokens[msg.idx])/1000))
		if !m.paused {
			cmds = append(cmds, m.stepCmd())
		}
	case errMsg:
		m.err = msg
	}
	return m, tea.Batch(cmds...)
}

func (m TeamModel) agentPanel() string {
	lines := make([]string, len(m.roles))
	for i, name := range m.roles {
		status := map[AgentStatus]string{
			StatusIdle:    "idle",
			StatusRunning: "run",
			StatusError:   "error",
			StatusStopped: "stopped",
		}[m.statuses[i]]
		lines[i] = fmt.Sprintf("%s [%s]", name, status)
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m TeamModel) View() string {
	cols := make([]string, len(m.vps))
	paneWidth := int(float64(m.width) / float64(len(m.vps)))
	for i, vp := range m.vps {
		style := lipgloss.NewStyle().Width(paneWidth)
		if i == m.focus {
			style = style.Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color(m.theme.UserBarColor))
		}
		bottom := m.bars[i].View() + " " + m.spinners[i].View()
		content := lipgloss.JoinVertical(lipgloss.Left, vp.View(), bottom)
		cols[i] = style.Render(content)
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, cols...)
	panel := m.agentPanel()
	footer := fmt.Sprintf("focus: %d | %s to pause", m.focus+1, m.keys.Pause)
	if m.err != nil {
		footer += " | ERR: " + m.err.Error()
	}
	footer = lipgloss.NewStyle().Width(m.width).Render(footer)
	return lipgloss.JoinVertical(lipgloss.Left, row, panel, footer)
}

var _ tea.Model = TeamModel{}
