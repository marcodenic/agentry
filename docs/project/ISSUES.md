# Known Issues and Current Status

## NEW CRITICAL ISSUES - Current Session

### TUI/Interface Issues
- [ ] **Command formatting**: Commands lack proper spacing and grouping
- [ ] **Spinner persistence**: Spinners getting stuck and not clearing properly  
- [ ] **Token count**: Footer token count not updating during conversations
- [ ] **Cost tracking**: Cost display stuck at $0.0000 despite API usage
- [ ] **Status sync**: Agent status panel sometimes out of sync with actual state

### Agent Orchestration Issues
- [ ] **Auto-delegation failure**: Agent 0 identifies tasks but doesn't spawn agents automatically
- [x] **Manual spawn error**: `/spawn coder` command fails with "tool names are reserved" error - FIXED
- [ ] **Team coordination**: No inter-agent communication or status sharing
- [ ] **Task assignment**: No proper task delegation workflow
- [ ] **Progress tracking**: No way to monitor delegated task progress

### Core Functionality Gaps
- [ ] **Agent lifecycle**: No proper agent creation/destruction management
- [x] **Tool restrictions**: Some tool names appear to be reserved/blocked (e.g., "powershell") - FIXED
- [ ] **Context sharing**: Agents don't share context or results
- [ ] **Workflow orchestration**: No multi-step task coordination

## Test Cases Needed

### Basic Orchestration Tests
1. **Simple task delegation**: "Write a Python script to parse CSV files"
2. **Multi-agent collaboration**: "Analyze this codebase and create documentation"
3. **Team formation**: "Plan and implement a web application with tests"
4. **Resource allocation**: "Research and implement best practices for 5 different topics"

### Advanced Workflow Tests  
1. **Project management**: Auto-assign team planner for complex projects
2. **Cross-agent communication**: Agents sharing findings and collaborating
3. **Dynamic scaling**: Adding/removing agents based on workload
4. **Quality assurance**: Automatic testing and review workflows

## Priority Order
1. Fix immediate TUI bugs (spinners, formatting, counters)
2. Resolve agent spawning and tool errors
3. Implement proper task delegation
4. Add inter-agent communication
5. Build comprehensive test scenarios

---

## RESOLVED ISSUES ✅

## TUI Mode Issues - ALL FIXED

**MAJOR DISCOVERY**: The main issue was that `main.go` was using `tui.NewChat()` (ChatModel) instead of `tui.New()` (Model) for single-agent scenarios. The working streaming functionality, proper agent panel, and immediate message display were already implemented in the main `Model` (model.go)!

### Issue #1: User message not immediately displayed
**Description:** When hitting send, the user's message is not immediately shown in the chat. The message appears to be delayed until the AI response arrives, making the UX feel broken.

**Priority:** High
**Status:** Open

### Issue #2: Message text cutoff
**Description:** Messages are being cut off at the end of the sentence rather than properly wrapping to new lines.

**Priority:** High
**Status:** Open

### Issue #3: User input scrolling off screen
**Description:** The user input should wrap text and grow as the user types rather than scrolling horizontally off screen.

**Priority:** Medium
**Status:** Open

### Issue #4: Agent sidebar not visible
**Description:** The right side agent sidebar showing agents and status info is not visible in the TUI.

**Priority:** Medium
**Status:** Open

## Investigation Status
- [x] Analyze TUI chat implementation
- [x] Identify root causes for each issue
- [x] Develop fix plan
- [x] Implement fixes
- [x] Build and test fixes

## Installation Complete ✅

The agentry executable has been successfully built with all TUI fixes applied:
- **Location**: `c:\Users\marco\Documents\GitHub\agentry\agentry.exe`
- **Size**: 27.9 MB
- **Launch Scripts**: 
  - `run-tui.bat` (Windows Batch)
  - `run-tui.ps1` (PowerShell)

### Usage:
```bash
# Direct usage
.\agentry.exe tui

# Or using the launcher scripts
.\run-tui.bat        # Windows Batch
.\run-tui.ps1        # PowerShell
```

## Implemented Fixes

### Issue #1: User message not immediately displayed ✅ FIXED
**Fix Applied**: Modified `callActive` method in `internal/tui/chat.go` to update the viewport content immediately after adding the user message to history, before making the AI call. This ensures user messages appear instantly.

### Issue #2: Messages being cut off ✅ FIXED  
**Fix Applied**: Added text wrapping using `lipgloss.NewStyle().Width(m.vps[idx].Width).Render()` when setting viewport content in both the `callActive` method and window resize handler.

### Issue #3: User input scrolling off screen ✅ PARTIALLY FIXED
**Fix Applied**: Added character limit (1000 chars) and updated placeholder text to indicate multi-line capability. Set input width to match viewport width for consistency.
**Note**: Full multi-line textarea support would require upgrading to bubbles v0.18+ or implementing a custom component.

### Issue #4: Agent sidebar not visible ✅ FIXED
**Fix Applied**: 
- Updated `ChatModel.View()` method to use horizontal layout with left panel (chat + input) and right panel (agent sidebar)
- Added `agentPanel()` method to create the agent status sidebar
- Added `statusDot()` method for colored status indicators
- Properly sized panels using 75%/25% split

