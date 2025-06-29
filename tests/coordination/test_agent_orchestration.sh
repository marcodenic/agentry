#!/bin/bash

# Agent Orchestration Visibility Test
# Focus: Prove Agent 0 spawns/delegates to other agents with clear visibility

echo "=== AGENT ORCHESTRATION VISIBILITY TEST ==="
echo "🎯 Testing Agent 0's ability to spawn and delegate to other agents"
echo "📺 Focus: Clear visibility into multi-agent collaboration"
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
rm -rf orchestration-test 2>/dev/null
rm -f /tmp/agent_orchestration.log

echo "🎯 ORCHESTRATION TEST PROMPT:"
echo "We need to test explicit agent delegation and spawning..."
echo ""

# Create a test that FORCES agent orchestration
ORCHESTRATION_TEST="I need you to coordinate a team to create a simple project called 'orchestration-test'. 

IMPORTANT: You must demonstrate clear agent orchestration by:

1. First use /spawn coder to create a coder agent
2. Then use /list to show me the active agents  
3. Then /switch coder and have the coder create 'hello.py' with a simple function
4. Then /switch system (back to Agent 0) 
5. Use /spawn reviewer to create a reviewer agent
6. Have the reviewer agent check the hello.py file
7. Use /status to show me all agent statuses
8. Finally use /quit to exit

This is a test of your team coordination abilities. I want to see explicit agent spawning, switching, and delegation happening."

echo "📝 Orchestration Test Commands:"
echo "$ORCHESTRATION_TEST"
echo ""

# Run with explicit orchestration monitoring
echo "🚀 Starting orchestration test..."
echo "📺 AGENT ORCHESTRATION MONITORING:"
echo "============================================================================================================"

# Create input file with orchestration commands
echo "$ORCHESTRATION_TEST" > /tmp/orchestration_input.txt

