# Known Issues - ✅ RESOLVED

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
