#!/bin/bash

# Test Multi-File Coordination
# Tests Agent 0's ability to coordinate the creation of multiple JavaScript files
# through proper delegation to coding agents

echo "=== Multi-File Coordination Test ==="
echo "Testing Agent 0's ability to coordinate multi-step coding tasks..."
echo ""

# Cleanup any existing test files
echo "🧹 Cleaning up existing test files..."
rm -f TEST_OUTPUT_1.js TEST_OUTPUT_2.js TEST_OUTPUT_3.js
rm -f /tmp/agentry_multi_file_test.log

# Build latest version
echo "🔨 Building latest Agentry..."
make build > /dev/null 2>&1
if [ $? -ne 0 ]; then
    echo "❌ Build failed!"
    exit 1
fi

echo "✅ Build successful"
echo ""

# Create test workspace with proper configuration
TEST_WORKSPACE="/tmp/agentry-multi-file-test"
echo "📁 Setting up test workspace at $TEST_WORKSPACE..."
rm -rf "$TEST_WORKSPACE"
mkdir -p "$TEST_WORKSPACE"
cd "$TEST_WORKSPACE"

# Copy necessary files
cp /home/marco/Documents/GitHub/agentry/.agentry.yaml .
if [ -f "/home/marco/Documents/GitHub/agentry/.env.local" ]; then
    cp "/home/marco/Documents/GitHub/agentry/.env.local" .
    echo "✅ Copied configuration files to test workspace"
else
    echo "⚠️  No .env.local found - API calls may fail"
fi

# Copy the agentry executable (handle both .exe and plain executable names)
if [ -f "/home/marco/Documents/GitHub/agentry/agentry.exe" ]; then
    cp /home/marco/Documents/GitHub/agentry/agentry.exe ./agentry
    chmod +x ./agentry
elif [ -f "/home/marco/Documents/GitHub/agentry/agentry" ]; then
    cp /home/marco/Documents/GitHub/agentry/agentry ./agentry
    chmod +x ./agentry
else
    echo "❌ No agentry executable found!"
    exit 1
fi

echo "✅ Test workspace ready"
echo ""

# Define the test prompt - a realistic coding scenario
TEST_PROMPT="I need you to coordinate the creation of multiple JavaScript files for a simple web project. Please create:

1. TEST_OUTPUT_1.js - A utility module that exports functions for mathematical operations (add, subtract, multiply, divide)
2. TEST_OUTPUT_2.js - A main application file that imports the utility module and demonstrates its usage with sample calculations
3. TEST_OUTPUT_3.js - A test file that imports both modules and runs some basic tests

Please delegate these file creation tasks to appropriate agents rather than creating them yourself. Each file should have proper JavaScript syntax, comments, and demonstrate realistic coding patterns."

echo "📝 Test Prompt:"
echo "$TEST_PROMPT"
echo ""

# Run the test with timeout and logging
echo "🚀 Starting multi-file coordination test..."
echo "Agent 0 will coordinate the creation of 3 JavaScript files..."
echo ""

# Create input file with the test prompt and quit command
echo "$TEST_PROMPT" > /tmp/test_input.txt
echo "/quit" >> /tmp/test_input.txt

# Run the command with input from file
timeout 300s ./agentry chat < /tmp/test_input.txt > /tmp/agentry_multi_file_test.log 2>&1

TEST_EXIT_CODE=$?

echo "⏱️  Test completed with exit code: $TEST_EXIT_CODE"
echo ""

# Analysis and validation
echo "=== TEST ANALYSIS ==="

# Check if files were created
echo "📂 Checking file creation..."
FILES_CREATED=0

if [ -f "TEST_OUTPUT_1.js" ]; then
    echo "✅ TEST_OUTPUT_1.js created"
    FILES_CREATED=$((FILES_CREATED + 1))
else
    echo "❌ TEST_OUTPUT_1.js NOT created"
fi

if [ -f "TEST_OUTPUT_2.js" ]; then
    echo "✅ TEST_OUTPUT_2.js created"
    FILES_CREATED=$((FILES_CREATED + 1))
