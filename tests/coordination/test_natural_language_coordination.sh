#!/bin/bash

# Natural Language Coordination Test script for Agentry
# Tests Agent 0's ability to spawn and coordinate agents through natural language
# Focus: NO CLI commands, pure natural language requests

echo "🌟 Testing Agentry Natural Language Team Coordination"
echo "====================================================="
echo "🎯 Goal: Agent 0 should spawn and coordinate agents through natural language only"
echo "❌ NO CLI commands like /spawn, /list - these will be deprecated"
echo "✅ PURE natural language like 'I need a coder to help with X'"
echo ""

# Create a test workspace
mkdir -p /tmp/agentry-nlc-test
cd /tmp/agentry-nlc-test

# Clean up any previous test files
rm -f *.txt *.py *.js *.md .env.local /tmp/nlc_input.txt /tmp/nlc_output_*.txt

# Copy the .env.local file to the test workspace so environment loading works
if [ -f "/home/marco/Documents/GitHub/agentry/.env.local" ]; then
    cp "/home/marco/Documents/GitHub/agentry/.env.local" .
    echo "📋 Copied .env.local to test workspace for API key access"
else
    echo "⚠️  No .env.local found in project directory"
fi

echo "📁 Test workspace: $(pwd)"
echo ""

# Function to run a single natural language test
run_nlc_test() {
    local test_num=$1
    local test_desc="$2" 
    local natural_request="$3"
    local timeout_duration=${4:-120}
    
    echo "=========================================="
    echo "🌟 Natural Language Test $test_num: $test_desc"
    echo "🗣️  Request: $natural_request"
    echo "=========================================="
    echo ""
    
    # Give user time to see what's about to happen
    echo "⏳ Starting natural language test in 3 seconds..."
    sleep 3
    
    # Create a temporary file for the command
    echo "$natural_request" > /tmp/nlc_input.txt
    echo "/quit" >> /tmp/nlc_input.txt
    
    # Run the command with input from file, capturing output for verification
    echo "🤖 Sending natural language request to Agent 0..."
    local output_file="/tmp/nlc_output_${test_num}.txt"
    timeout $timeout_duration /home/marco/Documents/GitHub/agentry/agentry.exe chat < /tmp/nlc_input.txt > "$output_file" 2>&1
    local exit_code=$?
    
    echo ""
    echo "📋 Agent 0 Response to Natural Language Request:"
    echo "------------------------------------------------"
    cat "$output_file"
    echo "------------------------------------------------"
    echo "Exit code: $exit_code"
    
    # Store output for later analysis
    echo "Test $test_num output stored in: $output_file"
    
    echo ""
    echo "⏸️  Natural language test $test_num completed. Waiting 5 seconds before next test..."
    echo ""
    sleep 5
}

# Function to verify natural language test results
verify_nlc_test() {
    local test_num=$1
    local expected_behavior="$2"
    local output_file="/tmp/nlc_output_${test_num}.txt"
    
    if [ -f "$output_file" ]; then
        # Check for explicit errors first
        if grep -q "❌ Error:" "$output_file"; then
            echo "❌ NLC Test $test_num FAILED: Found error in output"
            echo "   Error details:"
            grep "❌ Error:" "$output_file" | sed 's/^/   /'
            return 1
        fi
        
        # Check for agent spawning behavior
        if grep -q "Agent.*spawned\|✅.*spawned\|registered with team" "$output_file"; then
            echo "✅ NLC Test $test_num PASSED: Agent 0 spawned agents naturally"
            return 0
        fi
        
        # Check for intelligent analysis
        if grep -q "I'll.*agent\|need.*agent\|spawn.*agent\|create.*agent" "$output_file"; then
            echo "✅ NLC Test $test_num PASSED: Agent 0 shows spawning intent"
            return 0
        fi
        
        # Check for expected behavior patterns
        if [ -n "$expected_behavior" ] && grep -q "$expected_behavior" "$output_file"; then
            echo "✅ NLC Test $test_num PASSED: Found expected behavior '$expected_behavior'"
            return 0
        fi
        
        echo "⚠️  NLC Test $test_num UNCLEAR: Agent 0 responded but unclear if it would spawn agents"
        echo "   Response preview:"
        tail -3 "$output_file" | sed 's/^/   /'
        return 2
    else
        echo "❌ NLC Test $test_num FAILED: No output file found"
        return 1
    fi
}

echo "🚀 Starting Natural Language Coordination Tests"
echo "==============================================="

# Test 1: Natural language request for code help
run_nlc_test 1 "Request coding help naturally" \
    "I need help creating a Python script that processes CSV files. Can someone help me with this?" \
    120

verify_nlc_test 1 "python\|csv\|script"

# Test 2: Natural language request for multiple agents
run_nlc_test 2 "Request multiple specialists" \
    "I'm working on a web project and need both a coder to write JavaScript and someone to help with documentation. Can you get the right people to help?" \
    120

verify_nlc_test 2 "javascript\|documentation\|coder\|writer"

# Test 3: Natural language project analysis request 
run_nlc_test 3 "Request project analysis and coordination" \
    "I have a complex Go project that needs code review and testing. What kind of team should we assemble for this?" \
    120

verify_nlc_test 3 "go\|review\|testing\|team"

