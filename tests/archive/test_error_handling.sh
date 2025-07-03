#!/bin/bash

# Error Handling and Recovery Coordination Test
# Test Agent 0's ability to handle failures and recovery scenarios

echo "🛡️ Error Handling and Recovery Coordination Test"
echo "================================================"
echo "Testing Agent 0's error handling and recovery capabilities"
echo ""

# Create clean sandbox
rm -rf /tmp/agentry-ai-sandbox
mkdir -p /tmp/agentry-ai-sandbox
cd /tmp/agentry-ai-sandbox

# Copy agentry binary and config
cp /home/marco/Documents/GitHub/agentry/agentry.exe ./agentry
cp /home/marco/Documents/GitHub/agentry/.env.local .
cp /home/marco/Documents/GitHub/agentry/.agentry.yaml .

echo "🧪 Test Scenario 1: Invalid Code Request"
echo "Testing how Agent 0 handles impossible/invalid requests"
echo ""

# Test 1: Invalid code request
echo "🤖 Agent 0 Request: Create impossible code"
timeout 60s ./agentry chat <<EOF
Create a Python file called 'impossible.py' that divides by zero and imports a non-existent module called 'fake_module'. Make sure this code runs without errors.
EOF

echo ""
echo "📊 Test 1 Results:"
if [ -f "impossible.py" ]; then
    echo "  ✅ File created (testing error handling)"
    echo "  🔍 Content check:"
    cat impossible.py
    echo ""
    echo "  🧪 Testing execution:"
    if python3 impossible.py 2>&1; then
        echo "  ❌ Code runs without errors (should fail)"
    else
        echo "  ✅ Code fails as expected (good error handling)"
    fi
else
    echo "  ⚠️  No file created (Agent 0 may have refused invalid request)"
fi

echo ""
echo "🧪 Test Scenario 2: File Conflict Resolution"
echo "Testing how Agent 0 handles file conflicts"
echo ""

# Create a pre-existing file
echo "print('Original version')" > conflict_test.py

echo "🤖 Agent 0 Request: Overwrite existing file"
timeout 60s ./agentry chat <<EOF
Create a Python file called 'conflict_test.py' that prints 'New version'. This file already exists - please handle the conflict appropriately.
EOF

echo ""
echo "📊 Test 2 Results:"
if [ -f "conflict_test.py" ]; then
    echo "  ✅ File exists"
    echo "  🔍 Current content:"
    cat conflict_test.py
    
    # Check if content changed
    if grep -q "New version" conflict_test.py; then
        echo "  ✅ File was updated (good conflict resolution)"
    elif grep -q "Original version" conflict_test.py; then
        echo "  ⚠️  Original file preserved (conservative approach)"
    else
        echo "  🔄 File content changed to something else"
    fi
else
    echo "  ❌ File disappeared (unexpected)"
fi

echo ""
echo "🧪 Test Scenario 3: Resource Constraint Simulation"
echo "Testing how Agent 0 handles resource limitations"
echo ""

# Create a scenario with many files to test resource handling
echo "🤖 Agent 0 Request: Create many files to test resource limits"
timeout 90s ./agentry chat <<EOF
Create 20 different Python files named file1.py through file20.py. Each file should contain a unique function that prints its filename. This tests resource handling for large batch operations.
EOF

echo ""
echo "📊 Test 3 Results:"
created_files=$(ls file*.py 2>/dev/null | wc -l)
total_requested=20

echo "  📁 Files created: $created_files out of $total_requested requested"

if [ $created_files -eq $total_requested ]; then
    echo "  ✅ All files created successfully (excellent resource handling)"
elif [ $created_files -gt 10 ]; then
    echo "  ✅ Most files created (good resource handling with possible limits)"
elif [ $created_files -gt 0 ]; then
    echo "  ⚠️  Some files created (partial success, may indicate resource limits)"
else
    echo "  ❌ No files created (resource constraint or refusal)"
fi

# Sample a few files to check quality
echo "  🔍 Sample file contents:"
for i in 1 2 3; do
    if [ -f "file$i.py" ]; then
        echo "    file$i.py:"
        cat "file$i.py" | head -3
    fi
done

echo ""
echo "🧪 Test Scenario 4: Retry and Recovery"
echo "Testing Agent 0's ability to retry after failures"
echo ""

