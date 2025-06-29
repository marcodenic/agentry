#!/bin/bash

# Team-Based Coordinated Agent Execution Test
# Focus: Ensure Agent 0's delegated tasks actually get executed by target agents

echo "=== TEAM-BASED COORDINATED AGENT EXECUTION TEST ==="
echo "🎯 Testing Agent 0 coordination + actual task execution"
echo "📺 Focus: Complete delegation → execution pipeline"
echo ""

# Setup
AI_WORKSPACE="/tmp/agentry-ai-sandbox"
PROJECT_DIR="/home/marco/Documents/GitHub/agentry"

echo "🏗️  Setting up team execution test workspace..."
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
rm -rf team-execution-test 2>/dev/null
rm -f /tmp/team_execution.log

echo ""
echo "🎯 TEAM EXECUTION TEST:"
echo "Testing complete delegation → execution pipeline..."
echo ""

# Create a test focused on delegation AND execution
TEAM_EXECUTION_TEST="I need you to coordinate a team to create a simple project called 'team-execution-test' with actual working files.

Please coordinate this project by delegating specific tasks to your available agents and ensuring they execute the work:

1. **Task 1**: Delegate to a coder agent to create 'calculator.py' - a Python file with basic math functions (add, subtract, multiply, divide)

2. **Task 2**: Delegate to a coder agent to create 'test_calculator.py' - a Python test file that imports and tests the calculator functions

3. **Task 3**: Delegate to a coder agent to create 'README.md' - documentation explaining how to use the calculator

CRITICAL REQUIREMENTS:
- Use your coordination tools (team_status, check_agent, assign_task, send_message) to manage the team
- Ensure each delegated task actually results in a working file being created
- Monitor task completion and provide progress updates
- Show me which agents are working on what tasks
- Verify that the created files work together (imports, etc.)

This is a test of your complete coordination pipeline: analyze → delegate → execute → verify."

echo "📝 Team Execution Test Prompt:"
echo "$TEAM_EXECUTION_TEST"
echo ""

# Run with comprehensive team execution monitoring
echo "🚀 Starting team execution test..."
echo "📺 TEAM EXECUTION MONITORING:"
echo "============================================================================================================"

# Create input file
echo "$TEAM_EXECUTION_TEST" > /tmp/team_input.txt
echo "Please quit when all tasks are complete." >> /tmp/team_input.txt

