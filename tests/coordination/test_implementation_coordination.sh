#!/bin/bash

# Test Implementation-Focused Multi-Language Coordination
# Focus on actual code creation rather than project structure

echo "=== Implementation-Focused Coordination Test ==="
echo "🎯 Testing Agent 0's ability to coordinate actual code implementation..."
echo ""

# Setup workspace
AI_WORKSPACE="/tmp/agentry-ai-sandbox"
PROJECT_DIR="/home/marco/Documents/GitHub/agentry"

cd "$AI_WORKSPACE"
rm -rf taskmaster-impl 2>/dev/null

# More specific implementation-focused prompt
TEST_PROMPT="I need you to coordinate the implementation of specific files for a task management API. Please delegate the creation of these EXACT files with WORKING CODE:

🐍 **PYTHON FILES (delegate to coder agent):**
1. 'app.py' - Complete Flask API with these working endpoints:
   - GET /api/tasks (returns JSON list)
   - POST /api/tasks (accepts JSON, returns created task)
   - GET /api/health (returns status)

2. 'models.py' - SQLAlchemy Task model with id, title, description, completed fields

📜 **JAVASCRIPT FILES (delegate to coder agent):**  
3. 'frontend.html' - Complete HTML page that calls the Flask API endpoints
4. 'api-client.js' - JavaScript functions to interact with the Flask API

🗄️ **DATABASE FILE (delegate to coder agent):**
5. 'schema.sql' - CREATE TABLE statement for tasks table

Each file must contain WORKING, EXECUTABLE CODE - not placeholders or comments. The files should demonstrate actual cross-technology integration (HTML calls JS, JS calls Python API, Python uses SQL schema).

IMPORTANT: Focus on CREATING ACTUAL IMPLEMENTATION FILES, not project structure or documentation."

echo "📝 Implementation-focused test prompt:"
echo "$TEST_PROMPT"
echo ""

echo "$TEST_PROMPT" > /tmp/test_input.txt  
echo "/quit" >> /tmp/test_input.txt

echo "🚀 Starting implementation coordination test..."
timeout 300s ./agentry chat < /tmp/test_input.txt > /tmp/agentry_implementation_test.log 2>&1

echo "📊 IMPLEMENTATION RESULTS:"
echo "========================="

# Check for actual implementation files
PYTHON_FILES=$(find . -name "*.py" -type f 2>/dev/null | wc -l)
JS_FILES=$(find . -name "*.js" -type f 2>/dev/null | wc -l)
HTML_FILES=$(find . -name "*.html" -type f 2>/dev/null | wc -l)
SQL_FILES=$(find . -name "*.sql" -type f 2>/dev/null | wc -l)

echo "📁 Implementation files created:"
echo "   🐍 Python files: $PYTHON_FILES"
echo "   📜 JavaScript files: $JS_FILES"
echo "   🌐 HTML files: $HTML_FILES"
echo "   🗄️  SQL files: $SQL_FILES"

if [ $PYTHON_FILES -gt 0 ] || [ $JS_FILES -gt 0 ] || [ $HTML_FILES -gt 0 ] || [ $SQL_FILES -gt 0 ]; then
    echo ""
    echo "✅ ACTUAL IMPLEMENTATION FILES CREATED!"
    echo ""
    
    # Show the actual files and their contents
    find . -name "*.py" -o -name "*.js" -o -name "*.html" -o -name "*.sql" | while read file; do
        echo "📄 $file:"
        echo "----------------------------------------"
        cat "$file"
        echo ""
        echo "----------------------------------------"
        echo ""
    done
    
    # Check for cross-technology integration
    echo "🔍 CROSS-TECHNOLOGY INTEGRATION CHECK:"
    
    # Check if API endpoints are consistent
    if find . -name "*.py" -exec grep -l "api/tasks\|/api/" {} \; | head -1 >/dev/null && \
       find . -name "*.js" -o -name "*.html" -exec grep -l "api/tasks\|/api/" {} \; | head -1 >/dev/null; then
        echo "   ✅ API endpoints coordinated between backend and frontend"
    else
        echo "   ❌ API endpoints not coordinated"
    fi
    
    # Check if database schema matches models
    if find . -name "*.sql" -exec grep -l "tasks\|task" {} \; | head -1 >/dev/null && \
       find . -name "*.py" -exec grep -l "Task\|task" {} \; | head -1 >/dev/null; then
        echo "   ✅ Database schema appears aligned with models"
    else
        echo "   ❌ Database schema not aligned"
    fi
    
else
    echo ""
    echo "❌ NO IMPLEMENTATION FILES CREATED"
    echo "   Agent 0 may have focused on structure rather than implementation"
fi

# Analyze coordination patterns
echo ""
echo "🤖 COORDINATION PATTERN ANALYSIS:"
if [ -f "/tmp/agentry_implementation_test.log" ]; then
    DELEGATION_COUNT=$(grep -c "agent\|delegate\|coder" /tmp/agentry_implementation_test.log 2>/dev/null || echo "0")
    FILE_CREATES=$(grep -c "create\|Creating.*file\|\.py\|\.js\|\.html\|\.sql" /tmp/agentry_implementation_test.log 2>/dev/null || echo "0")
    
    echo "   📊 Delegation activities: $DELEGATION_COUNT"
    echo "   📝 File creation activities: $FILE_CREATES"
    
    if [ $DELEGATION_COUNT -gt 3 ] && [ $FILE_CREATES -gt 3 ]; then
        echo "   ✅ Strong implementation-focused coordination"
    elif [ $DELEGATION_COUNT -gt 0 ]; then
        echo "   ⚠️  Some coordination but may lack implementation focus"
    else
        echo "   ❌ Weak coordination patterns"
    fi
    
    echo ""
    echo "📋 Key coordination events:"
    grep -n "delegate\|coder\|create.*\.\(py\|js\|html\|sql\)" /tmp/agentry_implementation_test.log | head -10 | sed 's/^/   /'
fi

echo ""
echo "=== IMPLEMENTATION COORDINATION TEST COMPLETE ==="
