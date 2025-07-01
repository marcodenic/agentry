#!/bin/bash

# Simple validation test for Agent 0's coordination capabilities
echo "=== AGENT 0 VALIDATION TEST ==="
echo "Testing if Agent 0 properly uses coordination tools"
echo

cd /home/marco/Documents/GitHub/agentry

echo "1. Testing Agent 0's available tools:"
timeout 15s ./agentry.exe "List your available tools briefly" 2>/dev/null

echo
echo "2. Testing team_status tool:"
timeout 15s ./agentry.exe "Use the team_status tool to check available agents" 2>/dev/null

echo
echo "3. Testing check_agent tool:"
timeout 15s ./agentry.exe "Use check_agent to see if 'coder' agent exists" 2>/dev/null

echo
echo "=== TEST COMPLETE ==="
