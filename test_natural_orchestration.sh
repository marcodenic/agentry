#!/bin/bash

# Natural Language Agent Orchestration Test
# Focus: Agent 0 using natural language to coordinate and delegate tasks
# NO /commands - pure natural language coordination

echo "=== NATURAL LANGUAGE AGENT ORCHESTRATION TEST ==="
echo "🎯 Testing Agent 0's natural language coordination and delegation"
echo "📺 Focus: Natural multi-agent collaboration without /commands"
echo ""

# Setup
AI_WORKSPACE="/tmp/agentry-ai-sandbox"
PROJECT_DIR="/home/marco/Documents/GitHub/agentry"

echo "🏗️  Setting up test workspace..."
mkdir -p "$AI_WORKSPACE"
cd "$AI_WORKSPACE"

# Copy configuration and executable
cp "$PROJECT_DIR/.agentry.yaml" . 2>/dev/null || echo "⚠️  No .agentry.yaml found"
if [ -f "$PROJECT_DIR/.env.local" ]; then
    cp "$PROJECT_DIR/.env.local" .
    echo "✅ Copied .env.local"
else
    echo "⚠️  No .env.local found"
fi

# Copy agentry executable
if [ -f "$PROJECT_DIR/agentry" ]; then
    cp "$PROJECT_DIR/agentry" .
    chmod +x agentry
    echo "✅ Copied agentry executable"
elif [ -f "$PROJECT_DIR/agentry.exe" ]; then
    cp "$PROJECT_DIR/agentry.exe" ./agentry
    chmod +x agentry
    echo "✅ Copied agentry.exe as agentry"
else
    echo "❌ No agentry executable found!"
    exit 1
fi

echo "📁 Working in: $(pwd)"

# Clean test environment
rm -rf natural-orchestration-test 2>/dev/null
rm -f /tmp/natural_orchestration.log

echo ""
echo "🎯 NATURAL LANGUAGE ORCHESTRATION TEST:"
echo "Testing Agent 0's ability to coordinate through natural language..."
echo ""

# Create a test that requires natural coordination and delegation
NATURAL_ORCHESTRATION_TEST="I need you to coordinate a development team to create a simple web application called 'natural-orchestration-test'. 

This project requires coordination between different types of developers:

1. I need a backend developer to create a Python Flask API file called 'backend_api.py' with endpoints for user management
2. I need a frontend developer to create an HTML file called 'frontend.html' that interacts with the API  
3. I need a database specialist to create a SQL schema file called 'database_schema.sql' for user data
4. I need a DevOps engineer to create a 'deployment.yaml' file for container deployment

Please coordinate this work by:
- Analyzing what team members you need
- Checking what agents are available to help
- Delegating specific tasks to appropriate agents
- Ensuring the different components work together
- Providing updates on team progress

Show me your coordination process - I want to see how you organize the team and delegate the work naturally."

echo "📝 Natural Language Test Prompt:"
echo "$NATURAL_ORCHESTRATION_TEST"
echo ""

# Run with natural language orchestration monitoring
echo "🚀 Starting natural language orchestration test..."
echo "📺 NATURAL LANGUAGE COORDINATION MONITORING:"
echo "============================================================================================================"

# Create input file
echo "$NATURAL_ORCHESTRATION_TEST" > /tmp/natural_input.txt
echo "Thank you. Please quit when done." >> /tmp/natural_input.txt

