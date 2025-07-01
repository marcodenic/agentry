#!/bin/bash

# Test script to verify Agent 0's tool restrictions are working
# Agent 0 should only have coordination tools, not implementation tools

echo "=== TESTING AGENT 0 TOOL RESTRICTIONS ==="
echo "Agent 0 should NOT have create, edit_range, write tools"
echo "Agent 0 should ONLY have coordination tools"
echo

# Test 1: Agent 0 trying to create a file directly (should fail/delegate)
echo "TEST 1: Agent 0 attempting to create file directly"
echo "Expected: Agent 0 should delegate to coder, not create file itself"
echo "-----"

timeout 30s ./agentry.exe -config .agentry.yaml chat << 'EOF'
Create a simple hello.py file that prints "Hello World"
EOF

echo
echo "TEST 1 COMPLETE"
echo

# Test 2: Agent 0 showing available tools
echo "TEST 2: Agent 0 showing what tools it has access to"
echo "Expected: Should show coordination tools only"
echo "-----"

timeout 30s ./agentry.exe -config .agentry.yaml chat << 'EOF'
What tools do you have available? List them all.
EOF

echo
echo "TEST 2 COMPLETE"
echo

# Test 3: Agent 0 coordination workflow
echo "TEST 3: Agent 0 coordination workflow"
echo "Expected: Should use team_status, check_agent, then delegate"
echo "-----"

timeout 30s ./agentry.exe -config .agentry.yaml chat << 'EOF'
I need to review the main.go file. Please show me your coordination workflow.
EOF

echo
echo "TEST 3 COMPLETE"
echo

echo "=== ALL TESTS COMPLETE ==="
echo "Review output above to verify Agent 0 is properly restricted"
