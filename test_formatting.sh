#!/bin/bash

# Test script to verify TUI formatting improvements
echo "Testing Agentry TUI formatting..."

# Test 1: Basic delegation
echo "Testing: ask coder to review the readme.md and report back" | timeout 30s ./agentry.exe smart-config.yaml

echo -e "\n\n=========================="
echo "Test completed. The TUI should now have:"
echo "1. Better spacing between status messages and responses"
echo "2. Agent 0 should have injected status (reducing check_agent calls)"
echo "3. Clean output without debug messages in TUI mode"
echo "=========================="