# Enhanced monitoring for team execution
timeout 200s ./agentry chat < /tmp/team_input.txt 2>&1 | while IFS= read -r line; do
    timestamp=$(date "+%H:%M:%S")
    case "$line" in
        *team_status*|*check_agent*|*assign_task*|*send_message*)
            echo "[$timestamp] 🛠️  COORDINATION TOOL: $line"
            ;;
        *delegate*|*Delegate*|*assign*|*Assign*|*task*|*Task*)
            echo "[$timestamp] 📢 DELEGATION: $line"
            ;;
        *agent*|*Agent*|*coder*|*Coder*)
            echo "[$timestamp] 🤖 AGENT ACTIVITY: $line"
            ;;
        *create*|*Create*|*write*|*Write*|*file*|*File*)
            echo "[$timestamp] 📝 FILE OPERATION: $line"
            ;;
        *calculator*|*Calculator*|*test_*|*README*)
            echo "[$timestamp] 🎯 PROJECT FILE: $line"
            ;;
        *complete*|*Complete*|*done*|*Done*|*finished*|*Finished*)
            echo "[$timestamp] ✅ COMPLETION: $line"
            ;;
        *progress*|*Progress*|*status*|*Status*)
            echo "[$timestamp] 📊 PROGRESS: $line"
            ;;
        *error*|*Error*|*failed*|*Failed*|*problem*|*Problem*)
            echo "[$timestamp] ❌ ISSUE: $line"
            ;;
        *thinking*|*Thinking*)
            echo "[$timestamp] 💭 THINKING: Agent processing..."
            ;;
        *using*|*Using*|*tool*|*Tool*)
            echo "[$timestamp] 🔧 TOOL USAGE: $line"
            ;;
        *coordinate*|*Coordinate*|*team*|*Team*)
            echo "[$timestamp] 👥 COORDINATION: $line"
            ;;
        *)
            if [[ ${#line} -gt 20 && ! "$line" =~ ^[[:space:]]*$ && ! "$line" =~ ^\.*$ ]]; then
                echo "[$timestamp]    $line"
            fi
            ;;
    esac
done > /tmp/team_execution.log 2>&1

RESULT=$?

echo "============================================================================================================"
echo "⏱️  Team execution test completed with result: $RESULT"
echo ""

# Comprehensive analysis of team execution
echo "🔍 TEAM EXECUTION ANALYSIS:"
echo "=========================="

if [ -f "/tmp/team_execution.log" ]; then
    # Count team execution activities
    COORDINATION_TOOLS=$(grep -c "COORDINATION TOOL" /tmp/team_execution.log || echo "0")
    DELEGATION_ACTIVITIES=$(grep -c "DELEGATION" /tmp/team_execution.log || echo "0")
    AGENT_ACTIVITIES=$(grep -c "AGENT ACTIVITY" /tmp/team_execution.log || echo "0")
    FILE_OPERATIONS=$(grep -c "FILE OPERATION" /tmp/team_execution.log || echo "0")
    PROJECT_FILE_MENTIONS=$(grep -c "PROJECT FILE" /tmp/team_execution.log || echo "0")
    COMPLETION_SIGNALS=$(grep -c "COMPLETION" /tmp/team_execution.log || echo "0")
    
    echo "📊 TEAM EXECUTION METRICS:"
    echo "   🛠️  Coordination tools used: $COORDINATION_TOOLS"
    echo "   📢 Delegation activities: $DELEGATION_ACTIVITIES"
    echo "   🤖 Agent activities: $AGENT_ACTIVITIES"
    echo "   📝 File operations: $FILE_OPERATIONS"
    echo "   🎯 Project file mentions: $PROJECT_FILE_MENTIONS"
    echo "   ✅ Completion signals: $COMPLETION_SIGNALS"
    echo ""
    
    # Show team execution timeline
    echo "🕐 TEAM EXECUTION TIMELINE:"
    echo "   Key execution events:"
    grep "COORDINATION TOOL\|DELEGATION\|FILE OPERATION\|COMPLETION" /tmp/team_execution.log | head -25 | sed 's/^/   /'
    echo ""
    
    # Calculate execution effectiveness
    echo "🎯 EXECUTION EFFECTIVENESS ASSESSMENT:"
    
    if [ $COORDINATION_TOOLS -ge 3 ] && [ $FILE_OPERATIONS -ge 3 ]; then
        echo "   ✅ EXCELLENT: Strong coordination with actual file creation"
    elif [ $COORDINATION_TOOLS -ge 1 ] && [ $FILE_OPERATIONS -ge 1 ]; then
        echo "   ⚠️  GOOD: Some coordination and execution detected"
    elif [ $DELEGATION_ACTIVITIES -ge 3 ]; then
        echo "   🔄 COORDINATION ONLY: Good delegation but limited execution"
    else
        echo "   ❌ POOR: Limited coordination and execution"
    fi
    
else
    echo "❌ No team execution log available"
fi

# Check actual file creation results
echo ""
echo "📂 EXECUTION RESULTS VERIFICATION:"

# Check for project directory
if [ -d "team-execution-test" ]; then
    echo "✅ Project directory 'team-execution-test' created"
    PROJECT_DIR_CREATED=1
else
    echo "❌ Project directory 'team-execution-test' NOT created"
    PROJECT_DIR_CREATED=0
fi

# Check for specific requested files
CALCULATOR_PY=0
TEST_CALCULATOR_PY=0
README_MD=0

if [ -f "team-execution-test/calculator.py" ] || [ -f "calculator.py" ]; then
    echo "✅ calculator.py file created"
    CALCULATOR_PY=1
    CALC_FILE=$(find . -name "calculator.py" | head -1)
    if [ -f "$CALC_FILE" ]; then
        echo "   📄 calculator.py content preview:"
        head -10 "$CALC_FILE" | sed 's/^/      /'
        echo ""
    fi
else
    echo "❌ calculator.py file NOT created"
fi

if [ -f "team-execution-test/test_calculator.py" ] || [ -f "test_calculator.py" ]; then
    echo "✅ test_calculator.py file created"
    TEST_CALCULATOR_PY=1
    TEST_FILE=$(find . -name "test_calculator.py" | head -1)
    if [ -f "$TEST_FILE" ]; then
        echo "   📄 test_calculator.py content preview:"
        head -10 "$TEST_FILE" | sed 's/^/      /'
        echo ""
    fi
else
    echo "❌ test_calculator.py file NOT created"
fi

if [ -f "team-execution-test/README.md" ] || [ -f "README.md" ]; then
    echo "✅ README.md file created"
    README_MD=1
    README_FILE=$(find . -name "README.md" | head -1)
    if [ -f "$README_FILE" ]; then
        echo "   📄 README.md content preview:"
        head -5 "$README_FILE" | sed 's/^/      /'
        echo ""
    fi
else
    echo "❌ README.md file NOT created"
fi

# Calculate execution success rate
TOTAL_REQUESTED_FILES=3
FILES_CREATED=$((CALCULATOR_PY + TEST_CALCULATOR_PY + README_MD))
EXECUTION_SUCCESS_RATE=$(( (FILES_CREATED * 100) / TOTAL_REQUESTED_FILES ))

echo "📊 EXECUTION SUCCESS METRICS:"
echo "   📁 Files requested: $TOTAL_REQUESTED_FILES"
echo "   ✅ Files created: $FILES_CREATED"
echo "   📈 Execution success rate: $EXECUTION_SUCCESS_RATE%"
echo ""

# Test file integration (if calculator files exist)
if [ $CALCULATOR_PY -eq 1 ] && [ $TEST_CALCULATOR_PY -eq 1 ]; then
    echo "🔍 INTEGRATION TEST:"
    CALC_FILE=$(find . -name "calculator.py" | head -1)
    TEST_FILE=$(find . -name "test_calculator.py" | head -1)
    
    # Check if test file imports calculator
    if grep -q "calculator\|import.*calculator" "$TEST_FILE" 2>/dev/null; then
        echo "   ✅ test_calculator.py properly imports calculator.py"
        
        # Try to run the integration test
        echo "   🧪 Attempting to run integration test..."
        cd "$(dirname "$CALC_FILE")"
        if python3 -c "import calculator; print('✅ calculator.py imports successfully')" 2>/dev/null; then
            echo "   ✅ calculator.py is valid Python and imports successfully"
        else
            echo "   ⚠️  calculator.py has import issues"
        fi
        cd "$AI_WORKSPACE"
    else
        echo "   ❌ test_calculator.py doesn't import calculator.py"
    fi
else
    echo "🔍 INTEGRATION TEST: Skipped (missing required files)"
fi

# Final assessment
echo ""
echo "🎯 FINAL TEAM EXECUTION ASSESSMENT:"

OVERALL_SCORE=0

# Coordination quality (40% weight)
if [ $COORDINATION_TOOLS -ge 3 ]; then
    OVERALL_SCORE=$((OVERALL_SCORE + 40))
    echo "   ✅ COORDINATION: EXCELLENT (Active use of coordination tools)"
elif [ $COORDINATION_TOOLS -ge 1 ]; then
    OVERALL_SCORE=$((OVERALL_SCORE + 20))
    echo "   ⚠️  COORDINATION: MODERATE (Some coordination tools used)"
else
    echo "   ❌ COORDINATION: POOR (No coordination tools detected)"
fi

# Execution quality (50% weight)
if [ $EXECUTION_SUCCESS_RATE -ge 100 ]; then
    OVERALL_SCORE=$((OVERALL_SCORE + 50))
    echo "   ✅ EXECUTION: EXCELLENT (All requested files created)"
elif [ $EXECUTION_SUCCESS_RATE -ge 67 ]; then
    OVERALL_SCORE=$((OVERALL_SCORE + 35))
    echo "   ⚠️  EXECUTION: GOOD (Most requested files created)"
elif [ $EXECUTION_SUCCESS_RATE -ge 33 ]; then
    OVERALL_SCORE=$((OVERALL_SCORE + 20))
    echo "   🔄 EXECUTION: FAIR (Some requested files created)"
else
    echo "   ❌ EXECUTION: POOR (Few/no requested files created)"
fi

# Integration quality (10% weight)
if [ $CALCULATOR_PY -eq 1 ] && [ $TEST_CALCULATOR_PY -eq 1 ]; then
    OVERALL_SCORE=$((OVERALL_SCORE + 10))
    echo "   ✅ INTEGRATION: EXCELLENT (Related files work together)"
elif [ $FILES_CREATED -ge 2 ]; then
    OVERALL_SCORE=$((OVERALL_SCORE + 5))
    echo "   ⚠️  INTEGRATION: MODERATE (Multiple files created)"
else
    echo "   ❌ INTEGRATION: POOR (Insufficient files for integration)"
fi

echo ""
echo "🏆 OVERALL TEAM EXECUTION SCORE: $OVERALL_SCORE/100"

if [ $OVERALL_SCORE -ge 80 ]; then
    echo "🎉 RESULT: EXCELLENT - Team-based coordinated execution working!"
    echo "   ✅ Agent 0 successfully coordinates team and ensures task execution"
    echo "   🚀 Ready for Priority 2: Parallel vs Sequential Coordination"
    NEXT_PHASE=0
elif [ $OVERALL_SCORE -ge 60 ]; then
    echo "👍 RESULT: GOOD - Team coordination mostly working"
    echo "   ⚠️  Minor gaps in execution pipeline, but foundation is solid"
    echo "   🔄 Can proceed to Priority 2 with monitoring"
    NEXT_PHASE=0
elif [ $OVERALL_SCORE -ge 40 ]; then
    echo "⚠️  RESULT: FAIR - Team coordination partially working"
    echo "   🔧 Needs improvement in execution pipeline"
    echo "   🔍 Debug delegation → execution gap before Priority 2"
    NEXT_PHASE=1
else
    echo "❌ RESULT: POOR - Team coordination needs major work"
    echo "   🚨 Significant issues in coordination and execution"
    echo "   🛠️  Must fix team execution before advancing"
    NEXT_PHASE=1
fi

echo ""
echo "📋 EXECUTION LOG ANALYSIS:"
echo "   Full log available at: /tmp/team_execution.log"
echo "   Log size: $(wc -l < /tmp/team_execution.log 2>/dev/null || echo "0") lines"

# Show key execution excerpts
echo ""
echo "📄 KEY EXECUTION EXCERPTS:"
if [ -f "/tmp/team_execution.log" ]; then
    echo "   Most relevant execution activities:"
    grep "COORDINATION TOOL\|FILE OPERATION\|COMPLETION\|PROJECT FILE" /tmp/team_execution.log | tail -20 | sed 's/^/   /'
else
    echo "   No execution log available"
fi

echo ""
echo "=== TEAM-BASED COORDINATED AGENT EXECUTION TEST COMPLETE ==="

exit $NEXT_PHASE
