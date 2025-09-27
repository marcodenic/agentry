package tui

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/google/uuid"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/cost"
	"github.com/marcodenic/agentry/internal/debug"
	"github.com/marcodenic/agentry/internal/glyphs"
	"github.com/marcodenic/agentry/internal/statusbar"
	"github.com/marcodenic/agentry/internal/team"
)

// Model is the root TUI model.
type Model struct {
	agents []*core.Agent
	infos  map[uuid.UUID]*AgentInfo
	order  []uuid.UUID
	active uuid.UUID

	team *team.Team

	vp      viewport.Model
	debugVp viewport.Model // Separate viewport for debug/memory view
	input   inputManager
	tools   list.Model

	cwd string

	activeTab int
	width     int
	height    int
	lastWidth int // Track width changes to avoid expensive reformatting

	// Splash screen state
	showInitialLogo bool

	// Robot companion for Agent 0
	robot *RobotFace

	// Status bar
	statusBarModel statusbar.Model

	pricing *cost.PricingTable

	err error

	keys Keybinds

	// Diagnostics
	diags       []Diag
	diagRunning bool

	// TODO Board
	todoBoard TodoBoard
}

type AgentStatus int

const (
	StatusIdle AgentStatus = iota
	StatusRunning
	StatusError
	StatusStopped
)

// ContentType represents the type of content last added to history
type ContentType int

const (
	ContentTypeEmpty ContentType = iota
	ContentTypeUserInput
	ContentTypeAIResponse
	ContentTypeStatusMessage
	ContentTypeLogo
)

type DebugTraceEvent struct {
	Timestamp time.Time
	Type      string
	Data      map[string]interface{}
	StepNum   int
	Details   string
}

// Diag is a structured diagnostic entry for rendering
type Diag struct {
	File     string
	Line     int
	Col      int
	Code     string
	Severity string
	Message  string
}

type AgentInfo struct {
	Agent               *core.Agent
	History             string
	Status              AgentStatus
	LastContentType     ContentType // Track what type of content was last added
	PendingStatusUpdate string      // Track ongoing status update for progressive completion
	CurrentTool         string
	// TokenCount removed - use Agent.Cost.TotalTokens() for accurate token counts
	TokenHistory        []int
	ActivityData        []float64   // Activity level per second (0.0 to 1.0)
	ActivityTimes       []time.Time // Timestamp for each activity data point
	LastToken           time.Time
	LastActivity        time.Time
	CurrentActivity     int // Tokens processed in current second
	ModelName           string
	Scanner             *bufio.Scanner
	Cancel              context.CancelFunc
	Spinner             spinner.Model
	TokenProgress       progress.Model // Animated progress bar for token usage
	Name                string
	Role                string // Agent role for display (e.g., "System", "Research", "DevOps")
	TokensStarted       bool   // Flag to stop thinking animation when tokens start
	StreamingResponse   string // Current AI response being streamed (unformatted)
	StreamingTokenCount int    // Live token count during streaming (reconciled on completion)

	// Debug and trace fields
	DebugTrace             []DebugTraceEvent // Debug trace events
	CurrentStep            int               // Current step number
	DebugStreamingResponse string            // Debug streaming response
	tracePipeWriter        io.WriteCloser
}

// New creates a new TUI model bound to an Agent.
func New(ag *core.Agent) Model {
	return NewWithConfig(ag, nil, "")
}

