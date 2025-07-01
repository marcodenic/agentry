#!/bin/bash
set -e

# Test Agent 0's delegation-first behavior
echo "======================================"
echo "DELEGATION-FIRST COORDINATION TEST"
echo "======================================"

# Clean slate
SANDBOX_DIR="/tmp/agentry-delegation-first-test"
rm -rf "$SANDBOX_DIR"
mkdir -p "$SANDBOX_DIR"
cd "$SANDBOX_DIR"

# Copy necessary configuration files
cp /home/marco/Documents/GitHub/agentry/agentry.exe ./agentry
cp /home/marco/Documents/GitHub/agentry/.env.local .
cp /home/marco/Documents/GitHub/agentry/.agentry.yaml .

echo "Test workspace: $SANDBOX_DIR"

echo "Testing if Agent 0 can delegate FIRST without doing the work itself..."

# Create input that explicitly asks Agent 0 to coordinate via delegation only
cat > /tmp/delegation_first_input.txt << 'EOF'
You are acting as a team coordinator. Your role is to DELEGATE tasks to appropriate agents, NOT to do the work yourself.

Task: Create a Python calculator that can perform basic operations (add, subtract, multiply, divide).

Requirements:
- You must delegate this to the appropriate specialist agent
- You should NOT write the code yourself
- You should coordinate and oversee, but let the specialist do the actual implementation
- Only delegate to agents that actually exist

Please act as a pure coordinator and delegate this task appropriately.
/quit
EOF

# Run the test
echo "Running delegation-first test..."
timeout 120s ./agentry chat < /tmp/delegation_first_input.txt > /tmp/delegation_first_output.txt 2>&1 \
  || echo "Test completed"

echo ""
echo "=== AGENT 0 DELEGATION-FIRST RESPONSE ==="
cat /tmp/delegation_first_output.txt
echo ""
echo "=========================================="

echo ""
echo "ANALYSIS:"
echo "Files in workspace after test:"
ls -la "$SANDBOX_DIR"

echo ""
echo "Checking for Python files:"
if [ -f "$SANDBOX_DIR"/*.py ]; then
    echo "  ⚠️  Python files found - Agent 0 may have done the work itself:"
    ls -la "$SANDBOX_DIR"/*.py
else
    echo "  ✅ No Python files found - Agent 0 acted as pure coordinator"
fi

echo ""
echo "Delegation analysis:"
grep -q "assign_task.*coder" /tmp/delegation_first_output.txt && echo "  ✅ Agent 0 assigned task to coder" || echo "  ❌ Agent 0 did not assign task to coder"

echo ""
echo "Self-execution analysis:"
grep -q "🔧.*→ Using tool: create" /tmp/delegation_first_output.txt && echo "  ❌ Agent 0 created files itself" || echo "  ✅ Agent 0 did not create files itself"

echo ""
echo "This test checks if Agent 0 can act as a pure coordinator without doing the implementation work."
