#!/bin/bash

# Focused Natural Language Agent Spawning Test
# Verify Agent 0 uses correct agent names and actually spawns agents

echo "🎯 Focused Natural Language Agent Spawning Test"
echo "=============================================="
echo "Testing if Agent 0 spawns agents with correct names from natural language"
echo ""

# Create a test workspace
mkdir -p /tmp/agentry-spawn-test
cd /tmp/agentry-spawn-test

# Copy .env.local
if [ -f "/home/marco/Documents/GitHub/agentry/.env.local" ]; then
    cp "/home/marco/Documents/GitHub/agentry/.env.local" .
    echo "📋 Environment ready"
fi

echo "📁 Test workspace: $(pwd)"
echo ""

# Test function for agent spawning verification
test_agent_spawning() {
    local test_name="$1"
    local request="$2"
    local expected_agent="$3"
    
    echo "🧪 Test: $test_name"
    echo "📝 Request: $request"
    echo "🎯 Expected agent: $expected_agent"
    echo "----------------------------------------"
    
    # Create input file
    echo "$request" > /tmp/spawn_input.txt
    echo "/list" >> /tmp/spawn_input.txt  
    echo "/quit" >> /tmp/spawn_input.txt
    
    # Run test and capture output
    local output_file="/tmp/spawn_output.txt"
    timeout 60 /home/marco/Documents/GitHub/agentry/agentry.exe chat < /tmp/spawn_input.txt > "$output_file" 2>&1
    
    echo "📋 Agent 0 Response:"
    cat "$output_file"
    echo ""
    
    # Check for agent spawning
    if grep -q "Agent.*spawned successfully" "$output_file"; then
        echo "✅ AGENT SPAWNING DETECTED"
        # Extract agent name from spawn message
        spawned_agent=$(grep "Agent.*spawned successfully" "$output_file" | sed -n "s/.*Agent '\([^']*\)'.*/\1/p")
        echo "   Spawned agent: $spawned_agent"
        if [ "$spawned_agent" = "$expected_agent" ]; then
            echo "   ✅ CORRECT AGENT NAME"
        else
            echo "   ⚠️  INCORRECT AGENT NAME (expected: $expected_agent, got: $spawned_agent)"
        fi
    else
        echo "❌ NO AGENT SPAWNING DETECTED"
    fi
    
    # Check if agent appears in list
    if grep -q "$expected_agent" "$output_file"; then
        echo "✅ AGENT APPEARS IN LIST"
    else
        echo "❌ AGENT NOT IN LIST"
    fi
    
    echo ""
    echo "=========================================="
    echo ""
    sleep 3
}

# Run focused tests
test_agent_spawning \
    "Coding Request" \
    "I need help writing a Python function that sorts a list" \
    "coder"

test_agent_spawning \
    "Documentation Request" \
    "I need someone to write documentation for this API" \
    "writer"

test_agent_spawning \
    "Analysis Request" \
    "I need help analyzing this data and finding patterns" \
    "analyst"

test_agent_spawning \
    "Testing Request" \
    "I need someone to write unit tests for my functions" \
    "tester"

# Final verification test - multiple agents
echo "🔄 Final Test: Multiple Agent Coordination"
echo "========================================="

echo "I need both a coder to write JavaScript code and a writer to document it" > /tmp/multi_input.txt
echo "/list" >> /tmp/multi_input.txt
echo "/quit" >> /tmp/multi_input.txt

timeout 90 /home/marco/Documents/GitHub/agentry/agentry.exe chat < /tmp/multi_input.txt > /tmp/multi_output.txt 2>&1

echo "📋 Multiple Agent Request Response:"
cat /tmp/multi_output.txt
echo ""

# Count how many agents were spawned
coder_spawned=$(grep -c "Agent 'coder' spawned successfully" /tmp/multi_output.txt)
writer_spawned=$(grep -c "Agent 'writer' spawned successfully" /tmp/multi_output.txt)

echo "📊 Multiple Agent Results:"
echo "   Coder agents spawned: $coder_spawned"
echo "   Writer agents spawned: $writer_spawned"

if [ $coder_spawned -gt 0 ] && [ $writer_spawned -gt 0 ]; then
    echo "   ✅ MULTIPLE AGENT COORDINATION SUCCESS"
else
    echo "   ⚠️  Partial or no multiple agent coordination"
fi

echo ""
echo "✅ Focused agent spawning tests completed"

# Clean up
rm -f /tmp/spawn_input.txt /tmp/spawn_output.txt /tmp/multi_input.txt /tmp/multi_output.txt