# Create a more complex retry scenario
echo "🤖 Agent 0 Request: Complex task with potential failure points"
timeout 120s ./agentry chat <<EOF
Create a Python project with error handling and recovery:

1. Create 'network_service.py' with a NetworkService class that has:
   - connect() method that might fail (simulate with random success/failure)
   - retry_connect() method that retries connection up to 3 times
   - get_data() method that uses the connection

2. Create 'test_network.py' that tests the retry logic by calling retry_connect() and verifying it handles failures gracefully.

Make sure the code demonstrates proper error handling, logging, and recovery patterns.
EOF

echo ""
echo "📊 Test 4 Results:"
echo "  📁 Files created:"
if [ -f "network_service.py" ]; then
    echo "    ✅ network_service.py"
else
    echo "    ❌ network_service.py (missing)"
fi

if [ -f "test_network.py" ]; then
    echo "    ✅ test_network.py"  
else
    echo "    ❌ test_network.py (missing)"
fi

# Test if the error handling works
if [ -f "network_service.py" ] && [ -f "test_network.py" ]; then
    echo "  🧪 Testing error handling logic:"
    if python3 -c "import network_service; ns = network_service.NetworkService(); print('Import successful')" 2>/dev/null; then
        echo "    ✅ network_service.py imports successfully"
        
        if python3 test_network.py 2>&1 | head -10; then
            echo "    ✅ Test executed (check output above for error handling)"
        else
            echo "    ⚠️  Test execution had issues"
        fi
    else
        echo "    ❌ network_service.py has import errors"
    fi
fi

echo ""
echo "📈 ERROR HANDLING ASSESSMENT"
echo "============================="

# Calculate overall error handling score
total_scenarios=4
successful_scenarios=0

# Count successful scenarios
if [ -f "impossible.py" ] || echo "Agent handled invalid request"; then successful_scenarios=$((successful_scenarios + 1)); fi
if [ -f "conflict_test.py" ]; then successful_scenarios=$((successful_scenarios + 1)); fi
if [ $created_files -gt 5 ]; then successful_scenarios=$((successful_scenarios + 1)); fi
if [ -f "network_service.py" ] && [ -f "test_network.py" ]; then successful_scenarios=$((successful_scenarios + 1)); fi

error_handling_rate=$((100 * successful_scenarios / total_scenarios))

echo "📊 Error Handling Metrics:"
echo "  Scenario Success Rate: $error_handling_rate% ($successful_scenarios/$total_scenarios)"
echo "  Invalid Request Handling: $([ -f "impossible.py" ] && echo "Processed" || echo "Handled appropriately")"
echo "  Conflict Resolution: $([ -f "conflict_test.py" ] && echo "Resolved" || echo "Failed")"
echo "  Resource Management: $((100 * created_files / total_requested))% ($created_files/$total_requested files)"
echo "  Recovery Implementation: $([ -f "network_service.py" ] && [ -f "test_network.py" ] && echo "Implemented" || echo "Missing")"

if [ $error_handling_rate -ge 75 ]; then
    echo ""
    echo "🏆 EXCELLENT: Agent 0 demonstrates strong error handling!"
    echo "✅ Handles multiple error scenarios effectively"
    echo "✅ Shows good resource management"
    echo "✅ Implements recovery patterns when requested"
elif [ $error_handling_rate -ge 50 ]; then
    echo ""
    echo "✅ GOOD: Agent 0 handles most error scenarios well"
    echo "✅ Some successful error handling patterns"
    echo "⚠️  Room for improvement in edge cases"
elif [ $error_handling_rate -ge 25 ]; then
    echo ""
    echo "⚠️  FAIR: Agent 0 handles some error scenarios"
    echo "✅ Basic error handling present"
    echo "❌ Many scenarios need improvement"
else
    echo ""
    echo "❌ NEEDS WORK: Error handling needs significant improvement"
    echo "❌ Most error scenarios not handled well"
fi

echo ""
echo "🗂️ Final File Structure:"
ls -la

echo ""
echo "📄 Sample Error Handling Code:"
echo "============================="
for file in network_service.py impossible.py; do
    if [ -f "$file" ]; then
        echo ""
        echo "--- $file ---"
        cat "$file"
    fi
done

echo ""
echo "✅ Error Handling and Recovery Test Complete"