# Test 4: Natural language workflow request
run_nlc_test 4 "Request end-to-end workflow" \
    "I want to create a new API endpoint, write tests for it, and document it. How should we approach this and who should be involved?" \
    120

verify_nlc_test 4 "api\|tests\|document\|approach"

# Test 5: Natural language delegation
run_nlc_test 5 "Request specific task delegation" \
    "There's a bug in the authentication module that needs investigating. I need someone technical to look at it and propose a fix." \
    120

verify_nlc_test 5 "bug\|authentication\|investigating\|technical"

# Test 6: Natural language resource coordination
run_nlc_test 6 "Request coordinated resource management" \
    "We have multiple configuration files that need updating and I want to make sure we don't have conflicts. Can you coordinate this work?" \
    120

verify_nlc_test 6 "configuration\|updating\|conflicts\|coordinate"

# Show final workspace contents
echo "=========================================="
echo "📂 Final workspace contents:"
ls -la
echo ""

# Cleanup
rm -f /tmp/nlc_input.txt

echo "=========================================="
echo "🔍 NATURAL LANGUAGE COORDINATION RESULTS"
echo "=========================================="

# Count results
passed=0
unclear=0
failed=0
declare -a nlc_test_results

for i in {1..6}; do
    if [ -f "/tmp/nlc_output_${i}.txt" ]; then
        echo "📄 NLC Test $i output file exists"
        # Run verification and capture return code
        verify_nlc_test $i > /tmp/verify_result_${i}.txt 2>&1
        verify_code=$?
        case $verify_code in
            0) nlc_test_results[$i]="PASSED"; ((passed++)) ;;
            1) nlc_test_results[$i]="FAILED"; ((failed++)) ;;
            2) nlc_test_results[$i]="UNCLEAR"; ((unclear++)) ;;
        esac
    else
        echo "❌ NLC Test $i output file missing"
        nlc_test_results[$i]="MISSING"
        ((failed++))
    fi
done

echo ""
echo "📊 Natural Language Coordination Results:"
for i in {1..6}; do
    case ${nlc_test_results[$i]} in
        "PASSED") echo "  ✅ NLC Test $i: PASSED - Agent 0 showed natural coordination behavior" ;;
        "UNCLEAR") echo "  ⚠️  NLC Test $i: UNCLEAR - Agent 0 responded but coordination intent unclear" ;;
        "FAILED") echo "  ❌ NLC Test $i: FAILED - No appropriate response or error occurred" ;;
        "MISSING") echo "  ❓ NLC Test $i: MISSING - No output captured" ;;
    esac
done

echo ""
echo "📊 Summary: $passed clearly passed, $unclear unclear behavior, $failed failed"

# Analysis of coordination patterns
echo ""
echo "🔍 Natural Language Coordination Analysis:"
echo "==========================================="

# Check if Agent 0 is trying to spawn agents naturally
agent_spawn_attempts=0
for i in {1..6}; do
    if [ -f "/tmp/nlc_output_${i}.txt" ]; then
        if grep -q "spawn\|create.*agent\|need.*agent\|get.*agent" "/tmp/nlc_output_${i}.txt"; then
            ((agent_spawn_attempts++))
        fi
    fi
done

echo "🤖 Agent spawning attempts: $agent_spawn_attempts/6 tests"

# Check for coordination language
coordination_language=0
for i in {1..6}; do
    if [ -f "/tmp/nlc_output_${i}.txt" ]; then
        if grep -q "coordinate\|team\|together\|collaborate" "/tmp/nlc_output_${i}.txt"; then
            ((coordination_language++))
        fi
    fi
done

echo "🤝 Coordination language usage: $coordination_language/6 tests"

# Check for task understanding
task_understanding=0
for i in {1..6}; do
    if [ -f "/tmp/nlc_output_${i}.txt" ]; then
        if grep -q "understand\|need\|require\|help.*with" "/tmp/nlc_output_${i}.txt"; then
            ((task_understanding++))
        fi
    fi
done

echo "🎯 Task understanding: $task_understanding/6 tests"

echo ""
echo "💡 Key Insights:"
if [ $agent_spawn_attempts -ge 3 ]; then
    echo "  ✅ Agent 0 shows good natural language agent spawning behavior"
else
    echo "  ⚠️  Agent 0 may need better natural language to agent spawning mapping"
fi

if [ $coordination_language -ge 3 ]; then
    echo "  ✅ Agent 0 uses appropriate coordination language"
else
    echo "  ⚠️  Agent 0 coordination language could be improved"
fi

if [ $task_understanding -ge 4 ]; then
    echo "  ✅ Agent 0 shows good task understanding"
else
    echo "  ⚠️  Agent 0 task understanding needs improvement"
fi

echo ""
echo "🎯 Next Steps Based on Results:"
if [ $passed -ge 4 ]; then
    echo "  🌟 Natural language coordination is working well!"
    echo "  📈 Focus on refining coordination protocols and testing complex scenarios"
else
    echo "  🔧 Natural language coordination needs improvement"
    echo "  📝 Consider enhancing Agent 0 prompt to better map natural language to actions"
fi

echo ""
echo "✅ Natural Language Coordination testing completed"
echo "📋 All outputs captured for analysis and refinement"
