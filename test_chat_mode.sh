#!/bin/bash

# Sequential Test script for Agentry CLI chat mode with proper timing
# Tests basic team coordination functionality one at a time to avoid LLM flooding

echo "🧪 Testing Agentry CLI Chat Mode - Sequential Real LLM Tests"
echo "============================================================"

# Create a test workspace
mkdir -p /tmp/agentry-test-workspace
cd /tmp/agentry-test-workspace

# Clean up any previous test files
rm -f agent_test_file.txt coder_test.txt .env.local /tmp/test_input.txt /tmp/test_output_*.txt

# Copy the .env.local file to the test workspace so environment loading works
if [ -f "/home/marco/Documents/GitHub/agentry/.env.local" ]; then
    cp "/home/marco/Documents/GitHub/agentry/.env.local" .
    echo "📋 Copied .env.local to test workspace for API key access"
else
    echo "⚠️  No .env.local found in project directory"
fi

echo "📁 Test workspace: $(pwd)"
echo ""

# Function to run a single test with proper timing
run_test() {
    local test_num=$1
    local test_desc=$2
    local question=$3
    local timeout_duration=${4:-90}
    
    echo "=========================================="
    echo "Test $test_num: $test_desc"
    echo "Question: $question"
    echo "=========================================="
    echo ""
    
    # Give user time to see what's about to happen
    echo "⏳ Starting test in 3 seconds..."
    sleep 3
    
    # Create a temporary file for the command
    echo "$question" > /tmp/test_input.txt
    echo "/quit" >> /tmp/test_input.txt
    
    # Run the command with input from file, capturing output for verification
    echo "🤖 Sending to Agent 0..."
    local output_file="/tmp/test_output_${test_num}.txt"
    timeout $timeout_duration /home/marco/Documents/GitHub/agentry/agentry.exe chat < /tmp/test_input.txt > "$output_file" 2>&1
    local exit_code=$?
    
    echo ""
    echo "📋 Response from Agent 0:"
    echo "------------------------"
    cat "$output_file"
    echo "------------------------"
    echo "Exit code: $exit_code"
    
    # Store output for later analysis
    echo "Test $test_num output stored in: $output_file"
    
    echo ""
    echo "⏸️  Test $test_num completed. Waiting 5 seconds before next test..."
    echo ""
    sleep 5
}

# Function to verify test results with proper success/failure detection
verify_test() {
    local test_num=$1
    local expected_success="$2"
    local expected_failure="$3"
    local output_file="/tmp/test_output_${test_num}.txt"
    
    if [ -f "$output_file" ]; then
        # Check for explicit errors first
        if grep -q "❌ Error:" "$output_file"; then
            echo "❌ Test $test_num FAILED: Found error in output"
            echo "   Error details:"
            grep "❌ Error:" "$output_file" | sed 's/^/   /'
            return 1
        fi
        
        # Check for expected success patterns
        if [ -n "$expected_success" ] && grep -q "$expected_success" "$output_file"; then
            echo "✅ Test $test_num PASSED: Found success pattern '$expected_success'"
            return 0
        fi
        
        # Check for expected failure patterns (some tests expect specific failures)
        if [ -n "$expected_failure" ] && grep -q "$expected_failure" "$output_file"; then
            echo "⚠️  Test $test_num EXPECTED FAILURE: Found expected pattern '$expected_failure'"
            return 0
        fi
        
        echo "❌ Test $test_num FAILED: No success pattern found"
        echo "   Output preview:"
        tail -5 "$output_file" | sed 's/^/   /'
        return 1
    else
        echo "❌ Test $test_num FAILED: No output file found"
        return 1
    fi
}

# Test 1: Basic system agent self-awareness
run_test 1 "Agent 0 self-awareness and team status" \
    "What are your capabilities as Agent 0? What tools do you have for team coordination?" \
    120

verify_test 1 "capabilities.*File Management\|delegation.*agents\|System Operations" ""

