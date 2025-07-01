#!/bin/bash
set -e

# Test Agent 0's awareness of its coordination tools
echo "================================"
echo "COORDINATION TOOLS AWARENESS TEST"
echo "================================"

# Clean slate
SANDBOX_DIR="/tmp/agentry-coordination-awareness-test"
rm -rf "$SANDBOX_DIR"
mkdir -p "$SANDBOX_DIR"
cd "$SANDBOX_DIR"

# Copy necessary configuration files
cp /home/marco/Documents/GitHub/agentry/agentry.exe ./agentry
cp /home/marco/Documents/GitHub/agentry/.env.local .
cp /home/marco/Documents/GitHub/agentry/.agentry.yaml .

echo "Test workspace: $SANDBOX_DIR"

echo "Testing Agent 0's awareness of its coordination tools..."

# Create input that explicitly asks Agent 0 to check its capabilities
cat > /tmp/coordination_awareness_input.txt << 'EOF'
I need you to help me understand your coordination capabilities. Please:

1. Use your team_status tool to check what team agents are available
2. Use your check_agent tool to test if the 'coder' agent is available
3. If you have other agents available, check them too
4. Then explain how you would coordinate a task across multiple agents

Show me your actual coordination tools in action.
/quit
EOF

# Run the test
echo "Running coordination awareness test..."
timeout 120s ./agentry chat < /tmp/coordination_awareness_input.txt > /tmp/coordination_awareness_output.txt 2>&1 \
  || echo "Test completed"

echo ""
echo "=== AGENT 0 COORDINATION AWARENESS RESPONSE ==="
cat /tmp/coordination_awareness_output.txt
echo ""
echo "============================================="

echo ""
echo "ANALYSIS:"
echo "- Did Agent 0 use team_status tool?"
grep -q "team_status" /tmp/coordination_awareness_output.txt && echo "  ✅ YES - Agent 0 used team_status" || echo "  ❌ NO - Agent 0 did not use team_status"

echo "- Did Agent 0 use check_agent tool?"
grep -q "check_agent" /tmp/coordination_awareness_output.txt && echo "  ✅ YES - Agent 0 used check_agent" || echo "  ❌ NO - Agent 0 did not use check_agent"

echo "- Did Agent 0 demonstrate actual tool usage (not just description)?"
grep -q "🔧.*using a tool" /tmp/coordination_awareness_output.txt && echo "  ✅ YES - Agent 0 actually used tools" || echo "  ❌ NO - Agent 0 only described tools"

echo ""
echo "This test checks if Agent 0 is aware of and willing to use its coordination tools when explicitly asked."
