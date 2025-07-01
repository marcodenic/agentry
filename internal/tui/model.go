package tui

import (
	"bufio"
	"context"
	"io"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/google/uuid"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/team"
)

// applyGradientToLogo applies a beautiful gradient effect to the ASCII logo
func applyGradientToLogo(logo string) string {
	lines := strings.Split(logo, "\n")
	var styledLines []string

	// Vibrant gradient: Magenta → Purple → Blue → Cyan (matching the stunning visuals)
	colors := []string{
		"#FF44FF", // Bright neon magenta
		"#F542F5", // Magenta
		"#EB40EB", // Pink-magenta
		"#E13EE1", // Purple-pink
		"#D73CD7", // Purple-magenta
		"#CD3ACD", // Purple
		"#C338C3", // Deep purple
		"#B936B9", // Purple-blue
		"#AF34AF", // Blue-purple
		"#A532A5", // Purple-blue
		"#9B309B", // Blue-purple
		"#912E91", // Blue
		"#872C87", // Deep blue-purple
		"#7D2A7D", // Blue
		"#732873", // Blue-cyan
		"#44AAFF", // Bright cyan-blue
	}

	totalLines := len(lines)

	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			styledLines = append(styledLines, line)
			continue
		}

		// Calculate which color to use based on line position
		colorIndex := (i * len(colors)) / totalLines
		if colorIndex >= len(colors) {
			colorIndex = len(colors) - 1
		}

		// Apply the color to the line with subtle styling
		style := lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors[colorIndex]))

		styledLines = append(styledLines, style.Render(line))
	}

	return strings.Join(styledLines, "\n")
}

// Model is the root TUI model.
type Model struct {
	agents []*core.Agent
	infos  map[uuid.UUID]*AgentInfo
	order  []uuid.UUID
	active uuid.UUID

	team *team.Team

	vp    viewport.Model
	debugVp viewport.Model // Separate viewport for debug/memory view
	input textinput.Model
	tools list.Model

	cwd string

	activeTab int
	width     int
	height    int
	lastWidth int // Track width changes to avoid expensive reformatting

	// Splash screen state
	showInitialLogo bool

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

type DebugTraceEvent struct {
	Timestamp time.Time
	Type      string
	Data      map[string]interface{}
	StepNum   int
	Details   string
}

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
	StreamingResponse string // Current AI response being streamed (unformatted)
	DebugTrace      []DebugTraceEvent // Detailed trace history for debug view
	CurrentStep     int              // Current step number for trace events
	DebugStreamingResponse string    // Separate streaming response for debug view
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
	debugVp := viewport.New(0, 0)
	cwd, _ := os.Getwd()

	// Initialize with ASCII logo as welcome content
	rawLogoContent := `
                                                             
                                                             
                  ████▒               ▒████                  
                    ▒▓███▓▒       ▒▓███▓▒                    
                      ▒█▒████▓▒▓████▓█▒                      
                      ▒█   ▓█████▓▒  █▒                      
                      ▒█▓███▓▓█▓▓███▓█▒                      
                   ▒▓███▓▒   ▒▓▒   ▒▓███▓▒                   
                 ▒███▓▓█     ▒▓▒     █▓▓▓██▒                 
                      ▒█     ▒▓▒     █▒                      
                      ▒█     ▒▓▒     █▒                      
                      ▒█     ▒▓▒     █▒                      
                      ▒█     ▒▓▒     █▒                      
                      ▒█     ▒▓▒     █▒                      
                      ▒█     ▒▓▒     █▒                      
                             ▒▓▒                             
                                                             
                                      v0.2.0                 
                 █▀█ █▀▀ █▀▀ █▀█ ▀█▀ █▀▄ █ █                 
                 █▀█ █ █ █▀▀ █ █  █  █▀▄  █                  
                 ▀ ▀ ▀▀▀ ▀▀▀ ▀ ▀  ▀  ▀ ▀  ▀                  
               AGENT  ORCHESTRATION  FRAMEWORK               
                                                             `
	
	// Apply beautiful gradient coloring to the logo
	logoContent := applyGradientToLogo(rawLogoContent)

	info := &AgentInfo{
		Agent:   ag,
		Status:  StatusIdle,
		Spinner: spinner.New(),
		Name:    "Agent 0",
		Role:    "System", 
		History: logoContent,
		ActivityData: make([]float64, 0),
		ActivityTimes:   make([]time.Time, 0),
		CurrentActivity: 0,
		LastActivity:    time.Time{}, // Start with zero time so first tick will initialize properly
		// Initialize with empty activity for real-time chart
		TokenHistory: []int{},
		TokensStarted: false,
		StreamingResponse: "",
		DebugTrace: make([]DebugTraceEvent, 0), // Initialize debug trace
		CurrentStep: 0,
		DebugStreamingResponse: "", // Initialize debug streaming response
	}
	infos := map[uuid.UUID]*AgentInfo{ag.ID: info}

	// Create team context without pre-spawning agents
	tm, err := team.NewTeam(ag, 10, "")
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
		debugVp: debugVp,
		input:  ti,
		tools:  l,
		cwd:    cwd,
		theme:  th,
		keys:   th.Keybinds,
		showInitialLogo: true,
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

// Cleanup cancels all running agents and performs necessary cleanup.
// This should be called when the application is shutting down.
func (m *Model) Cleanup() {
	for id, info := range m.infos {
		if info.Cancel != nil {
			info.Cancel() // Cancel all running agent contexts
		}
		if info.Status == StatusRunning {
			info.Status = StatusStopped
			m.infos[id] = info
		}
	}
}
