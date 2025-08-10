#!/bin/bash

# Demonstration of Agentry's Enhanced Error Resilience
echo "🛡️  Agentry Error Resilience Demonstration"
echo "==========================================="
echo ""

echo "This demonstrates how agentry now handles errors gracefully instead of"
echo "terminating immediately when tools fail or agents encounter issues."
echo ""

# Build agentry if not already built
if [ ! -f "agentry" ]; then
    echo "🔨 Building agentry..."
    go build -o agentry ./cmd/agentry
fi

echo "📋 Key Improvements:"
echo "==================="
echo ""
echo "✅ Tool errors are now returned as feedback to agents"
echo "✅ Agents can see error messages and try alternative approaches"
echo "✅ Agent delegation failures provide helpful error context"
echo "✅ Configurable retry limits prevent infinite error loops"
echo "✅ Detailed error context helps agents understand what went wrong"
echo ""

echo "🧪 Running Tests:"
echo "================"
echo ""

echo "1. Testing error resilience with mock agents..."
go test -v ./tests/error_resilience_test.go -run TestErrorHandlingWithResilientAgent

echo ""
echo "2. Testing delegation error handling..."
go test -v ./tests/agent_tool_test.go

echo ""
echo "3. Testing Agent 0 debugging capabilities..."
go test -v ./tests/agent_0_debug_test.go

echo ""
echo "🎯 Configuration Example:"
echo "========================"
echo ""
echo "Error handling is now configurable in agent creation:"
echo ""
cat << 'EOF'
agent.ErrorHandling.TreatErrorsAsResults = true  // Errors become feedback
agent.ErrorHandling.MaxErrorRetries = 3          // Allow 3 consecutive errors
agent.ErrorHandling.IncludeErrorContext = true   // Provide detailed context
EOF

echo ""
echo "📊 Benefits:"
echo "============"
echo ""
echo "• Agents can recover from tool failures and try alternatives"
echo "• Multi-agent workflows continue even if individual agents hit errors"
echo "• Better debugging with detailed error context"
echo "• Configurable resilience levels for different use cases"
echo "• Graceful degradation instead of hard failures"
echo ""

echo "✨ Agentry is now significantly more robust and resilient!"
