#!/bin/bash

# Agent 0 Team System Testing with Verbose Logging
# Tests the delegation workflow: team_status → check_agent → agent delegation

echo "🧪 AGENT 0 TEAM SYSTEM TESTING"
echo "=============================="
echo "Testing Agent 0's delegation capabilities with verbose logging"
echo "Expected: Agent 0 should use coordination tools, then delegate to specialist agents"
echo

cd /home/marco/Documents/GitHub/agentry

# Ensure binary is built
if [ ! -f "./agentry.exe" ]; then
    echo "Building agentry binary..."
    make build
fi

echo "🔍 TEST 1: Agent 0 Team Status Discovery"
echo "Expected: Agent 0 should use team_status tool to discover available agents"
echo "Command: team_status"
echo "----------------------------------------"

timeout 30s ./agentry.exe "Use the team_status tool to show me what agents are available" 2>&1 | tee test_output_1.log
TEST1_EXIT=$?

echo
echo "Exit code: $TEST1_EXIT"
echo "Full output saved to: test_output_1.log"
echo

echo "🔍 TEST 2: Agent 0 Agent Discovery"
echo "Expected: Agent 0 should use check_agent tool to verify specific agents exist"
echo "Command: check_agent"
echo "----------------------------------------"

timeout 30s ./agentry.exe "Use the check_agent tool to verify if the 'coder' agent exists and is available" 2>&1 | tee test_output_2.log
TEST2_EXIT=$?

echo
echo "Exit code: $TEST2_EXIT"
echo "Full output saved to: test_output_2.log"
echo

echo "🔍 TEST 3: Agent 0 Task Delegation"
echo "Expected: Agent 0 should delegate file creation to 'coder' agent"
echo "Command: delegate via agent tool"
echo "----------------------------------------"

timeout 45s ./agentry.exe "Create a simple hello.py file that prints 'Hello World'. Use your coordination workflow: check team status, verify coder agent exists, then delegate the task." 2>&1 | tee test_output_3.log
TEST3_EXIT=$?

echo
echo "Exit code: $TEST3_EXIT"
echo "Full output saved to: test_output_3.log"
echo

echo "🔍 TEST 4: Agent 0 Tool Access Verification"
echo "Expected: Agent 0 should only show coordination tools, no implementation tools"
echo "----------------------------------------"

timeout 20s ./agentry.exe "List all your available tools with their exact names and purposes. Be comprehensive." 2>&1 | tee test_output_4.log
TEST4_EXIT=$?

echo
echo "Exit code: $TEST4_EXIT"  
echo "Full output saved to: test_output_4.log"
echo

echo "📊 TEST SUMMARY"
echo "==============="
echo "Test 1 (team_status): Exit code $TEST1_EXIT"
echo "Test 2 (check_agent): Exit code $TEST2_EXIT"
echo "Test 3 (delegation): Exit code $TEST3_EXIT"
echo "Test 4 (tool verification): Exit code $TEST4_EXIT"
echo

echo "📁 LOG FILES CREATED:"
echo "- test_output_1.log (team_status test)"
echo "- test_output_2.log (check_agent test)"
echo "- test_output_3.log (delegation test)"
echo "- test_output_4.log (tool verification test)"
echo

echo "🔍 ANALYSIS HINTS:"
echo "- Look for 'Using tool:' messages to see which tools Agent 0 uses"
echo "- Check for 'team not found in context' errors"
echo "- Verify Agent 0 attempts coordination before implementation"
echo "- Confirm Agent 0 has only 10 tools, not 15"

echo
echo "Run 'cat test_output_*.log' to review detailed logs"
