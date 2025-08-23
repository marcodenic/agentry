#!/bin/bash

echo "Testing tool call argument parsing fix..."

# Test 1: Echo tool (should work if arguments are parsed)
echo "Test 1: Echo tool with arguments"
timeout 10s ./agentry "echo hello world" 2>&1 | grep -E "(args:|tool.*succeeded|text:)"

echo -e "\n---\n"

# Test 2: Check for tool call count
echo "Test 2: Tool call count (should be > 0)"
timeout 10s ./agentry "echo test" 2>&1 | grep "Streaming completed with" | head -1

echo -e "\n---\n"

echo "Fix verification complete!"
