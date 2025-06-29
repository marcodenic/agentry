#!/bin/bash

# Interactive Test Session for Agentry CLI Chat Mode
# This script runs an interactive session where each command is sent individually
# and we wait for proper completion before sending the next one

echo "🧪 Agentry CLI Interactive Test Session"
echo "======================================="

# Create a test workspace
mkdir -p /tmp/agentry-test-workspace
cd /tmp/agentry-test-workspace

echo "📁 Test workspace: $(pwd)"
echo ""

# Function to send command and wait for response
send_command() {
    local cmd="$1"
    local description="$2"
    local timeout_duration="${3:-120}"
    
    echo "=========================================="
    echo "📤 Sending: $description"
    echo "Command: $cmd"
    echo "=========================================="
    
    # Create a named pipe for communication
    local pipe_name="/tmp/agentry_test_$$"
    mkfifo "$pipe_name"
    
    # Start agentry chat in background, feeding from the pipe
    timeout $timeout_duration /home/marco/Documents/GitHub/agentry/agentry.exe chat < "$pipe_name" &
    local agent_pid=$!
    
    # Send the command
    echo "$cmd" > "$pipe_name" &
    
    # Wait a bit for the response to start
    sleep 3
    
    # Check if process is still running (means it's waiting for more input)
    if kill -0 $agent_pid 2>/dev/null; then
        echo "🤖 Agent is processing... waiting for completion"
        
        # Wait for the agent to finish processing or timeout
        local count=0
        while kill -0 $agent_pid 2>/dev/null && [ $count -lt $(($timeout_duration/5)) ]; do
            sleep 5
            count=$((count + 1))
            echo "⏳ Still processing... ($((count * 5))s elapsed)"
        done
        
        # Send quit command if still running
        if kill -0 $agent_pid 2>/dev/null; then
            echo "/quit" > "$pipe_name" &
            sleep 2
        fi
        
        # Force kill if still running
        if kill -0 $agent_pid 2>/dev/null; then
            kill $agent_pid 2>/dev/null
            wait $agent_pid 2>/dev/null
        fi
    fi
    
    # Cleanup
    rm -f "$pipe_name"
    
    echo "✅ Command completed"
    echo ""
    sleep 2
}

echo "🚀 Starting interactive test session..."
echo "Each command will be sent individually with proper timing"
echo ""

# Test sequence
send_command "What are your capabilities as Agent 0? What tools do you have for team coordination?" \
    "Agent 0 self-awareness test" 90

send_command "What is the current team status? Use the team_status tool to check." \
    "Team status check" 60

send_command "Create a file called agent_test_file.txt with the content 'Hello from Agent 0 - $(date)'" \
    "File creation task" 90

# Check if file was created
echo "📋 Checking if file was created..."
if [ -f "agent_test_file.txt" ]; then
    echo "✅ File created successfully:"
    cat agent_test_file.txt
else
    echo "❌ File was not created"
fi
echo ""

send_command "/spawn coder \"Help with coding tasks\"" \
    "Spawn coder agent command" 120

send_command "/list" \
    "List all agents" 30

send_command "What files are in the current directory? Please list and analyze them." \
    "Workspace analysis" 60

send_command "Send a message to any running coder agents asking them to create a test file called 'coder_hello.txt'" \
    "Inter-agent communication test" 90

# Final workspace check
echo "=========================================="
echo "📂 Final workspace contents:"
ls -la
echo ""

echo "✅ Interactive test session completed!"
echo ""
echo "🔍 What we tested:"
echo "- Agent 0 capabilities and tool awareness"
echo "- Team status checking with proper tool usage"
echo "- Direct file creation"
echo "- Agent spawning with CLI commands"
echo "- Agent listing"
echo "- Workspace analysis"
echo "- Inter-agent communication"
echo ""
echo "💡 Each command was executed individually with proper response timing"
