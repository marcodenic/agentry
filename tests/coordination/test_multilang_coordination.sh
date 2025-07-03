#!/bin/bash

# Test Multi-Language Project Coordination - Priority 1
# Tests Agent 0's ability to coordinate complex polyglot projects
# Focus: Coordination intelligence, task decomposition, and cross-technology integration

echo "=== Multi-Language Project Coordination Test ==="
echo "🎯 Testing Agent 0's coordination intelligence across technologies..."
echo "🛡️  Using isolated AI workspace for safety"
echo ""

# Cleanup any existing test files
echo "🧹 Cleaning up existing test files..."
rm -rf /tmp/agentry-multilang-test
rm -f /tmp/agentry_multilang_test.log

# Build latest version
echo "🔨 Building latest Agentry..."
cd /home/marco/Documents/GitHub/agentry
make build > /dev/null 2>&1
if [ $? -ne 0 ]; then
    echo "❌ Build failed!"
    exit 1
fi

echo "✅ Build successful"
echo ""

# Create isolated AI workspace
AI_WORKSPACE="/tmp/agentry-ai-sandbox"
PROJECT_DIR="/home/marco/Documents/GitHub/agentry"

echo "🏗️  Setting up isolated AI workspace..."
echo "🛡️  Workspace: $AI_WORKSPACE (isolated from project)"
rm -rf "$AI_WORKSPACE"
mkdir -p "$AI_WORKSPACE"
cd "$AI_WORKSPACE"

# Copy only necessary files (not source code)
cp "$PROJECT_DIR/.agentry.yaml" .
if [ -f "$PROJECT_DIR/.env.local" ]; then
    cp "$PROJECT_DIR/.env.local" .
    echo "✅ Copied configuration files to isolated workspace"
else
    echo "⚠️  No .env.local found - API calls may fail"
fi

# Copy the agentry executable
if [ -f "$PROJECT_DIR/agentry.exe" ]; then
    cp "$PROJECT_DIR/agentry.exe" ./agentry
    chmod +x ./agentry
elif [ -f "$PROJECT_DIR/agentry" ]; then
    cp "$PROJECT_DIR/agentry" ./agentry
    chmod +x ./agentry
else
    echo "❌ No agentry executable found!"
    exit 1
fi

echo "✅ Isolated AI workspace ready"
echo "📁 Working directory: $(pwd)"
echo "🛡️  Project source code is safely isolated at: $PROJECT_DIR"
echo ""

# Define the comprehensive multi-language test prompt
TEST_PROMPT="I need you to coordinate the creation of a complete full-stack web application called 'taskmaster-app'. This requires coordination across multiple technologies and careful attention to how components integrate.

Please create the following coordinated project structure:

