#!/bin/bash

# Test script to verify Ctrl+C fix for Agentry TUI
# This script builds the application and tests that Ctrl+C properly terminates it

echo "🔨 Building Agentry..."
go build -o agentry.exe ./cmd/agentry

if [ $? -ne 0 ]; then
    echo "❌ Build failed"
    exit 1
fi

echo "✅ Build successful"

echo "🚀 Starting Agentry TUI (will auto-terminate after 5 seconds)..."
echo "   In a real test, you would press Ctrl+C to test termination"

# Start the TUI in background
timeout 5s ./agentry.exe tui &
PID=$!

# Wait a moment for it to start
sleep 2

echo "🔄 Sending SIGINT (Ctrl+C equivalent) to process $PID..."
kill -SIGINT $PID

# Wait for it to terminate
wait $PID 2>/dev/null
EXIT_CODE=$?

if [ $EXIT_CODE -eq 0 ] || [ $EXIT_CODE -eq 130 ]; then
    echo "✅ Process terminated gracefully with exit code $EXIT_CODE"
    echo "🎉 Ctrl+C fix appears to be working!"
else
    echo "❌ Process did not terminate gracefully (exit code: $EXIT_CODE)"
    echo "🔍 Check if there are hanging processes:"
    ps aux | grep agentry | grep -v grep
fi

echo "🧹 Cleaning up..."
rm -f agentry.exe

echo "📋 Test completed"
