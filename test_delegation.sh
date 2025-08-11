#!/bin/bash

# Test agent delegation outside of TUI
export AGENTRY_TUI_MODE=0
source .env.local

echo "Testing agent delegation..."
echo "spawn a coder and tell them to review FEATURES.md and report back" | ./agentry --config .agentry.yaml 

echo "Test completed."
