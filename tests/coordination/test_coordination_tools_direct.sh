#!/bin/bash
set -e

# Test Agent 0's Coordination Tools Directly
echo "=================================================="
echo "DIRECT COORDINATION TOOLS TEST"
echo "=================================================="

# Clean slate
SANDBOX_DIR="/tmp/agentry-coordination-tools-test"
rm -rf "$SANDBOX_DIR"
mkdir -p "$SANDBOX_DIR"
cd "$SANDBOX_DIR"

# Copy necessary configuration files and updated agent role
cp /home/marco/Documents/GitHub/agentry/agentry.exe ./agentry
cp /home/marco/Documents/GitHub/agentry/.env.local .
cp /home/marco/Documents/GitHub/agentry/.agentry.yaml .

# Copy the templates directory
mkdir -p templates/roles
cp -r /home/marco/Documents/GitHub/agentry/templates/roles/* templates/roles/

echo "Test workspace: $SANDBOX_DIR"
echo "Testing if Agent 0's coordination tools work at all..."
echo "=================================================="

# Create input that FORCES Agent 0 to use coordination tools
cat > /tmp/coordination_tools_input.txt << 'EOF'
Before doing anything else, I want you to demonstrate your coordination capabilities by doing these exact steps:

1. Use the team_status tool to check what agents are available
2. Use the check_agent tool to check if the "coder" agent exists
3. Use the check_agent tool to check if the "tester" agent exists  
4. Use the assign_task tool to assign a simple task to the coder agent
5. Show me the results of each step

Do not create any files or implement anything - just demonstrate your coordination tools work.
/quit
EOF

echo "EXPLICIT COORDINATION TOOLS TEST:"
echo "Forcing Agent 0 to use team_status, check_agent, and assign_task"
echo "=================================================="

# Run with enhanced output
timeout 120s stdbuf -oL -eL ./agentry chat < /tmp/coordination_tools_input.txt 2>&1 | \
while IFS= read -r line; do
    echo "[$(date '+%H:%M:%S')] $line"
    
    # Highlight coordination activities
    case "$line" in
        *"🔧"*"team_status"*)
            echo ">>> [SUCCESS] Agent 0 used team_status tool!"
            ;;
        *"🔧"*"check_agent"*)
            echo ">>> [SUCCESS] Agent 0 used check_agent tool!"
            ;;
        *"🔧"*"assign_task"*)
            echo ">>> [SUCCESS] Agent 0 used assign_task tool!"
            ;;
        *"🔧"*"send_message"*)
            echo ">>> [SUCCESS] Agent 0 used send_message tool!"
            ;;
        *"available"*|*"not found"*|*"not available"*)
            if [[ "$line" == *"🔧"* ]]; then
                echo ">>> [TOOL RESULT] $line"
            fi
            ;;
        *"Task assigned"*)
            echo ">>> [TASK ASSIGNED] $line"
            ;;
    esac
done

echo ""
echo "=================================================="
echo "COORDINATION TOOLS TEST RESULTS"
echo "=================================================="
echo "If you see '>>> [SUCCESS]' messages above, the coordination tools work"
echo "If you see '>>> [TOOL RESULT]' messages, we can see what agents are available"
echo "If Agent 0 refuses to use tools, there's a deeper problem with the prompt"
echo ""
echo "This test directly commands Agent 0 to use its coordination tools."
