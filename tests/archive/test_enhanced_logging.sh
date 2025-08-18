#!/bin/bash
set -e

# Test with Enhanced Real-Time Logging
echo "================================================"
echo "ENHANCED LOGGING - AGENT 0 COORDINATION TEST"
echo "================================================"

# Clean slate
SANDBOX_DIR="/tmp/agentry-enhanced-logging-test"
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
echo "================================================"

# Create input that clearly asks for file creation with real-time monitoring
cat > /tmp/enhanced_logging_input.txt << 'EOF'
I need you to implement a simple Python password generator that creates secure passwords. Please create a file called "password_generator.py" with the following features:

1. Generate passwords of specified length
2. Include uppercase, lowercase, numbers, and special characters
3. Allow user to specify password requirements
4. Save the working file to the current directory

This requires actual file creation - not just code examples.
/quit
EOF

echo "TASK: Create a Python password generator file"
echo "MONITORING: Agent 0's coordination and delegation behavior"
echo "================================================"

# Run with enhanced output - show everything in real time
echo "Starting Agent 0 with enhanced real-time logging..."
echo ""

# Use unbuffered output and show everything as it happens
timeout 180s stdbuf -oL -eL ./agentry chat < /tmp/enhanced_logging_input.txt 2>&1 | \
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
            echo ">>> [COORDINATION] Agent 0 is sending a message to agents!"
            ;;
        *"🔧"*"create"*)
            echo ">>> [DIRECT IMPLEMENTATION] Agent 0 is creating files directly!"
            ;;
        *"🔧"*"edit_range"*)
            echo ">>> [DIRECT IMPLEMENTATION] Agent 0 is editing files directly!"
            ;;
        *"🔧"*"write"*)
            echo ">>> [DIRECT IMPLEMENTATION] Agent 0 is writing files directly!"
            ;;
        *"Agent"*"spawned"*)
            echo ">>> [SUCCESS] An agent was spawned successfully!"
            ;;
        *"available"*|*"not found"*)
            if [[ "$line" == *"🔧"* ]]; then
                echo ">>> [AGENT CHECK] Agent availability result: $line"
            fi
            ;;
    esac
done

echo ""
echo "================================================"
echo "POST-TEST ANALYSIS"
echo "================================================"

echo "Files created in workspace:"
ls -la "$SANDBOX_DIR" | grep -v "^total\|agentry\|\.env\|\.agentry\|templates"

echo ""
echo "Checking for password_generator.py:"
if [ -f "$SANDBOX_DIR/password_generator.py" ]; then
    echo "✅ password_generator.py was created!"
    echo "File size: $(stat -c%s "$SANDBOX_DIR/password_generator.py") bytes"
    echo "First 10 lines:"
    head -10 "$SANDBOX_DIR/password_generator.py"
    echo "..."
    echo ""
    echo "WHO CREATED IT?"
    echo "If we see '>>> [DIRECT IMPLEMENTATION]' messages above, Agent 0 created it directly"
    echo "If we see '>>> [DELEGATION]' messages above, an agent was delegated to create it"
else
    echo "❌ password_generator.py was not created"
    echo "Check the log above to see what Agent 0 actually did"
fi

echo ""
echo "COORDINATION BEHAVIOR SUMMARY:"
echo "- Look for '>>> [COORDINATION]' messages to see team management"
echo "- Look for '>>> [DELEGATION]' messages to see task delegation" 
echo "- Look for '>>> [DIRECT IMPLEMENTATION]' to see if Agent 0 did work itself"
echo ""
echo "This enhanced logging shows exactly what Agent 0 is doing in real-time."
