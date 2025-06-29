#!/bin/bash

# Quick Coordination Test with Immediate Feedback
# Tests Agent 0 coordination with visible output and shorter timeout

echo "=== Quick Multi-Language Coordination Test ==="
echo "🎯 Testing coordination with immediate feedback..."
echo ""

# Setup
AI_WORKSPACE="/tmp/agentry-ai-sandbox"
PROJECT_DIR="/home/marco/Documents/GitHub/agentry"

cd "$AI_WORKSPACE" 2>/dev/null || {
    echo "❌ Workspace not found, creating..."
    mkdir -p "$AI_WORKSPACE"
    cd "$AI_WORKSPACE"
    cp "$PROJECT_DIR/.agentry.yaml" .
    if [ -f "$PROJECT_DIR/.env.local" ]; then
        cp "$PROJECT_DIR/.env.local" .
    fi
    if [ -f "$PROJECT_DIR/agentry" ]; then
        cp "$PROJECT_DIR/agentry" .
        chmod +x agentry
    fi
}

# Clean previous test
rm -rf simple-coordination-test 2>/dev/null

# Simple, direct coordination test
SIMPLE_TEST="Create a simple web app called 'simple-coordination-test' with just 3 files:

1. 'backend.py' - A simple Flask API with one endpoint: GET /api/hello that returns {\"message\": \"Hello World\"}

2. 'frontend.html' - A simple HTML page with a button that calls the Flask API and displays the response

3. 'run.sh' - A bash script to run the Flask app

Create these 3 files with actual working code. Keep it simple but functional."

echo "📝 Simple Test Request:"
echo "$SIMPLE_TEST"
echo ""
echo "🚀 Starting test (30 second timeout)..."

# Run with shorter timeout and direct output monitoring
echo "$SIMPLE_TEST" | timeout 30s ./agentry chat --verbose 2>&1 | while IFS= read -r line; do
    case "$line" in
        *agent*|*Agent*|*delegate*|*spawn*)
            echo "🤖 AGENT: $line"
            ;;
        *create*|*Create*|*file*|*File*)
            echo "📝 FILE: $line"
            ;;
        *tool*|*Tool*|*using*|*Using*)
            echo "🛠️  TOOL: $line"
            ;;
        *error*|*Error*|*failed*|*Failed*)
            echo "❌ ERROR: $line"
            ;;
        *)
            if [[ ${#line} -gt 10 && ! "$line" =~ ^[[:space:]]*$ ]]; then
                echo "   $line"
            fi
            ;;
    esac
done

RESULT=$?

echo ""
echo "⏱️  Test completed with result: $RESULT"
echo ""

# Quick results check
echo "📊 QUICK RESULTS:"
if [ -d "simple-coordination-test" ]; then
    echo "✅ Project directory created"
    
    FILES_FOUND=$(find simple-coordination-test -name "*.py" -o -name "*.html" -o -name "*.sh" | wc -l)
    echo "📁 Implementation files found: $FILES_FOUND"
    
    if [ $FILES_FOUND -gt 0 ]; then
        echo ""
        echo "📄 Files created:"
        find simple-coordination-test -type f | while read file; do
            echo "   📄 $file ($(wc -l < "$file") lines)"
            echo "      Content preview:"
            head -5 "$file" | sed 's/^/         /'
            echo ""
        done
    fi
    
    if [ $FILES_FOUND -ge 3 ]; then
        echo "🎯 SUCCESS: Agent coordination created multiple files!"
    elif [ $FILES_FOUND -gt 0 ]; then
        echo "⚠️  PARTIAL: Some files created, coordination working"
    else
        echo "❌ FAILED: No implementation files created"
    fi
else
    echo "❌ No project directory created"
fi

echo ""
echo "=== QUICK TEST COMPLETE ==="
