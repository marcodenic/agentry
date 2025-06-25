package tui

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"io"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/marcodenic/agentry/internal/converse"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/trace"
)

// Model is the root TUI model.
type Model struct {
	agents []*core.Agent
	infos  map[uuid.UUID]*AgentInfo
	order  []uuid.UUID
	active uuid.UUID

	team *converse.Team

	vp    viewport.Model
	input textinput.Model
	tools list.Model

	cwd string

	activeTab int
	width     int
	height    int

	err error

	theme Theme
	keys  Keybinds
}

type AgentStatus int

const (
	StatusIdle AgentStatus = iota
	StatusRunning
	StatusError
	StatusStopped
)

type AgentInfo struct {
	Agent           *core.Agent
	History         string
	Status          AgentStatus
	CurrentTool     string
	TokenCount      int
	TokenHistory    []int
	ActivityData    []float64   // Activity level per second (0.0 to 1.0)
	ActivityTimes   []time.Time // Timestamp for each activity data point
	LastToken       time.Time
	LastActivity    time.Time
	CurrentActivity int // Tokens processed in current second
	ModelName       string
	Scanner         *bufio.Scanner
	Cancel          context.CancelFunc
	Spinner         spinner.Model
	Name            string
	Role            string // Agent role for display (e.g., "System", "Research", "DevOps")
}

// New creates a new TUI model bound to an Agent.
func New(ag *core.Agent) Model {
	th := LoadTheme()
	items := []list.Item{}
	for name, tl := range ag.Tools {
		items = append(items, listItem{name: name, desc: tl.Description()})
	}
	l := list.New(items, listItemDelegate{}, 0, 0)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Title = "Tools"
	l.SetShowHelp(false) // Hide help bar with navigation hints
	// Assign custom keymap to disable navigation keys
	l.KeyMap.CursorUp = NoNavKeyMap.CursorUp
	l.KeyMap.CursorDown = NoNavKeyMap.CursorDown
	l.KeyMap.PrevPage = NoNavKeyMap.PrevPage
	l.KeyMap.NextPage = NoNavKeyMap.NextPage
	l.KeyMap.GoToStart = NoNavKeyMap.GoToStart
	l.KeyMap.GoToEnd = NoNavKeyMap.GoToEnd
	l.KeyMap.Filter = NoNavKeyMap.Filter
	l.KeyMap.ClearFilter = NoNavKeyMap.ClearFilter
	l.KeyMap.CancelWhileFiltering = NoNavKeyMap.CancelWhileFiltering
	l.KeyMap.AcceptWhileFiltering = NoNavKeyMap.AcceptWhileFiltering
	l.KeyMap.ShowFullHelp = NoNavKeyMap.ShowFullHelp
	l.KeyMap.CloseFullHelp = NoNavKeyMap.CloseFullHelp
	l.KeyMap.Quit = NoNavKeyMap.Quit
	l.KeyMap.ForceQuit = NoNavKeyMap.ForceQuit
	ti := textinput.New()
	ti.Placeholder = "Type your message... (Press Enter to send)"
	ti.CharLimit = 2000 // Allow longer messages
	ti.Focus()

	vp := viewport.New(0, 0)
	cwd, _ := os.Getwd()
	info := &AgentInfo{
		Agent:   ag,
		Status:  StatusIdle,
		Spinner: spinner.New(),
		Name:    "master",
		Role:    "System", ActivityData: make([]float64, 0),
		ActivityTimes:   make([]time.Time, 0),
		CurrentActivity: 0,
		LastActivity:    time.Time{}, // Start with zero time so first tick will initialize properly
		// Initialize with empty activity for real-time chart
		TokenHistory: []int{},
	}
	infos := map[uuid.UUID]*AgentInfo{ag.ID: info}

	tm, err := converse.NewTeam(ag, 1, "")
	if err != nil {
		tm = &converse.Team{}
	}

	m := Model{
		agents: []*core.Agent{ag},
		infos:  infos,
		order:  []uuid.UUID{ag.ID},
		active: ag.ID,
		team:   tm,
		vp:     vp,
		input:  ti,
		tools:  l,
		cwd:    cwd,
		theme:  th,
		keys:   th.Keybinds,
	}
	return m
}

type listItem struct{ name, desc string }

func (l listItem) Title() string       { return l.name }
func (l listItem) Description() string { return l.desc }
func (l listItem) FilterValue() string { return l.name }

type listItemDelegate struct{}

