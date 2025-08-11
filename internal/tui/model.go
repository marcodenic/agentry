package tui

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
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

// applyGradientToLogo applies a beautiful gradient effect to the ASCII logo
func applyGradientToLogo(logo string) string {
	lines := strings.Split(logo, "\n")
	var styledLines []string

	// Define gradient colors - subtle purple to blue to teal (matching the image style)
	colors := []string{
		"#8B5FBF", // Soft purple
		"#8B5FBF", // Purple-blue
		"#6B76CF", // Lavender blue
		"#5B82D7", // Medium blue
		"#4B8EDF", // Light blue
		"#3B9AE7", // Sky blue
		"#2BA6EF", // Bright blue
		"#1BB2F7", // Cyan blue
		"#0BBEFF", // Light cyan
		"#00CAF7", // Teal cyan
		"#00D6EF", // Soft teal
		"#00E2E7", // Light teal
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

	vp      viewport.Model
	debugVp viewport.Model // Separate viewport for debug/memory view
	input   textarea.Model
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

	theme Theme
	keys  Keybinds

	// Diagnostics
	diags       []Diag
	diagRunning bool

	// Dynamic input sizing and history
	inputHeight  int
	inputHistory []string
	historyIndex int // -1 when not navigating history
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
	ti := textarea.New()
	// Hide the default line number gutter (avoids the leading "1 │")
	// Newer bubbles exposes SetShowLineNumbers; fall back to field if needed.
	// Use both patterns to be safe across versions.
	if setter, ok := interface{}(&ti).(interface{ SetShowLineNumbers(bool) }); ok {
		setter.SetShowLineNumbers(false)
	} else {
		ti.ShowLineNumbers = false
	}
	// Remove the default prompt (which renders a vertical bar) and any padding
	ti.Prompt = ""
	ti.Placeholder = "Type your message... (Press Enter to send)"
	// Prevent left padding from shifting the first row
	if styler, ok := interface{}(&ti).(interface{ SetBaseStyle(lipgloss.Style) }); ok {
		styler.SetBaseStyle(lipgloss.NewStyle().Padding(0))
	}
	ti.Focus()
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

	info := &AgentInfo{
		Agent:               ag,
		Status:              StatusIdle,
		LastContentType:     ContentTypeLogo, // Start with logo content
		PendingStatusUpdate: "",              // No pending status update initially
		Spinner:             spinner.New(),
		TokenProgress:       createTokenProgressBar(),
		Name:                "Agent 0",
		Role:                "System",
		History:             logoContent,
		ActivityData:        make([]float64, 0),
		ActivityTimes:       make([]time.Time, 0),
		CurrentActivity:     0,
		LastActivity:        time.Time{}, // Start with zero time so first tick will initialize properly
		// Initialize with empty activity for real-time chart
		TokenHistory:           []int{},
		TokensStarted:          false,
		StreamingResponse:      "",
		StreamingTokenCount:    0,                          // Initialize live token count
		DebugTrace:             make([]DebugTraceEvent, 0), // Initialize debug trace
		CurrentStep:            0,
		DebugStreamingResponse: "", // Initialize debug streaming response
	}

	// Get the model name from Agent 0
	info.ModelName = ag.ModelName
	if info.ModelName == "" {
		info.ModelName = "unknown"
	}

	// Set initial progress bar width (will be updated on first window resize event)
	// Assume a reasonable default window width of 120 characters
	defaultWindowWidth := 120
	panelWidth := int(float64(defaultWindowWidth) * 0.25)
	barWidth := panelWidth - 8
	if barWidth < 10 {
		barWidth = 10
	}
	if barWidth > 50 {
		barWidth = 50
	}
	info.TokenProgress.Width = barWidth

	infos := map[uuid.UUID]*AgentInfo{ag.ID: info}

	// Create team context with role loading support
	var tm *team.Team
	var err error
	if len(includePaths) > 0 {
		tm, err = team.NewTeamWithRoles(ag, 10, "", includePaths, configDir)
	} else {
		tm, err = team.NewTeam(ag, 10, "")
	}
	if err != nil {
		panic(fmt.Sprintf("failed to initialize team: %v", err))
	}

	ag.Prompt = core.GetDefaultPrompt()
	if strings.TrimSpace(ag.Prompt) == "" {
		debug.Printf("Warning: No default prompt found. Set AGENTRY_DEFAULT_PROMPT or install templates (see docs). Proceeding without a system prompt.")
	}

	tm.RegisterAgentTool(ag.Tools)

	// Initialize status bar with gradient colors from the agentry logo
	// Using the beautiful purple to teal gradient for a modern, cohesive look
	agentsColors := statusbar.ColorConfig{
		Foreground: lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#ffffff"},
		Background: lipgloss.AdaptiveColor{Light: "#8B5FBF", Dark: "#8B5FBF"}, // Soft purple
	}
	cwdColors := statusbar.ColorConfig{
		Foreground: lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#ffffff"},
		Background: lipgloss.AdaptiveColor{Light: "#5B82D7", Dark: "#5B82D7"}, // Medium blue
	}
	tokensColors := statusbar.ColorConfig{
		Foreground: lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#ffffff"},
		Background: lipgloss.AdaptiveColor{Light: "#2BA6EF", Dark: "#2BA6EF"}, // Bright blue
	}
	costColors := statusbar.ColorConfig{
		Foreground: lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#ffffff"},
		Background: lipgloss.AdaptiveColor{Light: "#00D6EF", Dark: "#00D6EF"}, // Soft teal
	}
	// Put agents first, CWD in the expandable middle, then tokens and cost
	statusBarModel := statusbar.New(agentsColors, cwdColors, tokensColors, costColors)

	m := Model{
		agents:          []*core.Agent{ag},
		infos:           infos,
		order:           []uuid.UUID{ag.ID},
		active:          ag.ID,
		team:            tm,
		vp:              vp,
		debugVp:         debugVp,
		input:           ti,
		tools:           l,
		cwd:             cwd,
		theme:           th,
		keys:            th.Keybinds,
		showInitialLogo: true,
		robot:           NewRobotFace(),
		statusBarModel:  statusBarModel,
		pricing:         cost.NewPricingTable(),
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