// NewWithConfig creates a new TUI model bound to an Agent with optional config.
func NewWithConfig(ag *core.Agent, includePaths []string, configDir string) Model {
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
	inputMgr := newInputManager()
	// Initialize viewport with reasonable default dimensions
	// This prevents text wrapping issues before the first window resize event
	defaultWidth := 90  // 75% of assumed 120 char window width
	defaultHeight := 20 // Reasonable default height
	vp := viewport.New(defaultWidth, defaultHeight)
	debugVp := viewport.New(defaultWidth, defaultHeight)
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

	// Set initial progress bar width (will be updated on first window resize event)
	// Assume a reasonable default window width of 120 characters
	const defaultWindowWidth = 120
	info := newAgentInfo(ag, logoContent, defaultWindowWidth)

	infos := map[uuid.UUID]*AgentInfo{ag.ID: info}

	// Create team context with role loading support
	tm, err := buildTeam(ag, includePaths, configDir)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize team: %v", err))
	}

	// Load base prompt from templates
	ag.Prompt = core.GetDefaultPrompt()
	if strings.TrimSpace(ag.Prompt) == "" {
		debug.Printf("Warning: No default prompt found. Set AGENTRY_DEFAULT_PROMPT or install templates (see docs). Proceeding without a system prompt.")
	}

	// Provide available roles via dedicated <agents> section (do not alter base prompt)
	if ag.Prompt != "" {
		availableRoles := tm.AvailableRoleNames()
		sort.Strings(availableRoles)
		var sb strings.Builder
		sb.WriteString("AVAILABLE AGENTS: You can delegate tasks to these specialized agents using the 'agent' tool:\n\n")
		for _, role := range availableRoles {
			if role == "agent_0" {
				continue
			}
			sb.WriteString(role)
			sb.WriteString("\n")
		}
		sb.WriteString("\nExample delegation: {\"agent\": \"coder\", \"input\": \"create a hello world program\"}")
		if ag.Vars == nil {
			ag.Vars = map[string]string{}
		}
		ag.Vars["AGENTS_SECTION"] = sb.String()
		if os.Getenv("AGENTRY_TUI_MODE") != "1" {
			debug.Printf("🔧 Agent0 agents section populated with %d available roles", len(availableRoles))
		}
	}

	tm.RegisterAgentTool(ag.Tools)

	statusBarModel := newStatusBarModel()

	m := Model{
		agents:          []*core.Agent{ag},
		infos:           infos,
		order:           []uuid.UUID{ag.ID},
		active:          ag.ID,
		team:            tm,
		vp:              vp,
		debugVp:         debugVp,
		input:           inputMgr,
		tools:           l,
		cwd:             cwd,
		keys:            DefaultKeybinds(),
		showInitialLogo: true,
		robot:           NewRobotFace(),
		statusBarModel:  statusBarModel,
		pricing:         cost.NewPricingTable(),
		todoBoard:       NewTodoBoard(),
	}
	return m
}

var sanitizeBlackANSIPattern = regexp.MustCompile(`\x1b\[(?:\d{1,3};)*(?:3[0]|4[0]|38;5;0|48;5;0)(?:;\d{1,3})*m`)

func sanitizeInputANSI(s string) string {
	return sanitizeBlackANSIPattern.ReplaceAllString(s, "")
}

func newTextareaModel() textarea.Model {
	ti := textarea.New()
	if setter, ok := interface{}(&ti).(interface{ SetShowLineNumbers(bool) }); ok {
		setter.SetShowLineNumbers(false)
	} else {
		ti.ShowLineNumbers = false
	}
	ti.Prompt = ""
	ti.Placeholder = "Type your message... (Press Enter to send, Up for previous)"

	noColor := lipgloss.NoColor{}
	clearStyle := lipgloss.NewStyle().Background(noColor)
	textStyle := lipgloss.NewStyle().Background(noColor).Foreground(lipgloss.Color(uiColorForegroundHex))
	placeholderStyle := lipgloss.NewStyle().Background(noColor).Foreground(lipgloss.Color(uiColorPlaceholderHex)).Faint(true)
	numberStyle := lipgloss.NewStyle().Background(noColor).Foreground(lipgloss.Color("#6B7280")).Faint(true)

	ti.BlurredStyle.Base = clearStyle
	ti.FocusedStyle.Base = clearStyle
	ti.BlurredStyle.Text = textStyle
	ti.FocusedStyle.Text = textStyle
	ti.FocusedStyle.CursorLine = textStyle
	ti.BlurredStyle.CursorLine = textStyle
	ti.FocusedStyle.Placeholder = placeholderStyle
	ti.BlurredStyle.Placeholder = placeholderStyle
	ti.FocusedStyle.EndOfBuffer = textStyle
	ti.BlurredStyle.EndOfBuffer = textStyle
	ti.FocusedStyle.LineNumber = numberStyle
	ti.BlurredStyle.LineNumber = numberStyle
	ti.FocusedStyle.CursorLineNumber = numberStyle
	ti.BlurredStyle.CursorLineNumber = numberStyle
	ti.FocusedStyle.Prompt = textStyle
	ti.BlurredStyle.Prompt = textStyle
	ti.Focus()

	return ti
}

func buildTeam(ag *core.Agent, includePaths []string, configDir string) (*team.Team, error) {
	if len(includePaths) > 0 {
		return team.NewTeamWithRoles(ag, 10, "", includePaths, configDir)
	}
	return team.NewTeam(ag, 10, "")
}

func newStatusBarModel() statusbar.Model {
	agentsColors := statusbar.ColorConfig{
		Foreground: lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#ffffff"},
		Background: lipgloss.AdaptiveColor{Light: "#8B5FBF", Dark: "#8B5FBF"},
	}
	cwdColors := statusbar.ColorConfig{
		Foreground: lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#ffffff"},
		Background: lipgloss.AdaptiveColor{Light: "#5B82D7", Dark: "#5B82D7"},
	}
	tokensColors := statusbar.ColorConfig{
		Foreground: lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#ffffff"},
		Background: lipgloss.AdaptiveColor{Light: "#2BA6EF", Dark: "#2BA6EF"},
	}
	costColors := statusbar.ColorConfig{
		Foreground: lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#ffffff"},
		Background: lipgloss.AdaptiveColor{Light: "#00D6EF", Dark: "#00D6EF"},
	}
	return statusbar.New(agentsColors, cwdColors, tokensColors, costColors)
}

