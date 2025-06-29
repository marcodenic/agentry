#!/bin/bash
set -e

# Test Fixed Agent 0 - Delegation First Behavior
echo "======================================="
echo "FIXED AGENT 0 - DELEGATION FIRST TEST"
echo "======================================="

# Clean slate
SANDBOX_DIR="/tmp/agentry-fixed-delegation-test"
rm -rf "$SANDBOX_DIR"
mkdir -p "$SANDBOX_DIR"
cd "$SANDBOX_DIR"

# Copy necessary configuration files and updated agent role
cp /home/marco/Documents/GitHub/agentry/agentry.exe ./agentry
cp /home/marco/Documents/GitHub/agentry/.env.local .
cp /home/marco/Documents/GitHub/agentry/.agentry.yaml .

# Copy the templates directory so Agent 0 has access to updated role definitions
mkdir -p templates/roles
cp -r /home/marco/Documents/GitHub/agentry/templates/roles/* templates/roles/

echo "Test workspace: $SANDBOX_DIR"

echo "Testing FIXED Agent 0 with delegation-first behavior..."

# Create input for pure autonomous task with no hints about which agents to use
cat > /tmp/fixed_delegation_input.txt << 'EOF'
Create a simple Python script that generates a to-do list manager. It should allow users to add, remove, and list tasks. Include proper documentation and error handling.
/quit
EOF

# Run the test
echo "Running fixed delegation test..."
timeout 180s ./agentry chat < /tmp/fixed_delegation_input.txt > /tmp/fixed_delegation_output.txt 2>&1 \
  || echo "Test completed"

echo ""
echo "=== FIXED AGENT 0 RESPONSE ==="
cat /tmp/fixed_delegation_output.txt
echo ""
echo "==============================="

echo ""
echo "ANALYSIS:"
echo "Files in workspace after test:"
ls -la "$SANDBOX_DIR"

echo ""
echo "Checking for Python files:"
if ls "$SANDBOX_DIR"/*.py 2>/dev/null; then
    echo "  Files found - checking if Agent 0 or delegated agent created them:"
    ls -la "$SANDBOX_DIR"/*.py
else
    echo "  ✅ No Python files found - proper delegation without fallback implementation"
fi

echo ""
echo "Delegation pattern analysis:"
grep -q "🔧.*check_agent" /tmp/fixed_delegation_output.txt && echo "  ✅ Agent 0 checked agent availability" || echo "  ❌ Agent 0 did not check agent availability"
grep -q "🔧.*assign_task.*coder\|🔧.*agent.*coder" /tmp/fixed_delegation_output.txt && echo "  ✅ Agent 0 delegated to coder" || echo "  ❌ Agent 0 did not delegate to coder"
grep -q "🔧.*create\|🔧.*edit_range\|🔧.*write" /tmp/fixed_delegation_output.txt && echo "  ❌ Agent 0 implemented directly (BAD)" || echo "  ✅ Agent 0 did not implement directly (GOOD)"

echo ""
echo "This test validates that the fixed Agent 0 delegates first instead of implementing directly."
