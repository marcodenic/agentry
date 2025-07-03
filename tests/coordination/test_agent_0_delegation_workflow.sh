#!/bin/bash

# Agent 0 Team System Testing with Verbose Logging
# Tests the delegation workflow: team_status → check_agent → agent delegation

# Source the test helpers script
# shellcheck source=/dev/null
source "$(dirname "$0")/../scripts/test-helpers.sh"

echo "🧪 AGENT 0 TEAM SYSTEM TESTING"
echo "=============================="
echo "Testing Agent 0's delegation capabilities with verbose logging"
echo "Expected: Agent 0 should use coordination tools, then delegate to specialist agents"
echo

# Ensure binary and config are in the temp directory
setup_test_environment

echo "🔍 TEST 1: Agent 0 Team Status Discovery"
echo "Expected: Agent 0 should use team_status tool to discover available agents"
echo "Command: team_status"
echo "----------------------------------------"

$AGENT_CMD "Use the team_status tool to show me what agents are available"
TEST1_EXIT=$?

echo
echo "Exit code: $TEST1_EXIT"
echo

echo "🔍 TEST 2: Agent 0 Agent Discovery"
echo "Expected: Agent 0 should use check_agent tool to verify specific agents exist"
echo "Command: check_agent"
echo "----------------------------------------"

$AGENT_CMD "Use the check_agent tool to verify if the 'coder' agent exists and is available"
TEST2_EXIT=$?

echo
echo "Exit code: $TEST2_EXIT"
echo

echo "🔍 TEST 3: Agent 0 Task Delegation"
echo "Expected: Agent 0 should delegate file creation to 'coder' agent"
echo "Command: delegate via agent tool"
echo "----------------------------------------"

$AGENT_CMD "Create a simple hello.py file that prints 'Hello World'. Use your coordination workflow: check team status, verify coder agent exists, then delegate the task."
TEST3_EXIT=$?

echo
echo "Exit code: $TEST3_EXIT"
echo

echo "🔍 TEST 4: Agent 0 Tool Access Verification"
echo "Expected: Agent 0 should only show coordination tools, no implementation tools"
echo "----------------------------------------"

$AGENT_CMD "List all your available tools with their exact names and purposes. Be comprehensive."
TEST4_EXIT=$?

echo
echo "Exit code: $TEST4_EXIT"  
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