else
    echo "❌ TEST_OUTPUT_2.js NOT created"
fi

if [ -f "TEST_OUTPUT_3.js" ]; then
    echo "✅ TEST_OUTPUT_3.js created"
    FILES_CREATED=$((FILES_CREATED + 1))
else
    echo "❌ TEST_OUTPUT_3.js NOT created"
fi

echo "📊 Files created: $FILES_CREATED/3"
echo ""

# Analyze coordination behavior
echo "🤖 Analyzing coordination behavior..."
if [ -f "/tmp/agentry_multi_file_test.log" ]; then
    # Check for delegation patterns
    DELEGATION_COUNT=$(grep -c "agent\|assign_task\|delegate" /tmp/agentry_multi_file_test.log 2>/dev/null || true)
    DIRECT_FILE_OPS=$(grep -c "write_file\|create_file" /tmp/agentry_multi_file_test.log 2>/dev/null || true)
    CONTEXT_USAGE=$(grep -c "project_tree\|team_status\|check_agent" /tmp/agentry_multi_file_test.log 2>/dev/null || true)
    
    echo "📈 Coordination Metrics:"
    echo "   - Delegation attempts: $DELEGATION_COUNT"
    echo "   - Direct file operations: $DIRECT_FILE_OPS"
    echo "   - Context tool usage: $CONTEXT_USAGE"
    
    if [ $DELEGATION_COUNT -gt 0 ]; then
        echo "✅ Agent 0 attempted delegation"
    else
        echo "⚠️  No clear delegation patterns detected"
    fi
    
    if [ $DIRECT_FILE_OPS -lt $DELEGATION_COUNT ]; then
        echo "✅ More delegation than direct operations (good)"
    else
        echo "⚠️  High direct operations vs delegation ratio"
    fi
else
    echo "❌ Could not analyze log file"
fi

echo ""

# Validate file contents
echo "📄 Validating file contents..."
CONTENT_SCORE=0

if [ -f "TEST_OUTPUT_1.js" ]; then
    if grep -q "function\|export\|module.exports" TEST_OUTPUT_1.js && grep -q "add\|subtract\|multiply\|divide" TEST_OUTPUT_1.js; then
        echo "✅ TEST_OUTPUT_1.js has expected utility functions"
        CONTENT_SCORE=$((CONTENT_SCORE + 1))
    else
        echo "⚠️  TEST_OUTPUT_1.js missing expected utility functions"
    fi
fi

if [ -f "TEST_OUTPUT_2.js" ]; then
    if grep -q "import\|require" TEST_OUTPUT_2.js && grep -q "TEST_OUTPUT_1" TEST_OUTPUT_2.js; then
        echo "✅ TEST_OUTPUT_2.js imports utility module"
        CONTENT_SCORE=$((CONTENT_SCORE + 1))
    else
        echo "⚠️  TEST_OUTPUT_2.js doesn't properly import utility module"
    fi
fi

if [ -f "TEST_OUTPUT_3.js" ]; then
    if grep -q "test\|Test" TEST_OUTPUT_3.js && (grep -q "import\|require" TEST_OUTPUT_3.js || grep -q "TEST_OUTPUT" TEST_OUTPUT_3.js); then
        echo "✅ TEST_OUTPUT_3.js appears to be a test file"
        CONTENT_SCORE=$((CONTENT_SCORE + 1))
    else
        echo "⚠️  TEST_OUTPUT_3.js doesn't appear to be a proper test file"
    fi
fi

echo "📊 Content validation score: $CONTENT_SCORE/3"
echo ""

# Overall assessment
echo "=== OVERALL ASSESSMENT ==="
TOTAL_SCORE=0

# File creation (40% weight)
FILE_SCORE_PERCENT=$((FILES_CREATED * 100 / 3))
if [ $FILES_CREATED -eq 3 ]; then
    TOTAL_SCORE=$((TOTAL_SCORE + 40))
    echo "✅ File Creation: EXCELLENT ($FILES_CREATED/3 files)"
