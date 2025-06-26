package tui

import (
	"bufio"
	"context"
	"io"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/google/uuid"
	"github.com/marcodenic/agentry/internal/converse"
	"github.com/marcodenic/agentry/internal/core"
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
	TokensStarted   bool   // Flag to stop thinking animation when tokens start
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
		Name:    "Agent 0",
		Role:    "System", ActivityData: make([]float64, 0),
		ActivityTimes:   make([]time.Time, 0),
		CurrentActivity: 0,
		LastActivity:    time.Time{}, // Start with zero time so first tick will initialize properly
		// Initialize with empty activity for real-time chart
		TokenHistory: []int{},
		TokensStarted: false,
	}
	infos := map[uuid.UUID]*AgentInfo{ag.ID: info}

	// Create team context without pre-spawning agents
	tm, err := converse.NewTeamContext(ag)
	if err != nil {
		panic(err) // For now, panic on error - TODO: handle gracefully
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