func (d listItemDelegate) Height() int                               { return 1 }
func (d listItemDelegate) Spacing() int                              { return 0 }
func (d listItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d listItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	it := item.(listItem)
	style := lipgloss.NewStyle()
	if index == m.Index() {
		style = style.Bold(true)
		io.WriteString(w, style.Render("> "+it.name))
	} else {
		io.WriteString(w, "  "+it.name)
	}
}

type tokenMsg struct {
	id    uuid.UUID
	token string
}

type toolUseMsg struct {
	id   uuid.UUID
	name string
}

type modelMsg struct {
	id   uuid.UUID
	name string
}

type activityTickMsg struct{}

type errMsg struct{ error }

type finalMsg struct {
	id   uuid.UUID
	text string
}

func streamTokens(id uuid.UUID, out string) tea.Cmd {
	runes := []rune(out)
	cmds := make([]tea.Cmd, len(runes))
	for i, r := range runes {
		tok := string(r)
		delay := time.Duration(i*30) * time.Millisecond
		cmds[i] = tea.Tick(delay, func(t time.Time) tea.Msg { return tokenMsg{id: id, token: tok} })
	}
	return tea.Batch(cmds...)
}

func (m *Model) readEvent(id uuid.UUID) tea.Msg {
	info := m.infos[id]
	if info == nil || info.Scanner == nil {
		return nil
	}
	for {
		if !info.Scanner.Scan() {
			if err := info.Scanner.Err(); err != nil {
				return errMsg{err}
			}
			return nil
		}
		var ev trace.Event
		if err := json.Unmarshal(info.Scanner.Bytes(), &ev); err != nil {
			return errMsg{err}
		}
		switch ev.Type {
		case trace.EventFinal:
			if s, ok := ev.Data.(string); ok {
				return finalMsg{id: id, text: s}
			}
		case trace.EventModelStart:
			if name, ok := ev.Data.(string); ok {
				return modelMsg{id: id, name: name}
			}
		case trace.EventToolEnd:
			if m2, ok := ev.Data.(map[string]any); ok {
				if name, ok := m2["name"].(string); ok {
					return toolUseMsg{id: id, name: name}
				}
			}
		default:
			continue
		}
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

func (m Model) Init() tea.Cmd {
	// Start the activity chart ticker (update every second)
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return activityTickMsg{}
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Enhanced agent navigation
		if key.Matches(msg, PrevAgentKey) {
			m = m.cycleActive(-1)
		} else if key.Matches(msg, NextAgentKey) {
			m = m.cycleActive(1)
		} else if key.Matches(msg, FirstAgentKey) {
			m = m.jumpToAgent(0)
		} else if key.Matches(msg, LastAgentKey) {
			m = m.jumpToAgent(len(m.order) - 1)
		}
		switch msg.String() {
		case m.keys.Quit:
			return m, tea.Quit
		case m.keys.ToggleTab:
			m.activeTab = 1 - m.activeTab
		case m.keys.Submit:
			if m.input.Focused() {
				txt := m.input.Value()
				m.input.SetValue("")
				if strings.HasPrefix(txt, "/") {
					var cmd tea.Cmd
					m, cmd = m.handleCommand(txt)
					return m, cmd
				}
				return m.startAgent(m.active, txt)
			}
		}
	case tokenMsg:
		info := m.infos[msg.id]
		info.History += msg.token
		info.TokenCount++
		info.CurrentActivity++ // Just increment counter, let activityTickMsg handle data points

		now := time.Now()

		// Legacy token history update (keep for compatibility)
		if info.LastToken.IsZero() || now.Sub(info.LastToken) > time.Second {
			info.TokenHistory = append(info.TokenHistory, 1)
			if len(info.TokenHistory) > 20 {
				info.TokenHistory = info.TokenHistory[1:]
			}
		} else if len(info.TokenHistory) > 0 {
			info.TokenHistory[len(info.TokenHistory)-1]++
		}
		info.LastToken = now
		m.infos[msg.id] = info // IMPORTANT: Save the updated info back to the map
		if msg.id == m.active {
			base := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.Palette.Foreground)).Background(lipgloss.Color(m.theme.Palette.Background))
			m.vp.SetContent(base.Copy().Width(m.vp.Width).Render(info.History))
			m.vp.GotoBottom()
		}
	case finalMsg:
		info := m.infos[msg.id]
		info.History += m.aiBar() + " "
		if msg.id == m.active {
			base := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.Palette.Foreground)).Background(lipgloss.Color(m.theme.Palette.Background))
			m.vp.SetContent(base.Copy().Width(m.vp.Width).Render(info.History))
			m.vp.GotoBottom()
		}
		info.Status = StatusIdle
		m.infos[msg.id] = info
		return m, tea.Batch(streamTokens(msg.id, msg.text+"\n"), m.readCmd(msg.id))
	case toolUseMsg:
		info := m.infos[msg.id]
		info.CurrentTool = msg.name
		m.infos[msg.id] = info
		return m, m.readCmd(msg.id)
	case modelMsg:
		info := m.infos[msg.id]
		info.ModelName = msg.name
		m.infos[msg.id] = info
		return m, m.readCmd(msg.id)
	case spinner.TickMsg:
		for id, ag := range m.infos {
			if ag.Status == StatusRunning {
				var c tea.Cmd
				ag.Spinner, c = ag.Spinner.Update(msg)
				cmds = append(cmds, c)
				m.infos[id] = ag
			}
		}
	case activityTickMsg:
		// Update activity data for all agents (to make chart scroll even when idle)
		now := time.Now()
		for id, info := range m.infos {
			// Always add a data point every second to make the chart scroll
			// If there was activity in this second, it will already be recorded
			// Otherwise, add a zero point to show time progression

			shouldAddDataPoint := false

			if len(info.ActivityTimes) == 0 {
				// First data point
				shouldAddDataPoint = true
			} else {
				// Check if we need a new data point (more than 1 second since last)
				lastTime := info.ActivityTimes[len(info.ActivityTimes)-1]
				if now.Sub(lastTime) >= time.Second {
					shouldAddDataPoint = true
				}
			}

			if shouldAddDataPoint {
				// Add activity level (either current activity or 0.0 for idle)
				activityLevel := 0.0
				if info.CurrentActivity > 0 {
					// Normalize current activity (10 tokens/sec = 100%)
					activityLevel = float64(info.CurrentActivity) / 10.0
					if activityLevel > 1.0 {
						activityLevel = 1.0
					}
					info.CurrentActivity = 0 // Reset for next second
				}

				info.ActivityData = append(info.ActivityData, activityLevel)
				info.ActivityTimes = append(info.ActivityTimes, now)

				// Keep only last 60 seconds of data
				cutoffTime := now.Add(-60 * time.Second)
				var newData []float64
				var newTimes []time.Time

				for i, t := range info.ActivityTimes {
					if t.After(cutoffTime) {
						newData = append(newData, info.ActivityData[i])
						newTimes = append(newTimes, info.ActivityTimes[i])
					}
				}

				info.ActivityData = newData
				info.ActivityTimes = newTimes
				info.LastActivity = now
				m.infos[id] = info
			}
		}
		// Schedule next tick
		cmds = append(cmds, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return activityTickMsg{}
		}))
	case errMsg:
		m.err = msg
		if info, ok := m.infos[m.active]; ok {
			info.Status = StatusError
			m.infos[m.active] = info
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.input.Width = int(float64(msg.Width)*0.75) - 2
		m.vp.Width = int(float64(msg.Width)*0.75) - 2 // Calculate viewport height: total height - input line - footer line
		m.vp.Height = msg.Height - 2
		m.tools.SetSize(int(float64(msg.Width)*0.25)-2, msg.Height-2)
		if info, ok := m.infos[m.active]; ok {
			base := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.Palette.Foreground)).Background(lipgloss.Color(m.theme.Palette.Background))
			m.vp.SetContent(base.Copy().Width(m.vp.Width).Render(info.History))
		}
	}

	m.input, _ = m.input.Update(msg)
	m.tools, _ = m.tools.Update(msg)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	var leftContent string
	if m.activeTab == 0 {
		leftContent = m.vp.View() + "\n" + m.input.View()
	} else {
		if info, ok := m.infos[m.active]; ok {
			leftContent = renderMemory(info.Agent)
		}
	}
	if m.err != nil {
		leftContent += "\nERR: " + m.err.Error()
	}
	base := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.Palette.Foreground)).
		Background(lipgloss.Color(m.theme.Palette.Background))
	left := base.Copy().Width(int(float64(m.width) * 0.75)).Render(leftContent)
	right := base.Copy().Width(int(float64(m.width) * 0.25)).Render(m.agentPanel())
	main := lipgloss.JoinHorizontal(lipgloss.Top, left, right)

	tokens := 0
	costVal := 0.0
	if info, ok := m.infos[m.active]; ok && info.Agent.Cost != nil {
		tokens = info.Agent.Cost.TotalTokens()
		costVal = info.Agent.Cost.TotalCost()
	}
	footer := fmt.Sprintf("cwd: %s | agents: %d | tokens: %d cost: $%.4f", m.cwd, len(m.infos), tokens, costVal)
	footer = base.Copy().Width(m.width).Render(footer)
	return lipgloss.JoinVertical(lipgloss.Left, main, footer)
}
