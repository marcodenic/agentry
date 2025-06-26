# Test Script for Agentry Fixes
# Tests both:
# 1. Tools are not created as agents
# 2. No duplicate responses in TUI

## Manual Test 1: Tool Name Validation
# Try to run agentry and ask it to delegate to tool names like "echo", "ping", "read_lines"
# Expected: Should get error messages indicating these are reserved tool names

## Manual Test 2: TUI Duplicate Response Check  
# Start agentry in TUI mode and ask a simple question
# Expected: Each agent response should appear only once in the chat history

## Testing Commands:

# Test 1 - Check tool name rejection (should fail with tool name error):
echo "Testing tool name rejection..."

# Test 2 - Run TUI and manually verify no duplicate responses:
echo "Testing TUI duplicate response fix..."

echo "To test:"
echo "1. Run: .\bin\agentry.exe"
echo "2. Try to create agent with tool name (should be blocked)"
echo "3. Create valid agent and verify no duplicate responses"
echo "4. Check that only real agents appear in agent list"

echo "Expected results:"
echo "- Tool names like 'echo', 'ping', 'read_lines' should be rejected as agent names"
echo "- Agent responses should appear only once in chat history"
echo "- Agent list should only show real agents, not tools"
