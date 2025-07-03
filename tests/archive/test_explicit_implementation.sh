#!/bin/bash
set -e

# Test Fixed Agent 0 with Explicit Implementation Request
echo "=================================================="
echo "FIXED AGENT 0 - EXPLICIT IMPLEMENTATION REQUEST"
echo "=================================================="

# Clean slate
SANDBOX_DIR="/tmp/agentry-explicit-implementation-test"
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

echo "Testing FIXED Agent 0 with explicit implementation request..."

# Create input that clearly asks for file creation
cat > /tmp/explicit_implementation_input.txt << 'EOF'
I need you to create actual files in the current directory. Please implement a simple Python calculator that can add, subtract, multiply, and divide numbers. 

The implementation should:
1. Create a Python file called "calculator.py" 
2. Include proper function definitions
3. Include a main function for user interaction
4. Save the file to disk in the current working directory

This is not a request for code examples - I need you to actually create the working files.
/quit
EOF

# Run the test
echo "Running explicit implementation test..."
timeout 180s ./agentry chat < /tmp/explicit_implementation_input.txt > /tmp/explicit_implementation_output.txt 2>&1 \
  || echo "Test completed"

echo ""
echo "=== FIXED AGENT 0 RESPONSE ==="
cat /tmp/explicit_implementation_output.txt
echo ""
echo "================================"

echo ""
echo "ANALYSIS:"
echo "Files in workspace after test:"
ls -la "$SANDBOX_DIR"

echo ""
echo "Checking for calculator.py:"
if [ -f "$SANDBOX_DIR/calculator.py" ]; then
    echo "  ✅ calculator.py was created!"
    echo "  Content preview:"
    head -10 "$SANDBOX_DIR/calculator.py"
    echo "  ..."
else
    echo "  ❌ calculator.py was not created"
fi

echo ""
echo "Coordination analysis:"
grep -q "🔧.*check_agent" /tmp/explicit_implementation_output.txt && echo "  ✅ Agent 0 checked agent availability" || echo "  ❌ Agent 0 did not check agent availability"
grep -q "🔧.*team_status" /tmp/explicit_implementation_output.txt && echo "  ✅ Agent 0 checked team status" || echo "  ❌ Agent 0 did not check team status"
grep -q "🔧.*assign_task.*coder\|🔧.*agent.*coder" /tmp/explicit_implementation_output.txt && echo "  ✅ Agent 0 delegated to coder" || echo "  ❌ Agent 0 did not delegate to coder"
grep -q "🔧.*create" /tmp/explicit_implementation_output.txt && echo "  ⚠️  Agent 0 used create tool directly" || echo "  ✅ Agent 0 did not use create tool directly"

echo ""
echo "This test checks if explicit implementation requests trigger proper delegation behavior."