# Test 2: Team status check
run_test 2 "Check current team status" \
    "What is the current team status? Are there any other agents running?" \
    60

verify_test 2 "only agent\|no other agents\|currently active" ""

# Test 3: Simple file creation task
run_test 3 "Simple file creation task" \
    "Create a file called agent_test_file.txt with the content 'Hello from Agent 0'" \
    90

verify_test 3 "🤖 system:" ""

# Check if file was created
echo "📋 Checking if file was created..."
if [ -f "agent_test_file.txt" ]; then
    echo "✅ File created successfully: $(cat agent_test_file.txt)"
else
    echo "❌ File was not created"
fi
echo ""
sleep 2

# Test 4: Agent spawning request - test direct CLI command
run_test 4 "Request to spawn a coder agent using CLI command" \
    "/spawn coder 'Help with coding tasks'" \
    120

verify_test 4 "✅ Agent.*spawned successfully\|Agent.*registered with team" ""

# Test 5: List agents to see if spawning worked - test direct CLI command  
run_test 5 "List current agents using CLI command" \
    "/list" \
    60

verify_test 5 "📋 Active agents:\|system.*coder" ""

# Test 6: Workspace analysis - better prompt for tool usage
run_test 6 "Workspace analysis with proper tool usage" \
    "Use the listfiles tool to show what files are in the current directory" \
    90

verify_test 6 "🤖 system:" ""

# Show actual directory contents
echo "=========================================="
echo "📂 Final workspace contents:"
ls -la
echo ""

# Cleanup
rm -f /tmp/test_input.txt

echo "=========================================="
echo "🔍 TEST VERIFICATION SUMMARY"
echo "=========================================="

# Count pass/fail results properly
passed=0
failed=0
declare -a test_results

for i in {1..6}; do
    if [ -f "/tmp/test_output_${i}.txt" ]; then
        echo "📄 Test $i output file exists"
        if grep -q "❌ Error:" "/tmp/test_output_${i}.txt"; then
            test_results[$i]="FAILED"
            ((failed++))
        else
            test_results[$i]="PASSED"
            ((passed++))
        fi
    else
        echo "❌ Test $i output file missing"
        test_results[$i]="MISSING"
        ((failed++))
    fi
done

echo ""
echo "📊 Detailed Results:"
for i in {1..6}; do
    case ${test_results[$i]} in
        "PASSED") echo "  ✅ Test $i: PASSED" ;;
        "FAILED") echo "  ❌ Test $i: FAILED" ;;
        "MISSING") echo "  ❓ Test $i: MISSING" ;;
    esac
done

echo ""
echo "📊 Summary: $passed tests passed, $failed tests failed"

# Show critical issues found
echo ""
echo "🔍 Critical Issues Identified:"
if grep -q "invalid agent name '/list'" /tmp/test_output_5.txt 2>/dev/null; then
    echo "  ⚠️  AI doesn't understand CLI commands (tried to delegate '/list')"
fi
if grep -q "file.*already exists" /tmp/test_output_3.txt 2>/dev/null; then
    echo "  ⚠️  File creation failed due to existing file (need overwrite handling)"
fi
if grep -q "failed to read file.*is a directory" /tmp/test_output_6.txt 2>/dev/null; then
    echo "  ⚠️  AI using wrong tool for directory analysis (fileinfo vs listfiles)"
fi

echo ""

echo "✅ Sequential CLI Chat Mode tests completed"
echo ""
echo "🔍 Summary of what we tested:"
echo "- Agent 0 self-awareness and capabilities"
echo "- Team status checking"
echo "- Direct file creation by Agent 0"
echo "- Agent spawning requests"
echo "- Agent listing"
echo "- Workspace analysis"
echo ""
echo "💡 Each test was run sequentially with proper timing to avoid LLM flooding"
echo "📋 All outputs captured for verification and analysis"