elif [ $FILES_CREATED -eq 2 ]; then
    TOTAL_SCORE=$((TOTAL_SCORE + 25))
    echo "⚠️  File Creation: GOOD ($FILES_CREATED/3 files)"
elif [ $FILES_CREATED -eq 1 ]; then
    TOTAL_SCORE=$((TOTAL_SCORE + 10))
    echo "⚠️  File Creation: POOR ($FILES_CREATED/3 files)"
else
    echo "❌ File Creation: FAILED ($FILES_CREATED/3 files)"
fi

# Coordination behavior (35% weight)
if [ $DELEGATION_COUNT -gt 2 ] && [ $CONTEXT_USAGE -gt 0 ]; then
    TOTAL_SCORE=$((TOTAL_SCORE + 35))
    echo "✅ Coordination: EXCELLENT (Strong delegation + context usage)"
elif [ $DELEGATION_COUNT -gt 0 ]; then
    TOTAL_SCORE=$((TOTAL_SCORE + 20))
    echo "⚠️  Coordination: GOOD (Some delegation detected)"
else
    echo "❌ Coordination: POOR (No clear delegation)"
fi

# Content quality (25% weight)
if [ $CONTENT_SCORE -eq 3 ]; then
    TOTAL_SCORE=$((TOTAL_SCORE + 25))
    echo "✅ Content Quality: EXCELLENT (All files have expected content)"
elif [ $CONTENT_SCORE -eq 2 ]; then
    TOTAL_SCORE=$((TOTAL_SCORE + 15))
    echo "⚠️  Content Quality: GOOD (Most files have expected content)"
elif [ $CONTENT_SCORE -eq 1 ]; then
    TOTAL_SCORE=$((TOTAL_SCORE + 5))
    echo "⚠️  Content Quality: POOR (Some files have expected content)"
else
    echo "❌ Content Quality: FAILED (No files have expected content)"
fi

echo ""
echo "🎯 FINAL SCORE: $TOTAL_SCORE/100"

if [ $TOTAL_SCORE -ge 80 ]; then
    echo "🏆 RESULT: EXCELLENT - Multi-file coordination working very well!"
elif [ $TOTAL_SCORE -ge 60 ]; then
    echo "👍 RESULT: GOOD - Multi-file coordination mostly working, minor issues"
elif [ $TOTAL_SCORE -ge 40 ]; then
    echo "⚠️  RESULT: FAIR - Multi-file coordination partially working, needs improvement"
else
    echo "❌ RESULT: POOR - Multi-file coordination not working effectively"
fi

echo ""

# Show sample file contents if created
echo "=== SAMPLE OUTPUTS ==="
for file in TEST_OUTPUT_1.js TEST_OUTPUT_2.js TEST_OUTPUT_3.js; do
    if [ -f "$file" ]; then
        echo "📄 $file (first 10 lines):"
        head -10 "$file" | sed 's/^/   /'
        echo ""
    fi
done

# Show coordination log excerpts
echo "=== COORDINATION LOG EXCERPTS ==="
if [ -f "/tmp/agentry_multi_file_test.log" ]; then
    echo "🤖 Key coordination activities:"
    grep -i "agent\|delegate\|assign\|create\|file" /tmp/agentry_multi_file_test.log | head -20 | sed 's/^/   /'
    echo ""
    
    echo "📊 Full log available at: /tmp/agentry_multi_file_test.log"
    echo "📊 Log size: $(wc -l < /tmp/agentry_multi_file_test.log) lines"
else
    echo "❌ No coordination log available"
fi

echo ""
echo "=== TEST COMPLETE ==="

# Automatic cleanup after analysis
echo "🧹 Cleaning up test files..."
cd /tmp/agentry-multi-file-test
rm -f TEST_OUTPUT_1.js TEST_OUTPUT_2.js TEST_OUTPUT_3.js
rm -f /tmp/test_input.txt
echo "✅ Test files cleaned up automatically"

# Return to original directory
cd /home/marco/Documents/GitHub/agentry

exit 0
