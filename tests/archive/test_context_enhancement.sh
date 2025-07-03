#!/bin/bash

# Context Enhancement Test
# Test Agent 0's new VSCode-level context capabilities

echo "🧠 Testing Enhanced Context Capabilities"
echo "========================================"
echo "Testing Agent 0's new project_tree tool and context-aware coordination"
echo ""

# Test in the Agentry project itself for realistic context
cd /home/marco/Documents/GitHub/agentry

# Copy environment if needed
if [ -f ".env.local" ]; then
    echo "📋 Environment ready"
fi

echo "📁 Testing in Agentry project: $(pwd)"
echo ""

# Test function for context-aware scenarios
test_context_scenario() {
    local scenario_name="$1"
    local request="$2"
    local timeout_duration=${3:-90}
    
    echo "🧪 Context Test: $scenario_name"
    echo "📝 Request: $request"
    echo "----------------------------------------"
    
    # Create input
    echo "$request" > /tmp/context_input.txt
    echo "/quit" >> /tmp/context_input.txt
    
    # Run test
    local output_file="/tmp/context_output_$(date +%s).txt"
    echo "⏳ Running context test..."
    timeout $timeout_duration ./agentry.exe chat < /tmp/context_input.txt > "$output_file" 2>&1
    
    echo ""
    echo "📋 Agent 0 Response:"
    echo "--------------------"
    cat "$output_file"
    echo "--------------------"
    
    # Analyze context usage
    project_tree_usage=$(grep -c "project_tree\|📂.*Structure\|📁\|📄" "$output_file")
    context_tools_usage=$(grep -c "🔧 system is using a tool" "$output_file")
    
    echo ""
    echo "📊 Context Analysis:"
    echo "   🌳 Project tree usage indicators: $project_tree_usage"
    echo "   🔧 Total tool usage: $context_tools_usage"
    
    if [ $project_tree_usage -gt 0 ]; then
        echo "   ✅ CONTEXT AWARENESS DETECTED - Agent 0 used project structure"
    else
        echo "   ⚠️  Limited context awareness - may need prompt refinement"
    fi
    
    if [ $context_tools_usage -gt 2 ]; then
        echo "   ✅ RICH CONTEXT GATHERING - Multiple tools used"
    elif [ $context_tools_usage -gt 0 ]; then
        echo "   🟡 SOME CONTEXT - Basic tool usage"
    else
        echo "   ❌ NO CONTEXT TOOLS - Direct response only"
    fi
    
    echo ""
    echo "=================================================="
    echo ""
    sleep 2
}

# Test scenarios focused on context awareness
test_context_scenario \
    "Project Analysis Request" \
    "Can you analyze this codebase and tell me what kind of project this is? I want to understand the structure and what we're working with."

test_context_scenario \
    "Context-Aware Code Review" \
    "I need a code review of the main agent logic. Can you understand the project structure and get the right people to help review the core functionality?"

test_context_scenario \
    "Planning with Context" \
    "I want to add new features to this project. Can you analyze what we have and suggest what team we need to plan and implement improvements?"

# Test the project_tree tool directly
echo "🌳 Direct Project Tree Test"
echo "==========================="

echo "Testing project_tree tool directly..."
echo "project_tree" > /tmp/direct_tree_test.txt
echo "/quit" >> /tmp/direct_tree_test.txt

timeout 30 ./agentry.exe chat < /tmp/direct_tree_test.txt > /tmp/direct_tree_output.txt 2>&1

echo "📋 Direct project_tree output:"
echo "-----------------------------"
cat /tmp/direct_tree_output.txt
echo "-----------------------------"

# Check if project tree shows smart filtering
if grep -q "📁\|📄" /tmp/direct_tree_output.txt; then
    echo "✅ PROJECT TREE WORKING - Shows structured output with emojis"
else
    echo "⚠️  Project tree may need debugging"
fi

if grep -q ".git\|node_modules" /tmp/direct_tree_output.txt; then
    echo "⚠️  Smart filtering may need improvement - showing ignored folders"
else
    echo "✅ SMART FILTERING WORKING - Ignoring common folders"
fi

echo ""
echo "🎯 CONTEXT ENHANCEMENT SUMMARY"
echo "=============================="
echo "✅ Added project_tree tool for VSCode-level context"
echo "✅ Enhanced Agent 0 with context gathering workflow"
echo "✅ Tested with real Agentry project structure"
echo ""
echo "💡 Next Steps:"
echo "1. Refine project_tree output formatting"
echo "2. Add project analysis tool for tech stack detection"
echo "3. Test context-aware delegation improvements"
echo "4. Measure coordination success rate improvements"

# Cleanup
rm -f /tmp/context_input.txt /tmp/direct_tree_test.txt /tmp/direct_tree_output.txt