# Run agentry and capture all output with timestamps
timeout 120s ./agentry chat < /tmp/orchestration_input.txt 2>&1 | while IFS= read -r line; do
    timestamp=$(date "+%H:%M:%S")
    case "$line" in
        *spawn*|*Spawn*|*spawning*|*Spawning*)
            echo "[$timestamp] 🤖 AGENT SPAWN: $line"
            ;;
        *switch*|*Switch*|*switching*|*Switching*)
            echo "[$timestamp] 🔄 AGENT SWITCH: $line"
            ;;
        *list*|*List*|*status*|*Status*)
            echo "[$timestamp] 📋 AGENT STATUS: $line"
            ;;
        *coder*|*Coder*|*reviewer*|*Reviewer*)
            echo "[$timestamp] 👥 AGENT ACTIVITY: $line"
            ;;
        *agent*|*Agent*)
            echo "[$timestamp] 🤖 AGENT INFO: $line"
            ;;
        *delegate*|*Delegate*|*assign*|*Assign*)
            echo "[$timestamp] 📢 DELEGATION: $line"
            ;;
        *create*|*Create*|*file*|*File*)
            echo "[$timestamp] 📝 FILE OPERATION: $line"
            ;;
        *tool*|*Tool*|*using*|*Using*)
            echo "[$timestamp] 🛠️  TOOL USAGE: $line"
            ;;
        *error*|*Error*|*failed*|*Failed*)
            echo "[$timestamp] ❌ ERROR: $line"
            ;;
        *Commands:*|*ready*|*Ready*)
            echo "[$timestamp] 🎮 SYSTEM: $line"
            ;;
        *thinking*|*Thinking*)
            echo "[$timestamp] 💭 THINKING: Agent processing..."
            ;;
        *)
            if [[ ${#line} -gt 20 && ! "$line" =~ ^[[:space:]]*$ && ! "$line" =~ ^\.*$ ]]; then
                echo "[$timestamp]    $line"
            fi
            ;;
    esac
done > /tmp/agent_orchestration.log 2>&1

RESULT=$?

echo "============================================================================================================"
echo "⏱️  Orchestration test completed with result: $RESULT"
echo ""

# Analyze the orchestration log for agent activities
echo "🔍 ORCHESTRATION ANALYSIS:"
echo "=========================="

if [ -f "/tmp/agent_orchestration.log" ]; then
    # Count different orchestration activities
    AGENT_SPAWNS=$(grep -c "AGENT SPAWN\|spawn.*coder\|spawn.*reviewer" /tmp/agent_orchestration.log || echo "0")
    AGENT_SWITCHES=$(grep -c "AGENT SWITCH\|switch.*coder\|switch.*system" /tmp/agent_orchestration.log || echo "0")
    AGENT_STATUS_CHECKS=$(grep -c "AGENT STATUS\|/list\|/status" /tmp/agent_orchestration.log || echo "0")
    DELEGATION_ACTIVITIES=$(grep -c "DELEGATION\|delegate\|assign" /tmp/agent_orchestration.log || echo "0")
    
    echo "📊 ORCHESTRATION METRICS:"
    echo "   🤖 Agent spawns detected: $AGENT_SPAWNS"
    echo "   🔄 Agent switches detected: $AGENT_SWITCHES" 
    echo "   📋 Status checks detected: $AGENT_STATUS_CHECKS"
    echo "   📢 Delegation activities: $DELEGATION_ACTIVITIES"
    echo ""
    
    # Show the orchestration timeline
    echo "🕐 ORCHESTRATION TIMELINE:"
    echo "   Key orchestration events in chronological order:"
    grep "AGENT SPAWN\|AGENT SWITCH\|AGENT STATUS\|DELEGATION" /tmp/agent_orchestration.log | head -15 | sed 's/^/   /'
    echo ""
    
    # Check if specific agents were mentioned
    CODER_MENTIONS=$(grep -c "coder\|Coder" /tmp/agent_orchestration.log || echo "0")
    REVIEWER_MENTIONS=$(grep -c "reviewer\|Reviewer" /tmp/agent_orchestration.log || echo "0")
    
    echo "👥 SPECIFIC AGENT TRACKING:"
    echo "   🔧 Coder agent mentions: $CODER_MENTIONS"
    echo "   🔍 Reviewer agent mentions: $REVIEWER_MENTIONS"
    
    # Determine orchestration quality
    echo ""
    echo "🎯 ORCHESTRATION ASSESSMENT:"
    
    if [ $AGENT_SPAWNS -ge 2 ] && [ $AGENT_SWITCHES -ge 2 ]; then
        echo "   ✅ EXCELLENT: Clear evidence of agent spawning and switching"
    elif [ $AGENT_SPAWNS -ge 1 ]; then
        echo "   ⚠️  MODERATE: Some agent spawning detected"
    else
        echo "   ❌ POOR: No clear agent spawning detected"
    fi
    
    if [ $DELEGATION_ACTIVITIES -ge 2 ]; then
        echo "   ✅ EXCELLENT: Strong delegation patterns"
    elif [ $DELEGATION_ACTIVITIES -ge 1 ]; then
        echo "   ⚠️  MODERATE: Some delegation detected"
    else
        echo "   ❌ POOR: Limited delegation patterns"
    fi
    
else
    echo "❌ No orchestration log available"
fi

# Check if files were actually created by the orchestrated agents
echo ""
echo "📂 ORCHESTRATION RESULTS:"
if [ -d "orchestration-test" ]; then
    echo "✅ Project directory created"
    
    FILES_FOUND=$(find orchestration-test -name "*.py" -o -name "*.txt" -o -name "*.md" | wc -l)
    echo "📁 Files created by orchestrated agents: $FILES_FOUND"
    
    if [ $FILES_FOUND -gt 0 ]; then
        echo ""
        echo "📄 Files created through orchestration:"
        find orchestration-test -type f | while read file; do
            echo "   📄 $file:"
            head -5 "$file" | sed 's/^/      /'
            echo ""
        done
    fi
else
    # Check for files in root directory
    DIRECT_FILES=$(find . -maxdepth 1 -name "hello.py" -o -name "*.py" -type f | wc -l)
    if [ $DIRECT_FILES -gt 0 ]; then
        echo "⚠️  Files created in root directory:"
        find . -maxdepth 1 -name "*.py" -type f | while read file; do
            echo "   📄 $file:"
            head -3 "$file" | sed 's/^/      /'
            echo ""
        done
    else
        echo "❌ No files created through orchestration"
    fi
fi

echo ""
echo "📋 FULL ORCHESTRATION LOG:"
echo "   Available at: /tmp/agent_orchestration.log"
echo "   Log size: $(wc -l < /tmp/agent_orchestration.log 2>/dev/null || echo "0") lines"

# Show a sample of the raw log for debugging
echo ""
echo "📄 SAMPLE ORCHESTRATION LOG (last 20 lines):"
tail -20 /tmp/agent_orchestration.log 2>/dev/null | sed 's/^/   /' || echo "   No log available"

echo ""
echo "=== AGENT ORCHESTRATION TEST COMPLETE ==="
echo ""
echo "🎯 NEXT: If orchestration is working, proceed to Priority 2 (Parallel Coordination)"
echo "🔧 If orchestration needs work, debug multi-agent communication patterns"
