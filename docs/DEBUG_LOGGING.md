# Agentry Debug Logging System

## Overview

The enhanced debug logging system provides comprehensive, structured logging of all agentry operations to help you troubleshoot agent interactions, tool executions, and model API calls.

## Key Features

✅ **Rolling Log Files**: Automatically creates new log files when size limit reached (1MB per file)
✅ **Structured Events**: Categorized events (TOOL, AGENT, MODEL, etc.) for easy filtering
✅ **TUI Compatible**: Works seamlessly in TUI mode without interfering with the interface
✅ **Real-time Monitoring**: Monitor logs while agentry is running
✅ **Comprehensive Coverage**: Logs everything from API calls to tool executions

## Quick Start

### Method 1: Debug Wrapper Script (Recommended)
```bash
# Use the debug wrapper for comprehensive logging
./scripts/debug-agentry.sh

# Or with specific commands
./scripts/debug-agentry.sh "create a hello world program"
./scripts/debug-agentry.sh tui
```

### Method 2: Environment Variables
```bash
# Set debug level for maximum verbosity
export AGENTRY_DEBUG_LEVEL=trace

# Run agentry normally
./agentry tui
```

### Method 3: Test Script
```bash
# Run the test script to verify logging works
./scripts/test_debug_logging.sh
```

## Log File Location

Debug logs are written to: `debug/agentry-debug-*.log`

Each log file is timestamped and limited to 1MB. When a file reaches the size limit, a new file is automatically created.

## Event Categories

### 🔧 TOOL Events
- Tool execution starts and completions
- Tool arguments and results
- Tool execution duration and success/failure status

### 🤖 AGENT Events  
- Agent actions and state changes
- Agent delegation and coordination
- Agent iteration tracking

### 🌐 MODEL Events
- API request/response details
- Token usage and costs
- Model interaction timing
- Streaming response processing

### 📡 TEAM Events
- Agent communication and coordination
- Team-level operations and status

## Real-time Monitoring

```bash
# Monitor logs in real-time
tail -f debug/agentry-debug-*.log

# Search for specific issues
grep -i 'error\|fail\|exception' debug/agentry-debug-*.log

# Find tool executions
grep 'TOOL.*call' debug/agentry-debug-*.log

# Track agent actions  
grep 'AGENT' debug/agentry-debug-*.log

# Monitor model interactions
grep 'MODEL' debug/agentry-debug-*.log
```

## Debug Levels

- `debug`: Standard debug output
- `trace`: Maximum verbosity including all API communications

## Integration with Existing Systems

The new debug system is fully backward compatible with existing logging:

- Maintains compatibility with `agent_communication.log`
- Works alongside existing `AGENTRY_DEBUG` and `AGENTRY_COMM_LOG` flags
- Enhances rather than replaces existing logging

## Common Use Cases

### Troubleshooting TUI Interactions

1. Start TUI with debug logging:
   ```bash
   ./scripts/debug-agentry.sh tui
   ```

2. In another terminal, monitor logs:
   ```bash
   tail -f debug/agentry-debug-*.log
   ```

3. Perform your interaction in the TUI

4. Review the logs to see exactly what happened

### Analyzing Tool Execution Issues

```bash
# Find all tool executions and their results
grep 'TOOL.*call' debug/agentry-debug-*.log | head -20

# Find failed tool executions
grep -A5 -B5 'has_error:true' debug/agentry-debug-*.log
```

### Model API Debugging

```bash
# See all model interactions with timing
grep 'MODEL.*interaction' debug/agentry-debug-*.log

# Find API errors
grep -i 'api_error\|http_request_failed' debug/agentry-debug-*.log
```

## Performance Impact

- **Minimal**: File logging is asynchronous and highly optimized
- **Rolling logs**: Automatic cleanup prevents disk space issues  
- **Structured format**: Easy parsing without performance overhead

## Example Log Entries

```
2025/09/26 22:58:56.269503 [EVENT] AGENT.tool_execution_start: map[agent_id:471c2d56 tool_name:agent tool_call_id:call_xyz args_count:7]

2025/09/26 22:58:59.779668 [EVENT] TOOL.call: map[tool:create args:map[content:Hello World path:hello.txt] result:{"created":true,"path":"hello.txt"} error:<nil>]

2025/09/26 22:58:56.268159 [EVENT] MODEL.interaction: map[provider:openai model:gpt-5-mini duration:11.739s tokens:map[input_tokens:1274 output_tokens:32]]
```

## Files Added/Modified

### New Files:
- `scripts/debug-agentry.sh` - Debug wrapper script
- `scripts/test_debug_logging.sh` - Test script for verification
- `docs/DEBUG_LOGGING.md` - This documentation

### Enhanced Files:
- `internal/debug/debug.go` - Rolling logger and structured events
- `internal/model/openai.go` - Enhanced model interaction logging  
- `internal/core/agent.go` - Detailed tool execution logging
- `internal/team/logging.go` - Team communication logging

This debug logging system gives you complete visibility into agentry's operation, making it much easier to diagnose issues and understand exactly what your agents are doing.
