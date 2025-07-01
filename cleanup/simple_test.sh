#!/bin/bash

# Simple test to verify our key findings
echo "🧪 Simple CLI Test Verification"
echo "================================"

cd /tmp/agentry-test-workspace
rm -f agent_test_file.txt
cp /home/marco/Documents/GitHub/agentry/.env.local .

echo "Test 1: File creation (should work)"
echo "Create a file called agent_test_file.txt with the content 'Hello from Agent 0'" > /tmp/test_input.txt
echo "/quit" >> /tmp/test_input.txt
/home/marco/Documents/GitHub/agentry/agentry.exe chat < /tmp/test_input.txt > /tmp/result1.txt 2>&1

echo "File created? " && if [ -f "agent_test_file.txt" ]; then echo "✅ YES"; else echo "❌ NO"; fi
echo "AI Response: " && grep "🤖 system:" /tmp/result1.txt -A 3 | tail -3

echo ""
echo "Test 2: Direct CLI spawn command (should work)"
echo "/spawn coder 'Help with coding'" > /tmp/test_input.txt
echo "/quit" >> /tmp/test_input.txt
/home/marco/Documents/GitHub/agentry/agentry.exe chat < /tmp/test_input.txt > /tmp/result2.txt 2>&1

echo "Spawn worked? " && if grep -q "✅ Agent.*spawned successfully" /tmp/result2.txt; then echo "✅ YES"; else echo "❌ NO"; fi
echo "CLI Response: " && grep "✅ Agent" /tmp/result2.txt || echo "No success message found"

echo ""
echo "Test 3: Natural language vs CLI understanding"
echo "Please use the /spawn command to create a coder agent" > /tmp/test_input.txt
echo "/quit" >> /tmp/test_input.txt
/home/marco/Documents/GitHub/agentry/agentry.exe chat < /tmp/test_input.txt > /tmp/result3.txt 2>&1

echo "AI understood CLI command? " && if grep -q "✅ Agent.*spawned successfully" /tmp/result3.txt; then echo "✅ YES"; else echo "❌ NO"; fi
echo "AI Response: " && grep "🤖 system:" /tmp/result3.txt -A 3 | tail -3

echo ""
echo "✅ Simple verification complete"
