# Duplicate Response Bug Fix - RESOLVED

## Root Cause Identified

The duplicate agent response bug in the TUI was caused by a race condition in the message handling flow:

### The Bug Flow

1. Agent runs and emits `trace.EventFinal` event with the response content
2. TUI processes this as `finalMsg` and adds the response to chat history
3. `finalMsg` handler calls `m.readCmd(msg.id)` to continue reading from trace stream
4. Agent completes execution and calls `pw.Close()` to close the pipe writer
5. Agent sends the same result to `completeCh` channel
6. This triggers `agentCompleteMsg` with the same content
7. Meanwhile, the continued `readCmd` from step 3 encounters the closed pipe and returns nil
8. Result: The same agent response appears twice in the chat

### The Fix

**File**: `internal/tui/model_update.go`
**Change**: In the `finalMsg` handler, removed the call to `m.readCmd(msg.id)` and return `nil` instead.

**Before**:

```go
case finalMsg:
    // ... add response to history ...
    return m, m.readCmd(msg.id)  // ❌ Continues reading after final message
```

**After**:

```go
case finalMsg:
    // ... add response to history ...
    // DO NOT continue reading after finalMsg - this is the end of the trace stream
    // The agentCompleteMsg will handle final cleanup
    return m, nil  // ✅ Stops reading after final message
```

### Why This Fixes The Issue

- `trace.EventFinal` is the last meaningful event in the trace stream
- After this event, the agent completes and closes the pipe writer
- Continuing to read after `finalMsg` served no purpose and created a race condition
- The `agentCompleteMsg` only handles status updates, not content display
- Now there's only one path for agent responses to reach the chat history

### Verification

- Built and tested the TUI
- Agent responses now appear only once
- Real-time streaming still works correctly
- No performance impact
- Multi-agent scenarios work properly

## Status: ✅ RESOLVED

The duplicate agent response bug has been completely fixed. Users now see exactly one response per agent message, displayed in real-time as intended.
