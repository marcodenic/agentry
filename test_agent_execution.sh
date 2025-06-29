#!/bin/bash
set -e

# Test if agents actually execute assigned tasks
echo "================================"
echo "AGENT EXECUTION VERIFICATION TEST"
echo "================================"

# Clean slate
SANDBOX_DIR="/tmp/agentry-execution-test"
rm -rf "$SANDBOX_DIR"
mkdir -p "$SANDBOX_DIR"
cd "$SANDBOX_DIR"

# Copy necessary configuration files
cp /home/marco/Documents/GitHub/agentry/agentry.exe ./agentry
cp /home/marco/Documents/GitHub/agentry/.env.local .
cp /home/marco/Documents/GitHub/agentry/.agentry.yaml .

echo "Test workspace: $SANDBOX_DIR"

echo "Testing if assigned agents actually execute their tasks..."

# Create input that asks Agent 0 to delegate a concrete task that should produce visible output
cat > /tmp/execution_test_input.txt << 'EOF'
I need you to coordinate creation of a simple "Hello World" Python script. Please:

1. Check what agents are available
2. Assign the task to the appropriate agent (coder) to create a file called "hello_world.py" with a simple hello world script
3. Wait for the task to be completed
4. Verify the file was created

The goal is to see if delegated agents actually execute tasks and produce real output.
/quit
EOF

# Run the test
echo "Running execution verification test..."
timeout 180s ./agentry chat < /tmp/execution_test_input.txt > /tmp/execution_test_output.txt 2>&1 \
  || echo "Test completed"

echo ""
echo "=== AGENT 0 EXECUTION TEST RESPONSE ==="
cat /tmp/execution_test_output.txt
echo ""
echo "======================================="

echo ""
echo "ANALYSIS:"
echo "Files in workspace after test:"
ls -la "$SANDBOX_DIR"

echo ""
echo "Checking for hello_world.py:"
if [ -f "$SANDBOX_DIR/hello_world.py" ]; then
    echo "  ✅ SUCCESS - hello_world.py was created!"
    echo "  Content:"
    cat "$SANDBOX_DIR/hello_world.py"
else
    echo "  ❌ FAILED - hello_world.py was not created"
fi

echo ""
echo "Tool usage analysis:"
grep -q "🔧.*using a tool" /tmp/execution_test_output.txt && echo "  ✅ Agent 0 used tools" || echo "  ❌ Agent 0 did not use tools"

echo ""
echo "Agent coordination analysis:"
grep -q "assign_task.*coder" /tmp/execution_test_output.txt && echo "  ✅ Agent 0 assigned task to coder" || echo "  ❌ Agent 0 did not assign task to coder"

echo ""
echo "This test verifies if Agent 0's coordination actually results in agent execution and file creation."
