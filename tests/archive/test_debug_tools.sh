#!/bin/bash
set -e

# Debug Agent 0's Actual Tool Registry
echo "============================================="
echo "DEBUG: AGENT 0 ACTUAL TOOL REGISTRY CHECK"
echo "============================================="

# Clean slate
SANDBOX_DIR="/tmp/agentry-debug-tools-test"
rm -rf "$SANDBOX_DIR"
mkdir -p "$SANDBOX_DIR"
cd "$SANDBOX_DIR"

# Copy necessary configuration files
cp /home/marco/Documents/GitHub/agentry/agentry.exe ./agentry
cp /home/marco/Documents/GitHub/agentry/.env.local .
cp /home/marco/Documents/GitHub/agentry/.agentry.yaml .

# Copy the templates directory
mkdir -p templates/roles
cp -r /home/marco/Documents/GitHub/agentry/templates/roles/* templates/roles/

echo "Test workspace: $SANDBOX_DIR"
echo "Checking what tools Agent 0 actually has access to..."
echo "============================================="

# Create input that asks Agent 0 to list its tools
cat > /tmp/debug_tools_input.txt << 'EOF'
I need to debug what's happening. Please tell me:

1. What tools do you have access to?
2. Can you see a "create" tool in your available tools?
3. Can you see "edit_range" or "write" tools?
4. List all tools you can use for file operations

This is a debugging session to understand your tool configuration.
/quit
EOF

echo "DEBUGGING: Asking Agent 0 to report its available tools"
echo "Expected: Agent 0 should NOT have create, edit_range, write tools"
echo "============================================="

# Run and capture output
timeout 60s ./agentry chat < /tmp/debug_tools_input.txt > /tmp/debug_tools_output.txt 2>&1 || true

echo "=== AGENT 0 TOOL DEBUGGING RESPONSE ==="
cat /tmp/debug_tools_output.txt
echo "========================================"

echo ""
echo "ANALYSIS:"
grep -i "create\|edit_range\|write" /tmp/debug_tools_output.txt && echo "❌ Agent 0 still has forbidden tools!" || echo "✅ Agent 0 properly reports no forbidden tools"

echo ""
echo "Additionally, let's check the configuration files being used:"
echo ""
echo "=== .agentry.yaml configuration ==="
head -20 .agentry.yaml
echo "..."

echo ""
echo "=== agent_0.yaml builtins section ==="
grep -A 20 "builtins:" templates/roles/agent_0.yaml || echo "No builtins section found"

echo ""
echo "This debug test checks if the tool restrictions are actually being applied."
