#!/usr/bin/env bash

# Test script to verify TUI improvements
echo "Testing Agentry TUI improvements..."

# Start Agentry with a simple test
echo "Starting Agentry with test input..."

# Create a test config if it doesn't exist
if [ ! -f test-config.yaml ]; then
    cat > test-config.yaml << EOF
agents:
  - name: agent_0
    role: templates/roles/agent_0.yaml
    model: openai:gpt-4
    tools:
      - registry
      - bash
      - agent
EOF
fi

echo "Configuration ready. You can now test the TUI improvements:"
echo "1. Run: ./agentry.exe -c test-config.yaml"
echo "2. Try these commands to test formatting:"
echo "   - Type: 'Hello, can you help me with a task?'"
echo "   - Check if spinners clear properly"
echo "   - Verify token/cost tracking in footer"
echo "   - Test command grouping and spacing"
echo ""
echo "Expected improvements:"
echo "✓ Better spacing between user input and AI responses"
echo "✓ Cleaner command output formatting"
echo "✓ Spinners should clear when tokens start arriving"
echo "✓ Footer should show live token/cost updates"
echo "✓ Better visual separation between message types"
