package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/marcodenic/agentry/internal/converse"
	"github.com/marcodenic/agentry/internal/core"
)

// DEPRECATED: ChatModel is a legacy unified model that manages a converse.Team.
//
// This model is DEPRECATED and will be removed in a future version.
// Use Model (tui.New) instead for the unified interface that handles
// both single agents and teams through the agent panel and spawn commands.
//
// The unified Model provides:
// - Consistent btop-style agent panel for all scenarios
// - Advanced real-time status indicators and token tracking
// - Streamlined command interface (/spawn, /switch, /stop, /converse)
// - Better performance and code maintainability
// - Full-screen optimized layout with no empty space
//
// Migration: Replace calls to NewChat() with tui.New() and use /spawn commands
// to create additional agents as needed.
type ChatModel struct {
	parent *core.Agent
	team   *converse.Team
	infos  map[uuid.UUID]*AgentInfo
	names  []string
	active uuid.UUID

	vps    []viewport.Model
	input  textinput.Model
	width  int
	height int

	theme Theme
	keys  Keybinds
	err   error
}

// NewChat creates a team with n agents talking about topic.
//
// DEPRECATED: This function is deprecated and will be removed in a future version.
// Use tui.New() instead for the unified interface. The unified Model supports
// spawning multiple agents via /spawn commands and provides a consistent,
// advanced agent panel for all scenarios.
//
// Migration example:
//
//	Old: model, err := tui.NewChat(agent, 3, "topic")
//	New: model := tui.New(agent)
//	     // Then use /spawn commands to create additional agents
//	     // Use /converse command for team conversations
func NewChat(parent *core.Agent, n int, topic string) (ChatModel, error) {
	th := LoadTheme()
	t, err := converse.NewTeam(parent, n, topic)
	if err != nil {
		return ChatModel{}, err
	}
	vps := make([]viewport.Model, n)
	infos := make(map[uuid.UUID]*AgentInfo, n)
	for i, ag := range t.Agents() {
		vps[i] = viewport.New(0, 0)
		infos[ag.ID] = &AgentInfo{Agent: ag, Status: StatusIdle, Name: t.Names()[i]}
	}
	ti := textinput.New()
	ti.Placeholder = "Message (Shift+Enter for new line)"
	ti.CharLimit = 1000 // Set reasonable character limit
	ti.Focus()
	return ChatModel{parent: parent, team: t, infos: infos, names: t.Names(), active: t.Agents()[0].ID,
		vps: vps, input: ti, theme: th, keys: th.Keybinds}, nil
}

func (m ChatModel) Init() tea.Cmd { return nil }

// Update handles Bubble Tea messages.
func (m ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case m.keys.Quit:
			return m, tea.Quit
		case m.keys.Submit:
			if m.input.Focused() {
				txt := m.input.Value()
				m.input.SetValue("")
				if strings.HasPrefix(txt, "/") {
					var cmd tea.Cmd
					m, cmd = m.handleCommand(txt)
					return m, cmd
				}
				return m.callActive(txt)
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Set input width to match viewport width
		inputWidth := int(float64(msg.Width)*0.75) - 2
		m.input.Width = inputWidth
		for i := range m.vps {
			m.vps[i].Width = inputWidth
			m.vps[i].Height = msg.Height - 5
			if info, ok := m.infos[m.team.Agents()[i].ID]; ok {
				// Apply text wrapping when setting content
				wrappedContent := lipgloss.NewStyle().Width(m.vps[i].Width).Render(info.History)
				m.vps[i].SetContent(wrappedContent)
			}
		}
	}

	m.input, _ = m.input.Update(msg)
	return m, nil
}

func (m ChatModel) View() string {
	vp := viewport.Model{}
	if info, ok := m.infos[m.active]; ok {
		vp = m.vps[m.indexOf(m.active)]
		// Apply text wrapping when setting content
		wrappedContent := lipgloss.NewStyle().Width(vp.Width).Render(info.History)
		vp.SetContent(wrappedContent)
	}

	// Create left panel (chat + input)
	leftContent := vp.View() + "\n" + m.input.View()

	// Create right panel (agent sidebar)
	rightContent := m.agentPanel()

	// Layout
	base := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.Palette.Foreground)).
		Background(lipgloss.Color(m.theme.Palette.Background))

	left := base.Copy().Width(int(float64(m.width) * 0.75)).Render(leftContent)
	right := base.Copy().Width(int(float64(m.width) * 0.25)).Render(rightContent)
	main := lipgloss.JoinHorizontal(lipgloss.Top, left, right)

	// Footer
	tokens := 0
	costVal := 0.0
	if m.parent != nil && m.parent.Cost != nil {
		tokens = m.parent.Cost.TotalTokens()
		costVal = m.parent.Cost.TotalCost()
	}
	footer := fmt.Sprintf("agents: %d | tokens: %d cost: $%.4f", len(m.infos), tokens, costVal)
	footer = base.Copy().Width(m.width).Render(footer)

	return lipgloss.JoinVertical(lipgloss.Left, main, footer)
}
