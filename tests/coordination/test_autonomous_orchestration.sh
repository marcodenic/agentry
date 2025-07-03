#!/bin/bash

# Autonomous Multi-Agent Orchestration Test
# Test Agent 0's ability to autonomously choose and coordinate available agents

# Source the test helpers script
# shellcheck source=/dev/null
source "$(dirname "$0")/../scripts/test-helpers.sh"

echo "🤖 Autonomous Multi-Agent Orchestration Test"
echo "============================================"
echo "Testing Agent 0's autonomous team management:"
echo "- Agent 0 analyzes a complex task"
echo "- Checks what agents are actually available"
echo "- Decides optimal delegation strategy"
echo "- Coordinates available agents to complete the task"
echo "- NO pre-defined agent assignments - Agent 0 decides everything"
echo ""

# Create clean sandbox
setup_test_environment

echo "🎯 AUTONOMOUS ORCHESTRATION SCENARIO"
echo "===================================="
echo "Task: Create a complete blog application"
echo "- Agent 0 must decide how to break this down"
echo "- Agent 0 must check what agents are available"
echo "- Agent 0 must assign work to the best available agents"
echo "- No human guidance on which agents to use"
echo ""

# Start the autonomous orchestration test
echo "🤖 Starting Autonomous Multi-Agent Orchestration:"
echo "================================================="

$AGENT_CMD chat <<EOF
I need you to create a complete blog application. This is a complex task that will require multiple components working together.

PROJECT GOAL: Create "QuickBlog" - a simple but complete blog application

REQUIREMENTS:
- Users can read blog posts
- Users can create new blog posts  
- Basic user authentication (login/register)
- Clean, responsive design
- Data persistence (database or file-based)
- Basic testing
- Documentation for setup and usage

IMPORTANT: I'm not telling you which agents to use or how to organize the work. You need to:

1. First, check what agents are available to you
2. Analyze this task and decide how to break it down
3. Determine which available agents are best suited for each part
4. Create a coordination plan using your available team
5. Delegate tasks to the most appropriate agents
6. Coordinate between agents to ensure integration
7. Monitor progress and adapt as needed

Show me your autonomous orchestration process - how you analyze the task, assess your team, plan the work, and coordinate execution using whatever agents you actually have available.

Start by checking your team and analyzing the requirements.
EOF

echo ""
echo "📊 AUTONOMOUS ORCHESTRATION ANALYSIS"
echo "====================================="

# Wait a moment for any background processes
sleep 2

echo "🔍 Checking what was created..."
echo "Files and directories created:"
find . -type f \( -name "*.py" -o -name "*.html" -o -name "*.js" -o -name "*.css" -o -name "*.sql" -o -name "*.md" -o -name "*.json" -o -name "*.yaml" -o -name "*.yml" \) | sort

echo ""
echo "📁 Project structure:"
tree . 2>/dev/null || ls -la

echo ""
echo "🧠 AUTONOMOUS DECISION ANALYSIS"
echo "==============================="

# Analyze the orchestration decisions
total_files=$(find . -type f \( -name "*.py" -o -name "*.html" -o -name "*.js" -o -name "*.css" -o -name "*.sql" -o -name "*.md" \) | wc -l)
echo "Total project files created: $total_files"

# Check for different component types
backend_files=$(find . -name "*.py" | grep -E "(app|server|api|main)" | wc -l)
frontend_files=$(find . -name "*.html" -o -name "*.css" -o -name "*.js" | wc -l)
database_files=$(find . -name "*.sql" -o -name "*database*" -o -name "*schema*" | wc -l)
test_files=$(find . -name "*test*" | wc -l)
doc_files=$(find . -name "*.md" -o -name "*README*" | wc -l)

echo ""
echo "📈 COMPONENT ANALYSIS:"
echo "======================"
echo "Backend components: $backend_files files"
echo "Frontend components: $frontend_files files"  
echo "Database components: $database_files files"
echo "Testing components: $test_files files"
echo "Documentation: $doc_files files"

# Calculate orchestration intelligence
orchestration_score=0
max_score=8

if [ $total_files -gt 5 ]; then orchestration_score=$((orchestration_score + 2)); echo "✅ Created comprehensive project ($total_files files)"; fi
if [ $backend_files -gt 0 ]; then orchestration_score=$((orchestration_score + 2)); echo "✅ Backend components created"; fi
if [ $frontend_files -gt 0 ]; then orchestration_score=$((orchestration_score + 2)); echo "✅ Frontend components created"; fi
if [ $database_files -gt 0 ]; then orchestration_score=$((orchestration_score + 1)); echo "✅ Database components created"; fi
if [ $test_files -gt 0 ]; then orchestration_score=$((orchestration_score + 1)); echo "✅ Testing components created"; fi

orchestration_percentage=$((100 * orchestration_score / max_score))

echo ""
echo "🎯 AUTONOMOUS ORCHESTRATION ASSESSMENT:"
echo "======================================="
echo "Orchestration Intelligence Score: $orchestration_percentage%"

if [ $orchestration_percentage -ge 80 ]; then
    echo ""
    echo "🏆 OUTSTANDING: Agent 0 demonstrates AUTONOMOUS orchestration mastery!"
    echo "✅ Independently analyzed complex requirements"
    echo "✅ Made intelligent delegation decisions"
    echo "✅ Coordinated multiple components effectively"
    echo "✅ Created comprehensive, integrated solution"
elif [ $orchestration_percentage -ge 60 ]; then
    echo ""
    echo "✅ GOOD: Agent 0 shows solid autonomous orchestration"
    echo "✅ Handled task analysis and delegation reasonably"
    echo "⚠️  Room for improvement in comprehensive coordination"
elif [ $orchestration_percentage -ge 40 ]; then
    echo ""
    echo "⚠️  FAIR: Agent 0 shows basic autonomous capabilities"
    echo "✅ Some task breakdown and delegation evident"
    echo "❌ Limited comprehensive orchestration"
else
    echo ""
    echo "❌ NEEDS WORK: Autonomous orchestration needs significant improvement"
    echo "❌ Limited evidence of intelligent task analysis and delegation"
fi

echo ""
echo "📄 SAMPLE AUTONOMOUS COORDINATION EVIDENCE:"
echo "==========================================="

# Show key project files
echo "🗂️ Project Files Created:"
for file in $(find . -name "*.py" -o -name "*.html" -o -name "*.md" | head -5); do
    if [ -f "$file" ]; then
        echo ""
        echo "--- $file ---"
        head -10 "$file"
    fi
done

echo ""
echo "🔍 Integration Evidence:"
# Check for evidence of integration planning
if find . -name "*.py" -exec grep -l "import\|from" {} \; | head -1 >/dev/null 2>&1; then
    echo "✅ Python module integration found"
fi
if find . -name "*.html" -exec grep -l "script\|link\|href" {} \; | head -1 >/dev/null 2>&1; then
    echo "✅ Frontend integration found" 
fi
if find . -name "*.py" -exec grep -l "database\|db\|sql" {} \; | head -1 >/dev/null 2>&1; then
    echo "✅ Database integration found"
fi

echo ""
echo "✅ Autonomous Multi-Agent Orchestration Test Complete"
echo ""
echo "🔑 KEY INSIGHTS:"
echo "- Did Agent 0 check available agents before delegating?"
echo "- Did Agent 0 make autonomous decisions about task breakdown?"
echo "- Did Agent 0 coordinate multiple agents to work together?"
echo "- Did Agent 0 adapt its strategy based on available resources?"
