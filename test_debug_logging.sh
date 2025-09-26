#!/bin/bash

# Test script to demonstrate the new comprehensive debug logging
# This will help you track exactly what's happening during agent interactions

echo "🔍 Testing Agentry's New Rolling Debug Log System"
echo "=============================================="
echo ""

# Clean up any old debug logs
echo "Cleaning up old debug logs..."
rm -rf debug/agentry-debug-*.log 2>/dev/null

# Set debug level for maximum verbosity
export AGENTRY_DEBUG_LEVEL=trace

echo "🚀 Starting a simple delegation test with debug logging enabled..."
echo "   Debug logs will be written to: debug/agentry-debug-*.log"
echo ""

# Run a simple test interaction
./agentry "delegate to coder: create a simple hello.txt file with 'Hello World' in it"

echo ""
echo "📋 Debug Logging Results:"
echo "========================"

# Show debug log files created
if ls debug/agentry-debug-*.log >/dev/null 2>&1; then
    echo "✅ Debug log files created:"
    for logfile in debug/agentry-debug-*.log; do
        echo "   📄 $logfile ($(wc -l < "$logfile") lines)"
    done
    echo ""
    
    echo "📖 Last 20 lines from most recent debug log:"
    echo "--------------------------------------------"
    latest_log=$(ls -t debug/agentry-debug-*.log | head -n1)
    tail -n 20 "$latest_log"
    
    echo ""
    echo "🔎 Search for tool executions:"
    echo "-----------------------------"
    grep -n "TOOL.*call" "$latest_log" | head -n 10
    
    echo ""
    echo "🤖 Search for agent actions:"
    echo "----------------------------"
    grep -n "AGENT" "$latest_log" | head -n 10
    
    echo ""
    echo "🌐 Search for model interactions:"
    echo "--------------------------------"
    grep -n "MODEL" "$latest_log" | head -n 10
    
else
    echo "❌ No debug log files found!"
    echo "   This might indicate an issue with the logging system."
fi

echo ""
echo "💡 To monitor debug logs in real-time during TUI usage:"
echo "   tail -f debug/agentry-debug-*.log"
echo ""
echo "💡 To search for specific issues:"
echo "   grep 'ERROR\\|FAIL\\|error' debug/agentry-debug-*.log"
echo ""
echo "💡 Debug log location: $(pwd)/debug/"