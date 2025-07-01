#!/bin/bash
set -e

# Test Fixed Agent 0 - No Direct Implementation Tools
echo "====================================================="
echo "FIXED AGENT 0 - NO DIRECT IMPLEMENTATION TOOLS TEST"
echo "====================================================="

# Clean slate
SANDBOX_DIR="/tmp/agentry-no-direct-tools-test"
rm -rf "$SANDBOX_DIR"
mkdir -p "$SANDBOX_DIR"
cd "$SANDBOX_DIR"

# Copy necessary configuration files and updated agent role
cp /home/marco/Documents/GitHub/agentry/agentry.exe ./agentry
cp /home/marco/Documents/GitHub/agentry/.env.local .
cp /home/marco/Documents/GitHub/agentry/.agentry.yaml .

# Copy the templates directory with the fixed Agent 0 role
mkdir -p templates/roles
cp -r /home/marco/Documents/GitHub/agentry/templates/roles/* templates/roles/

echo "Test workspace: $SANDBOX_DIR"
echo "Testing Agent 0 WITHOUT direct implementation tools..."
echo "Agent 0 can no longer use: create, edit_range, write, insert_at, search_replace"
echo "Agent 0 must delegate to coder agent or fail"
echo "====================================================="

# Create input for implementation task
cat > /tmp/no_direct_tools_input.txt << 'EOF'
Create a Python script called "fibonacci.py" that generates the Fibonacci sequence up to a specified number of terms. The script should include:

1. A function to generate the sequence
2. User input to specify how many terms to generate
3. Proper error handling for invalid input
4. Save the file to the current directory

This requires actual file creation.
/quit
EOF

echo "TASK: Create fibonacci.py script"
echo "EXPECTED: Agent 0 must delegate to coder agent (cannot create files itself)"
echo "====================================================="

# Run with enhanced output
timeout 180s stdbuf -oL -eL ./agentry chat < /tmp/no_direct_tools_input.txt 2>&1 | \
while IFS= read -r line; do
    echo "[$(date '+%H:%M:%S')] $line"
    
    # Highlight coordination activities
    case "$line" in
        *"🔧"*"team_status"*)
            echo ">>> [COORDINATION] Agent 0 is checking team status!"
            ;;
        *"🔧"*"check_agent"*)
            echo ">>> [COORDINATION] Agent 0 is checking agent availability!"
            ;;
        *"🔧"*"assign_task"*)
            echo ">>> [COORDINATION] Agent 0 is assigning a task!"
            ;;
        *"🔧"*"agent"*)
            echo ">>> [DELEGATION] Agent 0 is delegating to another agent!"
            ;;
        *"🔧"*"send_message"*)
            echo ">>> [COORDINATION] Agent 0 is sending a message!"
            ;;
        *"🔧"*"create"*)
            echo ">>> [ERROR] Agent 0 tried to create files directly - SHOULD BE IMPOSSIBLE!"
            ;;
        *"🔧"*"edit_range"*)
            echo ">>> [ERROR] Agent 0 tried to edit files directly - SHOULD BE IMPOSSIBLE!"
            ;;
        *"🔧"*"write"*)
            echo ">>> [ERROR] Agent 0 tried to write files directly - SHOULD BE IMPOSSIBLE!"
            ;;
        *"Agent"*"spawned"*|*"Agent"*"available"*)
            echo ">>> [AGENT STATUS] $line"
            ;;
        *"Task assigned"*)
            echo ">>> [TASK ASSIGNED] $line"
            ;;
        *"Error"*|*"failed"*|*"not available"*)
            echo ">>> [ERROR/FAILURE] $line"
            ;;
    esac
done

echo ""
echo "====================================================="
echo "ANALYSIS: FORCED DELEGATION TEST"
echo "====================================================="

echo "Files created in workspace:"
ls -la "$SANDBOX_DIR" | grep -v "^total\|agentry\|\.env\|\.agentry\|templates"

echo ""
echo "Checking for fibonacci.py:"
if [ -f "$SANDBOX_DIR/fibonacci.py" ]; then
    echo "✅ fibonacci.py was created!"
    echo "File size: $(stat -c%s "$SANDBOX_DIR/fibonacci.py") bytes"
    echo "First 10 lines:"
    head -10 "$SANDBOX_DIR/fibonacci.py"
    echo "..."
    echo ""
    echo "SUCCESS: Agent 0 successfully delegated to coder agent!"
else
    echo "❌ fibonacci.py was not created"
    echo "ISSUE: Either delegation failed or coder agent didn't execute"
fi

echo ""
echo "KEY INDICATORS:"
echo "- '>>> [COORDINATION]' = Agent 0 used team management tools"
echo "- '>>> [DELEGATION]' = Agent 0 delegated to another agent"
echo "- '>>> [ERROR]' = Agent 0 tried to use forbidden tools"
echo "- '>>> [TASK ASSIGNED]' = Delegation was successful"
echo ""
echo "This test forces Agent 0 to delegate by removing its direct implementation tools."
