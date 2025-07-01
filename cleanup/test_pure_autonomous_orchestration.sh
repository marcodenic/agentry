#!/bin/bash
set -e

# Pure Autonomous Orchestration Test
# Agent 0 receives ONLY a task description and must:
# 1. Discover what agents are available
# 2. Decide which agents to use
# 3. Delegate and coordinate autonomously

echo "================================"
echo "PURE AUTONOMOUS ORCHESTRATION TEST"
echo "================================"

# Clean slate
SANDBOX_DIR="/tmp/agentry-pure-autonomous-test"
rm -rf "$SANDBOX_DIR"
mkdir -p "$SANDBOX_DIR"
cd "$SANDBOX_DIR"

# Copy necessary configuration files
cp /home/marco/Documents/GitHub/agentry/agentry.exe ./agentry
cp /home/marco/Documents/GitHub/agentry/.env.local .
cp /home/marco/Documents/GitHub/agentry/.agentry.yaml .

echo "Test workspace: $SANDBOX_DIR"

# Create a simple task that requires multiple agent types
echo "Task: Create a simple Python web scraper that fetches quotes from a website, stores them in a JSON file, and includes proper error handling and logging."

echo ""
echo "Starting autonomous orchestration - Agent 0 will receive ONLY the task description..."
echo ""

# Create input for the agent
echo "Create a simple Python web scraper that fetches quotes from a website, stores them in a JSON file, and includes proper error handling and logging. The scraper should be well-structured with proper documentation." > /tmp/autonomous_input.txt
echo "/quit" >> /tmp/autonomous_input.txt

# Run with team orchestration
echo "Running Agent 0 in autonomous mode..."
timeout 180s ./agentry chat < /tmp/autonomous_input.txt > /tmp/autonomous_output.txt 2>&1 \
  || echo "Test completed (timeout or finished)"

echo ""
echo "=== AGENT 0 RESPONSE ==="
cat /tmp/autonomous_output.txt
echo ""
echo "========================="

echo ""
echo "=== ANALYSIS ==="
echo "Files created:"
ls -la "$SANDBOX_DIR" 2>/dev/null || echo "No files found"

if [ -f "$SANDBOX_DIR"/*.py ]; then
    echo "Python files found - checking content:"
    for file in "$SANDBOX_DIR"/*.py; do
        echo "=== $file ==="
        head -20 "$file"
        echo "..."
    done
fi

echo ""
echo "Test completed. Check the trace output above to see:"
echo "1. Did Agent 0 discover available agents autonomously?"
echo "2. Which agents did it choose to delegate to?"
echo "3. Did those agents actually execute their assigned tasks?"

echo ""
echo "Available agents that Agent 0 should have discovered:"
ls /home/marco/Documents/GitHub/agentry/templates/roles/*.yaml | sed 's/.*\///g' | sed 's/\.yaml//g'