func newAgentInfo(ag *core.Agent, history string, defaultWindowWidth int) *AgentInfo {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(uiColorAIAccentHex))

	info := &AgentInfo{
		Agent:               ag,
		Status:              StatusIdle,
		LastContentType:     ContentTypeLogo,
		PendingStatusUpdate: "",
		Spinner:             sp,
		TokenProgress:       createTokenProgressBar(),
		Name:                "Agent 0",
		Role:                "System",
		History:             history,
		ActivityData:        make([]float64, 0),
		ActivityTimes:       make([]time.Time, 0),
		TokenHistory:        []int{},
		DebugTrace:          make([]DebugTraceEvent, 0),
	}

	info.ModelName = ag.ModelName
	if info.ModelName == "" {
		info.ModelName = "unknown"
	}

	panelWidth := int(float64(defaultWindowWidth) * 0.25)
	barWidth := panelWidth - 8
	if barWidth < 10 {
		barWidth = 10
	}
	if barWidth > 50 {
		barWidth = 50
	}
	info.TokenProgress.Width = barWidth

	return info
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
		if info.tracePipeWriter != nil {
			info.tracePipeWriter.Close() // Close trace pipe writers for spawned agents
		}
		if info.Status == StatusRunning {
			info.Status = StatusStopped
			m.infos[id] = info
		}
	}
}

// addContentWithSpacing adds content to agent history with proper spacing based on content type transitions
func (info *AgentInfo) addContentWithSpacing(content string, contentType ContentType) {
	if info.History == "" {
		// First content ever - no spacing needed
		info.History = content
	} else {
		// Determine spacing based on content type transition
		spacing := ""

		switch info.LastContentType {
		case ContentTypeLogo, ContentTypeEmpty:
			// After logo or empty, no spacing needed
			spacing = ""
		case ContentTypeUserInput:
			if contentType == ContentTypeAIResponse {
				// User Input → AI Response: No extra spacing
				spacing = "\n"
			} else {
				// User Input → Status Message: Add spacing
				spacing = "\n\n"
			}
		case ContentTypeAIResponse:
			if contentType == ContentTypeUserInput {
				// AI Response → User Input: Add spacing
				spacing = "\n\n"
			} else {
				// AI Response → Status Message: Add spacing
				spacing = "\n\n"
			}
		case ContentTypeStatusMessage:
			if contentType == ContentTypeStatusMessage {
				// Status Message → Status Message: Group together
				spacing = "\n"
			} else {
				// Status Message → AI Response or User Input: Add spacing
				spacing = "\n\n"
			}
		}

		info.History += spacing + content
	}

	// Update the last content type
	info.LastContentType = contentType
}

// startProgressiveStatusUpdate begins a status update that can be completed later
func (info *AgentInfo) startProgressiveStatusUpdate(content string, m Model) {
	// Check if content already contains glyphs (starts with styled characters)
	// If it does, don't add statusBar prefix, just add proper spacing
	var statusFormatted string
	if strings.Contains(content, "✦") || strings.Contains(content, "▶") || strings.Contains(content, "●") {
		// Content already has glyphs, just add proper spacing alignment
		statusFormatted = "  " + content // Use 4 spaces to align with user input
	} else {
		// Content doesn't have glyphs, use status bar
		statusFormatted = m.statusBar() + "  " + content // Use 4 spaces to align with user input
	}
	info.addContentWithSpacing(statusFormatted, ContentTypeStatusMessage)
	info.PendingStatusUpdate = content // Track the pending update
}

// completeProgressiveStatusUpdate completes a pending status update with a green tick
func (info *AgentInfo) completeProgressiveStatusUpdate(m Model) {
	if info.PendingStatusUpdate == "" {
		return // No pending update to complete
	}

	// Find and replace the last status line in history
	lines := strings.Split(info.History, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		// Check if this line contains our pending status update
		if strings.Contains(line, info.PendingStatusUpdate) {
			// Replace orange bar with green bar and add tick
			updatedLine := strings.Replace(line, m.statusBar(), m.completedStatusBar(), 1)
			updatedLine += " " + glyphs.GreenCheckmark()
			lines[i] = updatedLine
			break
		}
	}

	info.History = strings.Join(lines, "\n")
	info.PendingStatusUpdate = "" // Clear pending update
}