📁 **taskmaster-app/**
├── **backend/**
│   ├── app.py (Flask API with endpoints: GET /api/tasks, POST /api/tasks, DELETE /api/tasks/<id>, GET /api/health)
│   ├── models.py (Task model with id, title, description, completed, created_at)
│   ├── database.py (SQLite database initialization and connection)
│   └── requirements.txt (Flask, SQLite dependencies)
├── **frontend/**
│   ├── index.html (Task management interface)
│   ├── app.js (JavaScript to interact with backend API)
│   └── style.css (Basic styling)
├── **database/**
│   └── schema.sql (Tasks table creation script)
├── **docker/**
│   ├── Dockerfile (Multi-stage build for the application)
│   └── docker-compose.yml (Full-stack orchestration)
└── README.md (Complete setup and usage instructions)

**CRITICAL COORDINATION REQUIREMENTS:**
1. API endpoints in app.py must match exactly what app.js calls
2. Database schema must align with models.py structure  
3. Docker configuration must properly orchestrate frontend + backend
4. All imports, URLs, and references must be consistent across files
5. README must provide accurate setup instructions for the actual files created

This is a test of your coordination intelligence - ensure all components work together as a cohesive system. Delegate file creation tasks appropriately but maintain overall coherence."

echo "📝 Test Prompt:"
echo "$TEST_PROMPT"
echo ""

# Run the test with timeout and logging
echo "🚀 Starting multi-language coordination test..."
echo "Agent 0 will coordinate creation of a full-stack application..."
echo ""

# Create input file with the test prompt and quit command
echo "$TEST_PROMPT" > /tmp/test_input.txt
echo "/quit" >> /tmp/test_input.txt

# Run the command with input from file and real-time monitoring
echo "🚀 Starting coordination test with real-time monitoring..."
echo "📺 Watch for agent spawning and tool usage patterns..."
echo ""

# Start the test in background while monitoring the log
timeout 600s ./agentry chat < /tmp/test_input.txt > /tmp/agentry_multilang_test.log 2>&1 &
AGENTRY_PID=$!

# Monitor the log file in real-time
echo "🔍 REAL-TIME COORDINATION MONITORING:"
echo "======================================================================================================="

# Function to highlight interesting log lines
monitor_coordination() {
    local log_file="/tmp/agentry_multilang_test.log"
    local last_size=0
    
    while kill -0 $AGENTRY_PID 2>/dev/null; do
        if [ -f "$log_file" ]; then
            current_size=$(wc -c < "$log_file" 2>/dev/null || echo "0")
            if [ $current_size -gt $last_size ]; then
                # Get new content since last check
                tail -c +$((last_size + 1)) "$log_file" | while IFS= read -r line; do
                    case "$line" in
                        *agent*|*Agent*|*delegate*|*Delegate*)
                            echo "🤖 AGENT ACTIVITY: $line"
                            ;;
                        *create*|*Create*|*write*|*Write*)
                            echo "📝 FILE OPERATION: $line"
                            ;;
                        *team_status*|*project_tree*|*list*|*find*)
                            echo "🔍 CONTEXT TOOL: $line"
                            ;;
                        *assign_task*|*send_message*)
                            echo "📢 COORDINATION: $line"
                            ;;
                        *error*|*Error*|*failed*|*Failed*)
                            echo "❌ ERROR: $line"
                            ;;
                        *backend*|*frontend*|*database*|*docker*)
                            echo "🏗️  PROJECT COMPONENT: $line"
                            ;;
                        *)
                            # Show other lines with less emphasis
                            if [[ ${#line} -gt 10 ]]; then
                                echo "   $line"
                            fi
                            ;;
                    esac
                done
                last_size=$current_size
            fi
        fi
        sleep 2
    done
}

# Start monitoring in background
monitor_coordination &
MONITOR_PID=$!

# Wait for the main process to complete
wait $AGENTRY_PID
TEST_EXIT_CODE=$?

# Stop monitoring
kill $MONITOR_PID 2>/dev/null || true

echo ""
echo "======================================================================================================="
echo "⏱️  Test completed with exit code: $TEST_EXIT_CODE"
echo ""

# Analysis and validation
echo "=== COORDINATION ANALYSIS ==="

# Check if project structure was created
echo "📂 Checking project structure creation..."
PROJECT_CREATED=0

if [ -d "taskmaster-app" ]; then
    echo "✅ Main project directory created"
    PROJECT_CREATED=1
else
    echo "❌ Main project directory NOT created"
fi

# Count files across different technologies
if [ -d "taskmaster-app" ]; then
    PYTHON_FILES=$(find taskmaster-app -name "*.py" | wc -l)
    JS_FILES=$(find taskmaster-app -name "*.js" | wc -l)  
    HTML_FILES=$(find taskmaster-app -name "*.html" | wc -l)
    CSS_FILES=$(find taskmaster-app -name "*.css" | wc -l)
    SQL_FILES=$(find taskmaster-app -name "*.sql" | wc -l)
    DOCKER_FILES=$(find taskmaster-app -name "Dockerfile" -o -name "docker-compose.yml" | wc -l)
    CONFIG_FILES=$(find taskmaster-app -name "requirements.txt" -o -name "*.json" | wc -l)
    TOTAL_FILES=$(find taskmaster-app -type f | wc -l)
    
    echo "📊 Files created by technology:"
    echo "   🐍 Python files: $PYTHON_FILES"
    echo "   📜 JavaScript files: $JS_FILES"
    echo "   🌐 HTML files: $HTML_FILES"
    echo "   🎨 CSS files: $CSS_FILES"
    echo "   🗄️  SQL files: $SQL_FILES"
    echo "   🐳 Docker files: $DOCKER_FILES"
    echo "   ⚙️  Config files: $CONFIG_FILES"
    echo "   📁 Total files: $TOTAL_FILES"
    echo ""
    
    # Show project structure
    echo "🏗️  Project structure created:"
    if command -v tree >/dev/null 2>&1; then
        tree taskmaster-app
    else
        find taskmaster-app -type f | sort | sed 's/^/   /'
    fi
    echo ""
fi

# Analyze coordination quality - check for cross-technology integration
echo "🔍 Analyzing coordination quality..."
COORDINATION_SCORE=0

# Test 1: API endpoint consistency between backend and frontend
if [ -f "taskmaster-app/backend/app.py" ] && [ -f "taskmaster-app/frontend/app.js" ]; then
    # Check if backend defines API endpoints
    API_ENDPOINTS_BACKEND=$(grep -c "^@app.route.*api" taskmaster-app/backend/app.py 2>/dev/null || echo "0")
    # Check if frontend calls matching endpoints
    API_CALLS_FRONTEND=$(grep -c "api/" taskmaster-app/frontend/app.js 2>/dev/null || echo "0")
    
    if [ $API_ENDPOINTS_BACKEND -gt 0 ] && [ $API_CALLS_FRONTEND -gt 0 ]; then
        echo "✅ API coordination: Backend defines $API_ENDPOINTS_BACKEND endpoints, frontend makes $API_CALLS_FRONTEND API calls"
        COORDINATION_SCORE=$((COORDINATION_SCORE + 25))
    else
        echo "❌ API coordination: No clear API integration between backend and frontend"
    fi
else
    echo "❌ Missing core backend or frontend files"
fi

# Test 2: Database schema alignment
if [ -f "taskmaster-app/database/schema.sql" ] && [ -f "taskmaster-app/backend/models.py" ]; then
    # Check if SQL schema and Python models have similar structure
    if grep -q "tasks\|task" taskmaster-app/database/schema.sql && grep -q "Task\|task" taskmaster-app/backend/models.py; then
        echo "✅ Database coordination: Schema and models appear aligned"
        COORDINATION_SCORE=$((COORDINATION_SCORE + 20))
    else
        echo "❌ Database coordination: Schema and models don't appear aligned"
    fi
else
    echo "⚠️  Database coordination: Missing schema or models files"
fi

# Test 3: Docker orchestration completeness
if [ -f "taskmaster-app/docker/docker-compose.yml" ]; then
    # Check if docker-compose references the application structure
    if grep -q "backend\|frontend\|app" taskmaster-app/docker/docker-compose.yml; then
        echo "✅ Docker coordination: Compose file references application components"
        COORDINATION_SCORE=$((COORDINATION_SCORE + 15))
    else
        echo "❌ Docker coordination: Compose file doesn't reference app components"
    fi
else
    echo "⚠️  Docker coordination: Missing docker-compose.yml"
fi

# Test 4: Configuration consistency
if [ -f "taskmaster-app/backend/requirements.txt" ]; then
    # Check if requirements include Flask (since we specified Flask API)
    if grep -q -i "flask" taskmaster-app/backend/requirements.txt; then
        echo "✅ Configuration coordination: Requirements include Flask as specified"
        COORDINATION_SCORE=$((COORDINATION_SCORE + 10))
    else
        echo "⚠️  Configuration coordination: Requirements don't include Flask"
    fi
else
    echo "⚠️  Configuration coordination: Missing requirements.txt"
fi

# Test 5: Documentation accuracy
if [ -f "taskmaster-app/README.md" ]; then
    # Check if README references the actual files created
    README_ACCURACY=0
    if grep -q "app.py\|backend" taskmaster-app/README.md; then
        README_ACCURACY=$((README_ACCURACY + 1))
    fi
    if grep -q "index.html\|frontend" taskmaster-app/README.md; then
        README_ACCURACY=$((README_ACCURACY + 1))
    fi
    if grep -q "docker\|Docker" taskmaster-app/README.md; then
        README_ACCURACY=$((README_ACCURACY + 1))
    fi
    
    if [ $README_ACCURACY -ge 2 ]; then
        echo "✅ Documentation coordination: README references actual project components"
        COORDINATION_SCORE=$((COORDINATION_SCORE + 10))
    else
        echo "⚠️  Documentation coordination: README doesn't accurately reflect project"
    fi
else
    echo "❌ Documentation coordination: Missing README.md"
fi

echo ""
echo "🎯 COORDINATION SCORE: $COORDINATION_SCORE/80"

# Analyze coordination behavior from logs
echo "🤖 DETAILED COORDINATION BEHAVIOR ANALYSIS..."
if [ -f "/tmp/agentry_multilang_test.log" ]; then
    # Count different types of activities
    AGENT_SPAWNS=$(grep -c "Spawning\|spawning\|new agent\|agent.*created" /tmp/agentry_multilang_test.log 2>/dev/null || echo "0")
    DELEGATION_COUNT=$(grep -c "agent\|assign_task\|delegate\|Delegating" /tmp/agentry_multilang_test.log 2>/dev/null || echo "0")
    CONTEXT_USAGE=$(grep -c "project_tree\|team_status\|list\|find\|fileinfo" /tmp/agentry_multilang_test.log 2>/dev/null || echo "0")
    FILE_OPERATIONS=$(grep -c "create\|write_file\|edit_range\|Creating file" /tmp/agentry_multilang_test.log 2>/dev/null || echo "0")
    TOOL_USAGE=$(grep -c "Using tool\|Tool:\|Executing tool" /tmp/agentry_multilang_test.log 2>/dev/null || echo "0")
    
    echo "📊 COORDINATION ACTIVITY BREAKDOWN:"
    echo "   🤖 Agent spawns/delegations: $AGENT_SPAWNS"
    echo "   📋 Total delegation activities: $DELEGATION_COUNT"
    echo "   🔍 Context gathering operations: $CONTEXT_USAGE"
    echo "   📝 File creation/edit operations: $FILE_OPERATIONS"
    echo "   🛠️  Total tool usage: $TOOL_USAGE"
    echo ""
    
    # Show specific agent coordination patterns
    echo "🎯 AGENT COORDINATION PATTERNS:"
    echo "   Agent spawn/delegation events:"
    grep -n "agent\|Agent\|delegate\|Delegate\|spawn\|Spawn" /tmp/agentry_multilang_test.log | head -10 | sed 's/^/      /'
    echo ""
    
    echo "   Context gathering activities:"
    grep -n "project_tree\|team_status\|list\|find" /tmp/agentry_multilang_test.log | head -8 | sed 's/^/      /'
    echo ""
    
    echo "   File creation activities:"
    grep -n "create\|Create\|write\|Write.*file" /tmp/agentry_multilang_test.log | head -10 | sed 's/^/      /'
    echo ""
    
    # Calculate coordination intelligence metrics
    if [ $FILE_OPERATIONS -gt 0 ]; then
        DELEGATION_RATIO=$((DELEGATION_COUNT * 100 / FILE_OPERATIONS))
        CONTEXT_RATIO=$((CONTEXT_USAGE * 100 / FILE_OPERATIONS))
        
        echo "🧠 COORDINATION INTELLIGENCE METRICS:"
        echo "   📊 Delegation efficiency: $DELEGATION_RATIO% (delegations per file operation)"
        echo "   🔍 Context awareness: $CONTEXT_RATIO% (context gathering per file operation)"
        
        if [ $DELEGATION_RATIO -gt 50 ]; then
            echo "   ✅ EXCELLENT: High delegation pattern indicates good coordination"
        elif [ $DELEGATION_RATIO -gt 25 ]; then
            echo "   ⚠️  MODERATE: Some delegation but could be more coordinated"
        else
            echo "   ❌ LOW: Mostly direct operations, limited coordination"
        fi
        
        if [ $CONTEXT_RATIO -gt 30 ]; then
            echo "   ✅ EXCELLENT: High context awareness indicates intelligent coordination"
        elif [ $CONTEXT_RATIO -gt 15 ]; then
            echo "   ⚠️  MODERATE: Some context gathering but could be more thorough"
        else
            echo "   ❌ LOW: Limited context awareness"
        fi
    fi
else
    echo "❌ Could not analyze coordination behavior - no log file"
fi

echo ""

# Overall assessment
echo "=== OVERALL ASSESSMENT ==="
TOTAL_SCORE=0

# Project structure (30% weight)
if [ $PROJECT_CREATED -eq 1 ] && [ $TOTAL_FILES -ge 8 ]; then
    TOTAL_SCORE=$((TOTAL_SCORE + 30))
    echo "✅ Project Structure: EXCELLENT (Complete multi-technology project created)"
elif [ $PROJECT_CREATED -eq 1 ] && [ $TOTAL_FILES -ge 5 ]; then
    TOTAL_SCORE=$((TOTAL_SCORE + 20))
    echo "⚠️  Project Structure: GOOD (Partial project created)"
elif [ $PROJECT_CREATED -eq 1 ]; then
    TOTAL_SCORE=$((TOTAL_SCORE + 10))
    echo "⚠️  Project Structure: POOR (Minimal project created)"
else
    echo "❌ Project Structure: FAILED (No project created)"
fi

# Technology coverage (25% weight)
TECH_COVERAGE=0
if [ $PYTHON_FILES -gt 0 ]; then TECH_COVERAGE=$((TECH_COVERAGE + 1)); fi
if [ $JS_FILES -gt 0 ]; then TECH_COVERAGE=$((TECH_COVERAGE + 1)); fi
if [ $SQL_FILES -gt 0 ]; then TECH_COVERAGE=$((TECH_COVERAGE + 1)); fi
if [ $DOCKER_FILES -gt 0 ]; then TECH_COVERAGE=$((TECH_COVERAGE + 1)); fi

if [ $TECH_COVERAGE -eq 4 ]; then
    TOTAL_SCORE=$((TOTAL_SCORE + 25))
    echo "✅ Technology Coverage: EXCELLENT (All 4 technologies present)"
elif [ $TECH_COVERAGE -eq 3 ]; then
    TOTAL_SCORE=$((TOTAL_SCORE + 18))
    echo "⚠️  Technology Coverage: GOOD (3/4 technologies present)"
elif [ $TECH_COVERAGE -ge 2 ]; then
    TOTAL_SCORE=$((TOTAL_SCORE + 10))
    echo "⚠️  Technology Coverage: FAIR (2+ technologies present)"
else
    echo "❌ Technology Coverage: POOR (< 2 technologies present)"
fi

# Coordination quality (35% weight)
COORD_PERCENT=$((COORDINATION_SCORE * 100 / 80))
if [ $COORDINATION_SCORE -ge 60 ]; then
    TOTAL_SCORE=$((TOTAL_SCORE + 35))
    echo "✅ Coordination Quality: EXCELLENT ($COORD_PERCENT% - Strong cross-technology integration)"
elif [ $COORDINATION_SCORE -ge 40 ]; then
    TOTAL_SCORE=$((TOTAL_SCORE + 25))
    echo "⚠️  Coordination Quality: GOOD ($COORD_PERCENT% - Moderate integration)"
elif [ $COORDINATION_SCORE -ge 20 ]; then
    TOTAL_SCORE=$((TOTAL_SCORE + 15))
    echo "⚠️  Coordination Quality: FAIR ($COORD_PERCENT% - Basic integration)"
else
    echo "❌ Coordination Quality: POOR ($COORD_PERCENT% - Weak integration)"
fi

# Delegation intelligence (10% weight)
if [ $DELEGATION_COUNT -gt 5 ] && [ $CONTEXT_USAGE -gt 3 ]; then
    TOTAL_SCORE=$((TOTAL_SCORE + 10))
    echo "✅ Delegation Intelligence: EXCELLENT (Strong coordination patterns)"
elif [ $DELEGATION_COUNT -gt 2 ]; then
    TOTAL_SCORE=$((TOTAL_SCORE + 6))
    echo "⚠️  Delegation Intelligence: GOOD (Some coordination patterns)"
else
    echo "❌ Delegation Intelligence: POOR (Weak coordination patterns)"
fi

echo ""
echo "🎯 FINAL SCORE: $TOTAL_SCORE/100"

if [ $TOTAL_SCORE -ge 85 ]; then
    echo "🏆 RESULT: EXCELLENT - Multi-language coordination working superbly!"
    echo "   Agent 0 demonstrates strong coordination intelligence across technologies"
elif [ $TOTAL_SCORE -ge 70 ]; then
    echo "👍 RESULT: GOOD - Multi-language coordination working well with minor gaps"
    echo "   Agent 0 shows solid coordination with room for improvement"
elif [ $TOTAL_SCORE -ge 50 ]; then
    echo "⚠️  RESULT: FAIR - Multi-language coordination partially working"
    echo "   Agent 0 shows basic coordination but needs enhancement"
else
    echo "❌ RESULT: POOR - Multi-language coordination needs significant work"
    echo "   Agent 0 coordination intelligence requires major improvements"
fi

echo ""

# Show sample outputs for analysis
echo "=== SAMPLE OUTPUTS ==="
if [ -d "taskmaster-app" ]; then
    for file in taskmaster-app/backend/app.py taskmaster-app/frontend/index.html taskmaster-app/README.md; do
        if [ -f "$file" ]; then
            echo "📄 $file (first 15 lines):"
            head -15 "$file" | sed 's/^/   /'
            echo ""
        fi
    done
fi

# Show coordination log excerpts
echo "=== COORDINATION EVENT TIMELINE ==="
if [ -f "/tmp/agentry_multilang_test.log" ]; then
    echo "🕐 KEY COORDINATION EVENTS (chronological order):"
    echo ""
    
    # Show timestamped coordination events
    grep -n "agent\|Agent\|delegate\|create\|backend\|frontend\|database\|docker\|spawn\|tool" /tmp/agentry_multilang_test.log | head -40 | while IFS=':' read -r line_num content; do
        # Color code different types of activities
        case "$content" in
            *agent*|*Agent*|*delegate*|*spawn*)
                echo "   [$line_num] 🤖 COORDINATION: $content"
                ;;
            *create*|*Create*|*write*)
                echo "   [$line_num] 📝 FILE_OP: $content"
                ;;
            *backend*|*frontend*|*database*|*docker*)
                echo "   [$line_num] 🏗️  COMPONENT: $content"
                ;;
            *tool*|*Tool*)
                echo "   [$line_num] 🛠️  TOOL: $content"
                ;;
            *)
                echo "   [$line_num] ℹ️  INFO: $content"
                ;;
        esac
    done
    
    echo ""
    echo "📊 Full log available at: /tmp/agentry_multilang_test.log"
    echo "📊 Log size: $(wc -l < /tmp/agentry_multilang_test.log) lines"
    
    # Show summary of coordination flow
    echo ""
    echo "🔄 COORDINATION FLOW SUMMARY:"
    if grep -q "team_status\|project_tree" /tmp/agentry_multilang_test.log; then
        echo "   ✅ Agent 0 gathered project context"
    else
        echo "   ❌ No context gathering detected"
    fi
    
    if grep -q "agent.*backend\|delegate.*python" /tmp/agentry_multilang_test.log; then
        echo "   ✅ Backend development delegated"
    else
        echo "   ⚠️  Backend delegation unclear"
    fi
    
    if grep -q "agent.*frontend\|delegate.*javascript" /tmp/agentry_multilang_test.log; then
        echo "   ✅ Frontend development delegated"
    else
        echo "   ⚠️  Frontend delegation unclear"
    fi
    
    if grep -q "agent.*database\|delegate.*sql" /tmp/agentry_multilang_test.log; then
        echo "   ✅ Database work delegated"
    else
        echo "   ⚠️  Database delegation unclear"
    fi
    
else
    echo "❌ No coordination log available"
fi

echo ""

# Safety Check
echo "🛡️  Safety Check: Verifying project isolation..."
cd "$PROJECT_DIR"
if git status --porcelain | grep -q .; then
    echo "⚠️  WARNING: Project files may have been modified during test!"
    echo "   Modified files:"
    git status --porcelain | sed 's/^/   /'
else
    echo "✅ Project files unchanged - isolation working correctly"
fi
cd "$AI_WORKSPACE"

echo ""
echo "=== PRIORITY 1 TEST COMPLETE ==="
echo "Next: Implement Priority 2 (Parallel vs Sequential Coordination)"

# Store results for trend analysis
mkdir -p /tmp/agentry-test-results
echo "$(date '+%Y-%m-%d %H:%M:%S'),multilang_coordination,$TOTAL_SCORE,$COORDINATION_SCORE,$TECH_COVERAGE,$DELEGATION_COUNT" >> /tmp/agentry-test-results/coordination_trends.csv

# Cleanup
echo ""
echo "🧹 Cleaning up isolated workspace..."
cd "$PROJECT_DIR"
rm -rf "$AI_WORKSPACE"
rm -f /tmp/test_input.txt
echo "✅ Isolated workspace cleaned up automatically"

echo ""
if [ $TOTAL_SCORE -ge 70 ]; then
    echo "🎯 READY for Priority 2: Parallel vs Sequential Coordination"
    exit 0
else
    echo "⚠️  Consider improving coordination before advancing to Priority 2"
    exit 1
fi
