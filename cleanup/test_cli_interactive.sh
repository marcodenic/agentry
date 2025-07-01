#!/bin/bash

# Interactive test script for Agentry CLI chat mode with real LLM responses
# This script simulates user interactions step by step with proper timing

echo "🧪 Testing Agentry CLI Chat Mode - Interactive LLM Tests"
echo "========================================================"

# Create a test workspace
mkdir -p /tmp/agentry-interactive-test
cd /tmp/agentry-interactive-test

echo "📁 Test workspace: $(pwd)"
echo ""

# Function to send command and wait for response
send_command() {
    local cmd="$1"
    local timeout_sec="$2"
    local description="$3"
    
    echo "Test: $description"
    echo "Command: $cmd"
    echo "----------------------------------------"
    
    # Create a temporary expect-like script
    cat > test_command.sh << EOF
#!/bin/bash
exec 3< <(echo "$cmd" | /home/marco/Documents/GitHub/agentry/agentry.exe chat 2>&1)
timeout $timeout_sec cat <&3
exec 3<&-
EOF
    
    chmod +x test_command.sh
    ./test_command.sh
    echo ""
    echo "----------------------------------------"
    echo ""
}

# Test 1: Basic startup and help
send_command "/help" 10 "Show help and available commands"

# Test 2: Team status check
send_command "/status" 15 "Check initial team status"

# Test 3: Test agent capabilities question
send_command "What are your capabilities as Agent 0?" 30 "Agent 0 self-awareness"

# Test 4: Simple file creation
send_command "Create a file called hello.txt with content 'Hello World'" 30 "Direct file creation"

# Check if file was created
echo "File creation check:"
if [ -f "hello.txt" ]; then
    echo "✅ File created: $(cat hello.txt)"
else
    echo "❌ File not found"
fi
echo ""

# Test 5: Team status after potential file creation
send_command "/status" 15 "Team status after file task"

# Test 6: Spawn agent request
send_command "Please spawn a coder agent to help with development tasks" 30 "Request to spawn coder agent"

# Test 7: List agents after spawn attempt
send_command "/list" 10 "List agents after spawn attempt"

# Cleanup
rm -f test_command.sh

echo "✅ Interactive CLI Chat Mode tests completed"
echo ""
echo "📂 Final directory contents:"
ls -la
echo ""
