#!/bin/bash

# Quick Context Tool Test
echo "🔧 Quick Context Tool Test"
echo "=========================="

cd /home/marco/Documents/GitHub/agentry

echo "Testing if Agent 0 uses project_tree tool correctly..."

echo "Show me the project structure using your project_tree tool" > /tmp/quick_test.txt
echo "/quit" >> /tmp/quick_test.txt

timeout 45 ./agentry.exe chat < /tmp/quick_test.txt > /tmp/quick_output.txt 2>&1

echo "📋 Agent 0 Response:"
echo "-------------------"
cat /tmp/quick_output.txt
echo "-------------------"

echo ""
echo "🔍 Analysis:"
if grep -q "📂.*Structure\|📁\|📄" /tmp/quick_output.txt; then
    echo "✅ PROJECT_TREE TOOL USED SUCCESSFULLY"
else
    echo "❌ Project tree tool not used or not working"
fi

if grep -q "cannot create agent with tool name" /tmp/quick_output.txt; then
    echo "❌ Still trying to use project_tree as agent name"
else
    echo "✅ No tool/agent name confusion"
fi

rm -f /tmp/quick_test.txt /tmp/quick_output.txt
