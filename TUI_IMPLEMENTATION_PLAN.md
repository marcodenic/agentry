# AGENTRY TUI COMPLETE IMPLEMENTATION PLAN

## Layout Overview

```
┌─────────────────────────────────┬─────────────────┐
│ CHAT / MEMORY (75%)             │ AGENTS (25%)    │
│                                 │                 │
│ User: Hello                     │ ● Agent0 idle   │
│ Agent0: How can I help?         │   role: Master  │
│                                 │   tokens: 120   │
│ > /spawn researcher "find X"    │   ████░░░░░░ 40% │
│                                 │                 │
│ Agent1: Starting research...    │ 🟡 Agent1 run   │
│                                 │   role: Research│
│ [Tab to switch: Chat|Memory]    │   tool: fetch   │
│ ┌─────────────────────────────┐ │   tokens: 45    │
│ │ Input: _                    │ │   ██░░░░░░░░ 15% │
│ └─────────────────────────────┘ │   ▄▂▁▃▆█▄▂▁▃▆█  │
│                                 │                 │
│                                 │ ❌ Agent2 error │
│                                 │   role: DevOps  │
│                                 │   tool: bash    │
│                                 │   tokens: 78    │
│                                 │   ███░░░░░░░ 26% │
├─────────────────────────────────┴─────────────────┤
│ Status: cwd: /workspace | agents: 3 | tokens: 243  │
└─────────────────────────────────────────────────────┘
```

## Current State Analysis

- ✅ Multi-agent model using `agents []*core.Agent`
- ✅ Basic command system (`/spawn`, `/switch`, `/stop`, `/converse`)
- ✅ Simple agent panel listing active agents
- ❌ No real-time spinners or progress bars
- ❌ Limited layout and theming options

## Required Complete Implementation

### 1. Multi-Agent Model Architecture

```go
type Model struct {
    // Replace single agent with agent orchestrator
    masterAgent *core.Agent           // Agent 0 - primary interface
    agents      map[uuid.UUID]*AgentInfo  // All active agents    // Enhanced UI components
    chatPanel    viewport.Model       // Left main - conversation
    memoryPanel  viewport.Model       // Left alt - debug/memory
    agentsPanel  AgentsPanel           // Right - agent status & activity
    statusBar    StatusBar            // Bottom - global status

    // Real-time state
    spinners     map[uuid.UUID]*spinner.Model
    timers       map[uuid.UUID]time.Time
    tokenCounts  map[uuid.UUID]int

    // Command system
    commandMode  bool
    commandInput textinput.Model    // Layout control
    leftWidth    int   // Chat area (75%)
    rightWidth   int   // Agents dashboard (25%)
    activeTab    int   // 0=chat, 1=memory

    // Multi-agent coordination
    eventStream  chan AgentEvent
    cancelFuncs  map[uuid.UUID]context.CancelFunc
}

type AgentInfo struct {
    ID           uuid.UUID
    Name         string
    Role         string
    Status       AgentStatus  // idle, running, error, stopped
    Spinner      spinner.Model
    StartTime    time.Time
    TokenCount   int
    CurrentTool  string       // Currently executing tool
    LastActivity string
    ModelName    string

    // Token usage tracking
    TokenHistory []int        // For sparkline/activity chart
    MaxTokens    int         // Context window size for percentage
}

type AgentStatus int
const (
    StatusIdle AgentStatus = iota
    StatusRunning
    StatusError
    StatusStopped
)
```

### 2. Enhanced Layout Management

```go
func (m Model) View() string {
    // Left Panel (75% width) - Main Chat Area
    var leftContent string
    switch m.activeTab {
    case 0: // Chat
        leftContent = m.chatPanel.View() + "\n" + m.renderInput()
    case 1: // Memory/Debug
        leftContent = m.renderMemoryView()
    }
    leftStyled := m.theme.LeftPanelStyle.Width(m.leftWidth).Render(leftContent)

    // Right Panel (25% width) - Agents Dashboard
    agentsView := m.renderAgentsPanel()
    rightStyled := m.theme.RightPanelStyle.Width(m.rightWidth).Render(agentsView)

    // Main layout
    main := lipgloss.JoinHorizontal(lipgloss.Top, leftStyled, rightStyled)

    // Status bar
    statusBar := m.renderStatusBar()

    return lipgloss.JoinVertical(lipgloss.Left, main, statusBar)
}
```

### 3. Real-Time Agent Status Visualization

