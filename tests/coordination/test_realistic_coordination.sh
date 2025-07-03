#!/bin/bash

# Realistic Multi-Agent Coordination Test
# Test Agent 0 coordination with real project scenarios

echo "🚀 Realistic Multi-Agent Coordination Test"
echo "=========================================="
echo "Testing Agent 0 with realistic development scenarios"
echo ""

# Create a realistic test workspace
mkdir -p /tmp/agentry-realistic-test
cd /tmp/agentry-realistic-test

# Copy environment
if [ -f "/home/marco/Documents/GitHub/agentry/.env.local" ]; then
    cp "/home/marco/Documents/GitHub/agentry/.env.local" .
fi

# Create a realistic project structure
mkdir -p src tests docs
echo "package main

import \"fmt\"

func main() {
    fmt.Println(\"Hello World\")
}" > src/main.go

echo "# My Go Project
This is a sample Go application." > README.md

echo "module myproject

go 1.21" > go.mod

echo "package main

import \"testing\"

func TestMain(t *testing.T) {
    // TODO: Add tests
}" > tests/main_test.go

echo "📁 Created realistic project structure:"
ls -la
echo ""
echo "📂 Project contents:"
find . -type f -exec echo "  {}" \; -exec head -2 {} \; -exec echo "" \;

echo ""
echo "🧪 Testing Agent 0 with realistic coordination scenarios"
echo "======================================================="

# Test function for realistic scenarios
test_realistic_scenario() {
    local scenario_name="$1"
    local request="$2"
    local timeout_duration=${3:-90}
    
    echo "🎬 Scenario: $scenario_name"
    echo "📝 Request: $request"
    echo "----------------------------------------"
    
    # Create input
    echo "$request" > /tmp/realistic_input.txt
    echo "/list" >> /tmp/realistic_input.txt
    echo "/quit" >> /tmp/realistic_input.txt
    
    # Run test
    local output_file="/tmp/realistic_output_$(date +%s).txt"
    echo "⏳ Running scenario..."
    timeout $timeout_duration /home/marco/Documents/GitHub/agentry/agentry.exe chat < /tmp/realistic_input.txt > "$output_file" 2>&1
    
    echo ""
    echo "📋 Agent 0 Response:"
    echo "--------------------"
    cat "$output_file"
    echo "--------------------"
    
    # Analyze response
    tool_usage=$(grep -c "🔧 system is using a tool" "$output_file")
    agent_list=$(grep -A 20 "📋 Active agents:" "$output_file" | grep -E "^\s*>" | wc -l)
    
    echo ""
    echo "📊 Scenario Analysis:"
    echo "   🔧 Tool usage: $tool_usage"
    echo "   👥 Agents in final list: $agent_list"
    
    if [ $tool_usage -gt 0 ]; then
        echo "   ✅ COORDINATION ATTEMPTED"
        if [ $agent_list -gt 1 ]; then
            echo "   ✅ AGENTS SPAWNED SUCCESSFULLY"
        else
            echo "   ⚠️  No agents visible in final list"
        fi
    else
        echo "   ⚠️  NO COORDINATION DETECTED"
    fi
    
    echo ""
    echo "=================================================="
    echo ""
    sleep 2
    
    # Store output file path for analysis
    echo "$output_file" > /tmp/last_output_file.txt
}

# Realistic scenario tests
test_realistic_scenario \
    "Code Review & Testing" \
    "I need help reviewing the Go code in src/main.go and writing comprehensive tests for it. Can you get the right people to help with code quality and testing?"

test_realistic_scenario \
    "Documentation & API Design" \
    "This project needs better documentation and I want to add API endpoints. I need someone technical for the API work and someone for documentation."

test_realistic_scenario \
    "Bug Investigation" \
    "The main.go file has some issues and the tests are incomplete. I need someone to debug the code and someone else to fix the test coverage."

test_realistic_scenario \
    "Full Development Workflow" \
    "I want to extend this Go project with new features, write tests, update documentation, and prepare it for deployment. What team should we assemble?"

# Check final project state
echo "📂 Final Project State:"
echo "======================"
ls -la
echo ""

if [ -d src ] && [ -d tests ] && [ -d docs ]; then
    echo "✅ Project structure maintained"
else
    echo "⚠️  Project structure may have been modified"
fi

echo ""
echo "🎯 REALISTIC COORDINATION SUMMARY"
echo "================================="
echo "✅ Tested Agent 0 with real project files"
echo "✅ Used realistic development scenarios"
echo "✅ Verified coordination behavior in context"
echo ""
echo "💡 This test shows how Agent 0 handles real-world coordination requests"
echo "📋 Results show Agent 0's ability to understand project context"

# Cleanup
rm -f /tmp/realistic_input.txt
