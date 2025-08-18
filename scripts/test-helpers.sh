#!/bin/bash
# Test helper script that handles binary naming and setup consistently

# Get the directory where this script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# Detect binary name
if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "win32" ]]; then
    BINARY_NAME="agentry.exe"
else
    BINARY_NAME="agentry"
fi

# Function to copy agentry binary to test workspace
setup_agentry_binary() {
    local target_dir="$1"
    
    # Look for binary in project root
    if [ -f "$PROJECT_DIR/$BINARY_NAME" ]; then
        cp "$PROJECT_DIR/$BINARY_NAME" "$target_dir/agentry"
        chmod +x "$target_dir/agentry"
        echo "✅ Copied $BINARY_NAME as agentry"
        return 0
    fi
    
    # Look for alternative name
    local alt_name
    if [ "$BINARY_NAME" = "agentry.exe" ]; then
        alt_name="agentry"
    else
        alt_name="agentry.exe"
    fi
    
    if [ -f "$PROJECT_DIR/$alt_name" ]; then
        cp "$PROJECT_DIR/$alt_name" "$target_dir/agentry"
        chmod +x "$target_dir/agentry"
        echo "✅ Copied $alt_name as agentry"
        return 0
    fi
    
    echo "❌ No agentry binary found in $PROJECT_DIR"
    echo "   Run: make build or scripts/build.sh"
    return 1
}

# Function to setup test workspace with all required files
setup_test_workspace() {
    local workspace="$1"
    
    echo "🏗️  Setting up test workspace: $workspace"
    mkdir -p "$workspace"
    cd "$workspace"
    
    # Copy agentry binary
    if ! setup_agentry_binary "$workspace"; then
        exit 1
    fi
    
    # Copy configuration files
    if [ -f "$PROJECT_DIR/.agentry.yaml" ]; then
        cp "$PROJECT_DIR/.agentry.yaml" .
        echo "✅ Copied .agentry.yaml"
    fi
    
    if [ -f "$PROJECT_DIR/.env.local" ]; then
        cp "$PROJECT_DIR/.env.local" .
        echo "✅ Copied .env.local"
    else
        echo "⚠️  No .env.local found - API calls may fail"
    fi
    
    # Copy templates directory
    if [ -d "$PROJECT_DIR/templates" ]; then
        cp -r "$PROJECT_DIR/templates" .
        echo "✅ Copied templates"
    fi
    
    echo "✅ Test workspace ready: $(pwd)"
}

# Function to run agentry with timeout and error handling
run_agentry() {
    local timeout_duration="$1"
    local mode="$2"
    shift 2
    local args="$@"
    
    if ! command -v timeout >/dev/null 2>&1; then
        echo "⚠️  timeout command not available, running without timeout"
        ./agentry "$mode" $args
    else
        timeout "$timeout_duration" ./agentry "$mode" $args
    fi
}

# Export functions for use by test scripts
export -f setup_agentry_binary
export -f setup_test_workspace  
export -f run_agentry
export PROJECT_DIR
export BINARY_NAME