### Viewport Sizing Fix ✅ FIXED
**Issue**: Empty black space at the bottom of the TUI, with the status bar not flush against the bottom edge of the terminal.

**Root Cause**: Viewport height calculation was too conservative, leaving unnecessary empty space.

**Fix Applied**:
- ✅ Changed viewport height calculation from `msg.Height - 5` to `msg.Height - 2`
- ✅ Removed extra padding that was causing empty space
- ✅ Status bar now appears flush against the bottom edge of the terminal
- ✅ Maximum viewport space utilization for chat content

**Technical Details**:
- Modified `WindowSizeMsg` handler in `model.go`
- Viewport height = total height - input line - footer line (exactly 2 lines)
- Updated theme snapshots to reflect improved layout
- All tests pass with tighter layout

## UI/UX Improvements ✅ COMPLETED

### Full Screen TUI Experience ✅ ADDED
**Enhancement**: Modified TUI to use the full terminal screen instead of sharing space with previous command history.

**Changes Made**:
- ✅ Added `tea.WithAltScreen()` option to TUI initialization
- ✅ Added `tea.WithMouseCellMotion()` for better mouse support
- ✅ Applied to both main TUI and team conversation interfaces
- ✅ TUI now takes over the entire terminal window
- ✅ Previous commands and prompt are hidden during TUI use
- ✅ Terminal is properly restored when exiting TUI

**Benefits**:
- **Immersive Experience**: TUI now feels like a dedicated application
- **More Space**: Full terminal height available for chat content
- **Professional Feel**: No distracting command history visible
- **Clean Interface**: TUI starts with a fresh, full-screen view

**Technical Details**:
- Modified `cmd/agentry/main.go` to use `tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())`
- Updated `/converse` command TUI to also use alternate screen
- All tests continue to pass

## Root Cause Analysis

### Issue #1: User message not immediately displayed
**Root Cause**: In `internal/tui/chat.go:119-129`, the `callActive` method adds the user message to history but doesn't update the viewport content until after the AI response completes. The viewport content is only updated at the end of the method after `m.team.Call()` returns.

**Fix**: Update the viewport immediately after adding the user message to history, before making the AI call.

### Issue #2: Messages being cut off
**Root Cause**: The viewport width calculation uses `int(float64(msg.Width)*0.75) - 2` but the text content is not being wrapped properly. The `SetContent` calls don't include text wrapping.

**Fix**: Implement proper text wrapping when setting viewport content using lipgloss styling.

### Issue #3: User input scrolling off screen 
**Root Cause**: The `textinput` component doesn't have multi-line support and doesn't have a height limit set. It's configured as a single-line input that scrolls horizontally.

**Fix**: Replace with a multi-line text area component or implement input wrapping.

### Issue #4: Agent sidebar not visible
**Root Cause**: In `internal/tui/chat.go:83-89`, the `View()` method only renders the main viewport and input, but doesn't include the agent sidebar that exists in `model.go`. The `ChatModel` doesn't have the same layout structure as the main `Model`.

**Fix**: Update the `ChatModel.View()` method to include a right sidebar similar to the main model's `agentPanel()`.

# Current Issues and Status

## TUI Unification ✅ COMPLETED

**STATUS**: ✅ **RESOLVED** - TUI has been successfully unified around a single, consistent interface.

**MAJOR DISCOVERY**: The main issue was that `main.go` was using `tui.NewChat()` (ChatModel) instead of `tui.New()` (Model) for single-agent scenarios. The working streaming functionality, proper agent panel, and immediate message display were already implemented in the main `Model` (model.go)!

**RESOLUTION SUMMARY**:
- ✅ **Unified Interface**: All TUI scenarios now use `tui.New()` (the unified Model) regardless of team size or topic
- ✅ **Deprecated ChatModel**: Marked `ChatModel` and `NewChat` as deprecated to prevent future confusion
- ✅ **Updated Tests**: All tests now use the unified Model instead of deprecated ChatModel
- ✅ **Enhanced Help**: Updated help text to reflect unified interface capabilities
- ✅ **Snapshot Updates**: Theme snapshot tests updated to reflect the improved interface

**KEY CHANGES**:
- `main.go`: Refactored to always use `tui.New()` for TUI mode
- `chat.go`: Marked deprecated with clear comments
- `model.go`: Enhanced help text and ensured all features work in unified interface
- `model_test.go`: Updated all tests to use unified Model
- Theme snapshots: Updated to reflect the improved interface

**BENEFITS ACHIEVED**:
- ✅ Single, consistent TUI interface for all scenarios
- ✅ No more confusion between multiple TUI models
- ✅ All features (streaming, agent panel, text wrapping) available everywhere
- ✅ Clean, maintainable codebase with deprecated code clearly marked
- ✅ Better user experience with comprehensive help and input guidance

**NEXT STEPS** (Optional future cleanup):
- Consider removing deprecated `ChatModel` code entirely in future cleanup
- Evaluate if `TeamModel` should also be unified or kept as specialized interface for multi-agent scenarios

**VERIFICATION**:
- ✅ All tests pass: `go test ./internal/tui/`
- ✅ Application builds successfully: `go build ./cmd/agentry`
- ✅ Unified TUI launches correctly for all scenarios
- ✅ Agent panel, streaming, and all features work in unified interface