# Run agentry and monitor for natural coordination patterns
timeout 150s ./agentry chat < /tmp/natural_input.txt 2>&1 | while IFS= read -r line; do
    timestamp=$(date "+%H:%M:%S")
    case "$line" in
        *team*|*Team*|*coordinate*|*Coordinate*)
            echo "[$timestamp] 👥 TEAM COORDINATION: $line"
            ;;
        *delegate*|*Delegate*|*assign*|*Assign*)
            echo "[$timestamp] 📢 DELEGATION: $line"
            ;;
        *backend*|*Backend*|*frontend*|*Frontend*|*database*|*Database*|*devops*|*DevOps*)
            echo "[$timestamp] 🎯 ROLE ASSIGNMENT: $line"
            ;;
        *agent*|*Agent*|*developer*|*Developer*)
            echo "[$timestamp] 🤖 AGENT ACTIVITY: $line"
            ;;
        *team_status*|*check_agent*|*send_message*|*assign_task*)
            echo "[$timestamp] 🛠️  COORDINATION TOOL: $line"
            ;;
        *create*|*Create*|*file*|*File*|*\.py*|*\.html*|*\.sql*|*\.yaml*)
            echo "[$timestamp] 📝 FILE OPERATION: $line"
            ;;
        *analyze*|*Analyze*|*need*|*Need*|*require*|*Require*)
            echo "[$timestamp] 🧠 PLANNING: $line"
            ;;
        *progress*|*Progress*|*status*|*Status*|*complete*|*Complete*)
            echo "[$timestamp] 📊 PROGRESS: $line"
            ;;
        *thinking*|*Thinking*)
            echo "[$timestamp] 💭 THINKING: Agent processing coordination..."
            ;;
        *using*|*Using*|*tool*|*Tool*)
            echo "[$timestamp] 🔧 TOOL USAGE: $line"
            ;;
        *error*|*Error*|*failed*|*Failed*)
            echo "[$timestamp] ❌ ERROR: $line"
            ;;
        *)
            if [[ ${#line} -gt 15 && ! "$line" =~ ^[[:space:]]*$ && ! "$line" =~ ^\.*$ ]]; then
                echo "[$timestamp]    $line"
            fi
            ;;
    esac
done > /tmp/natural_orchestration.log 2>&1

RESULT=$?

echo "============================================================================================================"
echo "⏱️  Natural orchestration test completed with result: $RESULT"
echo ""

# Analyze the natural coordination patterns
echo "🔍 NATURAL COORDINATION ANALYSIS:"
echo "================================="

if [ -f "/tmp/natural_orchestration.log" ]; then
    # Count natural coordination activities
    TEAM_COORDINATION=$(grep -c "TEAM COORDINATION\|team.*coordinate\|coordinate.*team" /tmp/natural_orchestration.log || echo "0")
    DELEGATION_ACTIVITIES=$(grep -c "DELEGATION\|delegate\|assign.*agent\|agent.*task" /tmp/natural_orchestration.log || echo "0")
    ROLE_ASSIGNMENTS=$(grep -c "ROLE ASSIGNMENT\|backend.*developer\|frontend.*developer\|database.*specialist" /tmp/natural_orchestration.log || echo "0")
    COORDINATION_TOOLS=$(grep -c "COORDINATION TOOL\|team_status\|check_agent\|send_message\|assign_task" /tmp/natural_orchestration.log || echo "0")
    PLANNING_ACTIVITIES=$(grep -c "PLANNING\|analyze\|need.*developer\|require.*specialist" /tmp/natural_orchestration.log || echo "0")
    
    echo "📊 NATURAL COORDINATION METRICS:"
    echo "   👥 Team coordination mentions: $TEAM_COORDINATION"
    echo "   📢 Delegation activities: $DELEGATION_ACTIVITIES"
    echo "   🎯 Role assignments: $ROLE_ASSIGNMENTS"
    echo "   🛠️  Coordination tools used: $COORDINATION_TOOLS"
    echo "   🧠 Planning activities: $PLANNING_ACTIVITIES"
    echo ""
    
    # Show natural coordination timeline
    echo "🕐 NATURAL COORDINATION TIMELINE:"
    echo "   Key coordination events:"
    grep "TEAM COORDINATION\|DELEGATION\|ROLE ASSIGNMENT\|COORDINATION TOOL\|PLANNING" /tmp/natural_orchestration.log | head -20 | sed 's/^/   /'
    echo ""
    
    # Check for specific natural language coordination patterns
    NATURAL_DELEGATION=$(grep -c "I need.*developer\|assign.*to\|delegate.*task\|coordinate.*team" /tmp/natural_orchestration.log || echo "0")
    AGENT_COMMUNICATION=$(grep -c "send_message\|check_agent\|team_status" /tmp/natural_orchestration.log || echo "0")
    
    echo "🗣️  NATURAL LANGUAGE PATTERNS:"
    echo "   💬 Natural delegation language: $NATURAL_DELEGATION instances"
    echo "   📡 Agent communication tools: $AGENT_COMMUNICATION uses"
    
    # Assess coordination quality
    echo ""
    echo "🎯 NATURAL COORDINATION ASSESSMENT:"
    
    TOTAL_COORDINATION=$((TEAM_COORDINATION + DELEGATION_ACTIVITIES + ROLE_ASSIGNMENTS + COORDINATION_TOOLS))
    
    if [ $TOTAL_COORDINATION -ge 8 ] && [ $AGENT_COMMUNICATION -ge 2 ]; then
        echo "   ✅ EXCELLENT: Strong natural language coordination with tool usage"
    elif [ $TOTAL_COORDINATION -ge 4 ] && [ $COORDINATION_TOOLS -ge 1 ]; then
        echo "   ⚠️  GOOD: Moderate coordination with some tool usage"
    elif [ $TOTAL_COORDINATION -ge 2 ]; then
        echo "   🔄 BASIC: Some coordination detected, needs improvement"
    else
        echo "   ❌ POOR: Limited natural coordination patterns"
    fi
    
    # Check if Agent 0 is actually using coordination tools
    if [ $COORDINATION_TOOLS -ge 2 ]; then
        echo "   ✅ Agent 0 is actively using coordination tools (team_status, check_agent, etc.)"
    elif [ $COORDINATION_TOOLS -eq 1 ]; then
        echo "   ⚠️  Agent 0 used some coordination tools"
    else
        echo "   ❌ Agent 0 not using coordination tools - may be working directly"
    fi
    
else
    echo "❌ No coordination log available"
fi

# Check if coordinated files were created
echo ""
echo "📂 COORDINATION RESULTS:"
if [ -d "natural-orchestration-test" ]; then
    echo "✅ Project directory created through coordination"
    
    # Check for the specific files we requested
    BACKEND_FILES=$(find natural-orchestration-test -name "*backend*" -o -name "*api*" | wc -l)
    FRONTEND_FILES=$(find natural-orchestration-test -name "*frontend*" -o -name "*.html" | wc -l)
    DATABASE_FILES=$(find natural-orchestration-test -name "*database*" -o -name "*schema*" -o -name "*.sql" | wc -l)
    DEPLOYMENT_FILES=$(find natural-orchestration-test -name "*deploy*" -o -name "*.yaml" -o -name "*.yml" | wc -l)
    
    echo "📁 Coordinated deliverables:"
    echo "   🐍 Backend files: $BACKEND_FILES"
    echo "   🌐 Frontend files: $FRONTEND_FILES"
    echo "   🗄️  Database files: $DATABASE_FILES"
    echo "   🚀 Deployment files: $DEPLOYMENT_FILES"
    
    TOTAL_DELIVERABLES=$((BACKEND_FILES + FRONTEND_FILES + DATABASE_FILES + DEPLOYMENT_FILES))
    
    if [ $TOTAL_DELIVERABLES -ge 4 ]; then
        echo "   ✅ EXCELLENT: All requested deliverables created through coordination"
    elif [ $TOTAL_DELIVERABLES -ge 2 ]; then
        echo "   ⚠️  GOOD: Most deliverables created"
    elif [ $TOTAL_DELIVERABLES -ge 1 ]; then
        echo "   🔄 BASIC: Some deliverables created"
    else
        echo "   ❌ POOR: No clear deliverables from coordination"
    fi
    
    # Show what was actually created
    if [ $TOTAL_DELIVERABLES -gt 0 ]; then
        echo ""
        echo "📄 Files created through natural coordination:"
        find natural-orchestration-test -type f | while read file; do
            echo "   📄 $file ($(wc -l < "$file" 2>/dev/null || echo "0") lines):"
            head -3 "$file" 2>/dev/null | sed 's/^/      /' || echo "      [binary file]"
            echo ""
        done
    fi
    
else
    # Check for files in current directory
    LOOSE_FILES=$(find . -maxdepth 1 -name "*backend*" -o -name "*frontend*" -o -name "*database*" -o -name "*deploy*" | wc -l)
    if [ $LOOSE_FILES -gt 0 ]; then
        echo "⚠️  Files created in root directory (not organized):"
        find . -maxdepth 1 -name "*backend*" -o -name "*frontend*" -o -name "*database*" -o -name "*deploy*" | while read file; do
            echo "   📄 $file"
        done
    else
        echo "❌ No coordinated files created"
    fi
fi

echo ""
echo "📋 COORDINATION LOG ANALYSIS:"
echo "   Full log available at: /tmp/natural_orchestration.log"
echo "   Log size: $(wc -l < /tmp/natural_orchestration.log 2>/dev/null || echo "0") lines"

# Show key coordination excerpts
echo ""
echo "📄 KEY COORDINATION EXCERPTS:"
if [ -f "/tmp/natural_orchestration.log" ]; then
    echo "   Most relevant coordination activities:"
    grep "TEAM COORDINATION\|DELEGATION\|COORDINATION TOOL\|AGENT ACTIVITY" /tmp/natural_orchestration.log | tail -15 | sed 's/^/   /'
else
    echo "   No coordination log available"
fi

echo ""
echo "=== NATURAL LANGUAGE ORCHESTRATION TEST COMPLETE ==="
echo ""

# Determine if we're ready for next phase
if [ -f "/tmp/natural_orchestration.log" ]; then
    COORDINATION_SCORE=$((TEAM_COORDINATION + DELEGATION_ACTIVITIES + COORDINATION_TOOLS))
    
    if [ $COORDINATION_SCORE -ge 6 ]; then
        echo "🎯 READY: Strong natural coordination detected - proceed to Priority 2"
        exit 0
    elif [ $COORDINATION_SCORE -ge 3 ]; then
        echo "⚠️  MODERATE: Some coordination detected - may proceed with caution"
        exit 0
    else
        echo "🔧 NEEDS WORK: Limited coordination - investigate Agent 0's coordination capabilities"
        exit 1
    fi
else
    echo "❌ FAILED: No coordination data - check Agent 0 functionality"
    exit 1
fi