```go
func (m Model) renderAgentsPanel() string {
    var lines []string
    lines = append(lines, m.theme.PanelTitleStyle.Render("🤖 AGENTS"))
    lines = append(lines, "") // Empty line for spacing

    for _, agent := range m.agents {
        statusDot := m.getStatusDot(agent.Status)

        // Agent name line with status
        var nameeLine string
        if agent.Status == StatusRunning {
            nameeLine = fmt.Sprintf("%s %s %s",
                statusDot, agent.Spinner.View(), agent.Name)
        } else {
            nameeLine = fmt.Sprintf("%s %s", statusDot, agent.Name)
        }        lines = append(lines, nameeLine)

        // Role/task line
        if agent.Role != "" {
            roleLine := fmt.Sprintf("  role: %s", agent.Role)
            lines = append(lines, m.theme.RoleColor.Render(roleLine))
        }

        // Current tool line (if active)
        if agent.CurrentTool != "" {
            toolLine := fmt.Sprintf("  tool: %s", agent.CurrentTool)
            lines = append(lines, m.theme.ToolColor.Render(toolLine))
        }

        // Token count line
        tokenLine := fmt.Sprintf("  tokens: %d", agent.TokenCount)
        lines = append(lines, tokenLine)

        // Token usage bar
        if agent.TokenCount > 0 {
            bar := m.renderTokenBar(agent.TokenCount, agent.MaxTokens)
            lines = append(lines, "  "+bar)
        }

        // Token activity sparkline (if we have history)
        if len(agent.TokenHistory) > 0 {
            sparkline := m.renderSparkline(agent.TokenHistory)
            lines = append(lines, "  "+sparkline)
        }

        lines = append(lines, "") // Spacing between agents
    }

    return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) getStatusDot(status AgentStatus) string {
    switch status {
    case StatusIdle:
        return m.theme.IdleColor.Render("●")
    case StatusRunning:
        return m.theme.RunningColor.Render("🟡")  // Yellow circle emoji for running
    case StatusError:
        return m.theme.ErrorColor.Render("❌")    // Red X for errors
    case StatusStopped:
        return m.theme.StoppedColor.Render("⏸️")  // Pause symbol for stopped
    default:
        return "○"
    }
}

func (m Model) renderTokenBar(count, max int) string {
    if max == 0 {
        max = 1000 // Default scale
    }
    pct := float64(count) / float64(max)
    if pct > 1.0 {
        pct = 1.0
    }

    filled := int(pct * 10) // 10-char bar
    bar := strings.Repeat("█", filled) + strings.Repeat("░", 10-filled)
    return fmt.Sprintf("%s %d%%", bar, int(pct*100))
}

func (m Model) renderSparkline(history []int) string {
    if len(history) == 0 {
        return ""
    }

    // Use braille characters for sparkline
    chars := []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}

    // Find min/max for scaling
    min, max := history[0], history[0]
    for _, v := range history {
        if v < min {
            min = v
        }
        if v > max {
            max = v
        }
    }

    var sparkline strings.Builder
    for _, v := range history {
        if max == min {
            sparkline.WriteString(chars[0])
        } else {
            normalized := float64(v-min) / float64(max-min)
            idx := int(normalized * float64(len(chars)-1))
            sparkline.WriteString(chars[idx])
        }
    }

    return sparkline.String()
}
```

### 4. Command System Implementation

```go
func (m Model) handleCommand(cmd string) (Model, tea.Cmd) {
    parts := strings.Fields(cmd)
    if len(parts) == 0 {
        return m, nil
    }

    switch parts[0] {
    case "/spawn":
        return m.handleSpawnCommand(parts[1:])
    case "/converse":
        return m.handleConverseCommand(parts[1:])
    case "/stop":
        return m.handleStopCommand(parts[1:])
    case "/switch":
        return m.handleSwitchCommand(parts[1:])
    case "/agents", "/status":
        return m.handleStatusCommand()
    case "/help":
        return m.handleHelpCommand()
    default:
        return m.addChatMessage("Unknown command: " + parts[0], true), nil
    }
}

func (m Model) handleSpawnCommand(args []string) (Model, tea.Cmd) {
    if len(args) < 1 {
        return m.addChatMessage("Usage: /spawn [name] [role] [prompt]", true), nil
    }

    name := args[0]
    var role string
    var prompt string

    if len(args) >= 2 {
        role = args[1]
        if len(args) >= 3 {
            prompt = strings.Join(args[2:], " ")
        }
    }

    // Spawn new agent
    subAgent := m.masterAgent.Spawn()
    agentInfo := &AgentInfo{
        ID:       subAgent.ID,
        Name:     name,
        Role:     role,
        Status:   StatusIdle,
        Spinner:  spinner.New(),
        StartTime: time.Now(),
        MaxTokens: 8000, // Default context window
    }

    m.agents[subAgent.ID] = agentInfo

    // Start agent if prompt provided
    if prompt != "" {
        return m.startAgent(subAgent.ID, prompt)
    }

    statusMsg := fmt.Sprintf("Spawned %s", name)
    if role != "" {
        statusMsg += fmt.Sprintf(" (role: %s)", role)
    }
    statusMsg += fmt.Sprintf(" [%s]", subAgent.ID.String()[:8])

    m = m.addChatMessage(statusMsg, false)
    return m, nil
}

func (m Model) startAgent(agentID uuid.UUID, input string) (Model, tea.Cmd) {
    agent := m.agents[agentID]
    agent.Status = StatusRunning
    agent.StartTime = time.Now()

    // Create cancellable context
    ctx, cancel := context.WithCancel(context.Background())
    m.cancelFuncs[agentID] = cancel

    // Start agent in goroutine
    cmd := func() tea.Msg {
        result, err := m.getAgent(agentID).Run(ctx, input)
        return agentCompleteMsg{
            agentID: agentID,
            result:  result,
            err:     err,
        }
    }

    return m, cmd
}
```

