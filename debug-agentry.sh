#!/bin/bash

# Agentry Debug Wrapper
# This script runs agentry with comprehensive debug logging enabled
# and provides easy access to the debug logs for troubleshooting.

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔍 Agentry Debug Wrapper${NC}"
echo -e "${YELLOW}   Comprehensive logging enabled${NC}"
echo ""

# Create debug directory if it doesn't exist
mkdir -p debug

# Set comprehensive debug logging
export AGENTRY_DEBUG_LEVEL=trace
export AGENTRY_TUI_MODE=1  # Ensure file logging works in TUI mode

# Show current debug configuration
echo -e "${GREEN}Debug Configuration:${NC}"
echo "   AGENTRY_DEBUG_LEVEL: $AGENTRY_DEBUG_LEVEL"
echo "   Log Directory: $(pwd)/debug/"
echo "   Rolling Log Size: 1MB per file"
echo ""

# Function to show log monitoring tips
show_log_tips() {
    echo ""
    echo -e "${BLUE}💡 Debug Log Monitoring Tips:${NC}"
    echo ""
    echo -e "${YELLOW}Real-time monitoring:${NC}"
    echo "   tail -f debug/agentry-debug-*.log"
    echo ""
    echo -e "${YELLOW}Search for errors:${NC}"
    echo "   grep -i 'error\\|fail\\|exception' debug/agentry-debug-*.log"
    echo ""
    echo -e "${YELLOW}Find tool executions:${NC}"
    echo "   grep 'TOOL.*call' debug/agentry-debug-*.log"
    echo ""
    echo -e "${YELLOW}Track agent actions:${NC}"
    echo "   grep 'AGENT' debug/agentry-debug-*.log"
    echo ""
    echo -e "${YELLOW}Monitor model interactions:${NC}"
    echo "   grep 'MODEL' debug/agentry-debug-*.log"
    echo ""
    echo -e "${YELLOW}View latest debug entries:${NC}"
    echo "   tail -n 50 debug/agentry-debug-*.log"
    echo ""
}

# If no arguments provided, show help and tips
if [ $# -eq 0 ]; then
    echo -e "${GREEN}Usage:${NC}"
    echo "   $0 [agentry_arguments...]"
    echo ""
    echo -e "${GREEN}Examples:${NC}"
    echo "   $0                                    # Start TUI with debug logging"
    echo "   $0 \"create a hello world program\"     # Direct command with logging"
    echo "   $0 tui                               # Explicit TUI mode with logging"
    echo ""
    show_log_tips
    exit 0
fi

# Show that we're starting with debug logging
echo -e "${GREEN}🚀 Starting agentry with debug logging...${NC}"
echo "   Arguments: $*"

# Function to show log summary after execution
show_log_summary() {
    echo ""
    echo -e "${BLUE}📋 Debug Log Summary:${NC}"
    if ls debug/agentry-debug-*.log >/dev/null 2>&1; then
        for logfile in debug/agentry-debug-*.log; do
            lines=$(wc -l < "$logfile")
            size=$(du -h "$logfile" | cut -f1)
            echo -e "   📄 $logfile: ${GREEN}$lines lines${NC}, ${YELLOW}$size${NC}"
        done
        
        latest_log=$(ls -t debug/agentry-debug-*.log | head -n1)
        echo ""
        echo -e "${GREEN}Last 10 debug entries:${NC}"
        echo "----------------------"
        tail -n 10 "$latest_log" | while read -r line; do
            echo "   $line"
        done
    else
        echo -e "   ${RED}❌ No debug logs found${NC}"
    fi
}

# Set up trap to show log summary on exit
trap show_log_summary EXIT

# Run agentry with all provided arguments
if [ "$1" = "tui" ] || [ $# -eq 0 ]; then
    echo -e "${YELLOW}   Running in TUI mode - debug logs will be written to files${NC}"
    echo -e "${YELLOW}   Use Ctrl+C to exit and view log summary${NC}"
    echo ""
fi

./agentry "$@"