### 5. Real-Time Event Handling

```go
type agentCompleteMsg struct {
    agentID uuid.UUID
    result  string
    err     error
}

type agentTokenMsg struct {
    agentID uuid.UUID
    token   string
}

type agentStatusMsg struct {
    agentID uuid.UUID
    status  string
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd

    switch msg := msg.(type) {
    case tea.KeyMsg:
        if m.commandMode {
            return m.handleCommandInput(msg)
        }
        return m.handleNormalInput(msg)

    case agentCompleteMsg:
        return m.handleAgentComplete(msg)

    case agentTokenMsg:
        return m.handleAgentToken(msg)

    case agentStatusMsg:
        return m.handleAgentStatus(msg)

    case spinner.TickMsg:
        return m.updateSpinners(msg)

    case tea.WindowSizeMsg:
        return m.handleResize(msg)
    }

    return m, tea.Batch(cmds...)
}

func (m Model) updateSpinners(msg spinner.TickMsg) (Model, tea.Cmd) {
    var cmds []tea.Cmd

    for id, agent := range m.agents {
        if agent.Status == StatusRunning {
            var cmd tea.Cmd
            agent.Spinner, cmd = agent.Spinner.Update(msg)
            cmds = append(cmds, cmd)
            m.agents[id] = agent
        }
    }

    return m, tea.Batch(cmds...)
}
```

### 6. Enhanced Theme System

```go
type Theme struct {
    // Agent colors (expandable array)
    AgentColors []lipgloss.Color    // Status colors
    IdleColor    lipgloss.Color
    RunningColor lipgloss.Color
    ErrorColor   lipgloss.Color
    StoppedColor lipgloss.Color
    ToolColor    lipgloss.Color    // For current tool display
    RoleColor    lipgloss.Color    // For agent role/task display

    // Panel styles
    LeftPanelStyle  lipgloss.Style
    RightPanelStyle lipgloss.Style
    PanelTitleStyle lipgloss.Style

    // Legacy compatibility
    UserBarColor string
    AIBarColor   string

    // Keybinds
    Keybinds Keybinds
}

func DefaultTheme() Theme {
    return Theme{
        AgentColors: []lipgloss.Color{
            lipgloss.Color("#8B5CF6"), // Master agent - purple
            lipgloss.Color("#FBBF24"), // Agent 1 - yellow
            lipgloss.Color("#34D399"), // Agent 2 - green
            lipgloss.Color("#F87171"), // Agent 3 - red
            lipgloss.Color("#60A5FA"), // Agent 4 - blue
        },        IdleColor:    lipgloss.Color("#22C55E"),
        RunningColor: lipgloss.Color("#FBBF24"),
        ErrorColor:   lipgloss.Color("#EF4444"),
        StoppedColor: lipgloss.Color("#6B7280"),
        ToolColor:    lipgloss.Color("#8B5CF6"),  // Purple for tool names
        RoleColor:    lipgloss.Color("#10B981"),  // Green for agent roles
          LeftPanelStyle: lipgloss.NewStyle().    // Chat/Memory area (75%)
            Border(lipgloss.NormalBorder()).
            BorderForeground(lipgloss.Color("#374151")).
            Padding(0, 1),
              RightPanelStyle: lipgloss.NewStyle().   // Agents dashboard (25%)
            Border(lipgloss.NormalBorder()).
            BorderForeground(lipgloss.Color("#374151")).
            Padding(0, 1),

        PanelTitleStyle: lipgloss.NewStyle().
            Bold(true).
            Foreground(lipgloss.Color("#9CA3AF")),
    }
}
```

## Implementation Priority Order

1. **Model Refactor** - Replace single agent with multi-agent orchestrator
2. **Layout System** - Implement 4-panel layout with proper sizing
3. **Agent Status Visualization** - Real-time status dots, spinners, token bars, role display
4. **Command System** - /spawn, /converse, /stop, /switch commands
5. **Event Handling** - Multi-agent event routing and UI updates
6. **Theme Enhancement** - Extended color system and styling
7. **Interactivity** - Scrolling, non-blocking input, cancellation
8. **Polish** - Error handling, edge cases, performance optimization

## Testing Strategy

1. **Unit Tests** - Each component in isolation
2. **Integration Tests** - Multi-agent scenarios
3. **Manual Testing** - Various terminal sizes and platforms
4. **Performance Testing** - Multiple agents with high-frequency updates
5. **Edge Case Testing** - Rapid spawning/stopping, cancellation, errors

This represents a complete rewrite of the TUI system to meet the world-class multi-agent requirements outlined in your specification.
