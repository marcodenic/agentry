#!/bin/bash
# Cross-platform build script for Agentry

set -e

# Detect OS and set binary name
if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "win32" ]]; then
    BINARY_NAME="agentry.exe"
else
    BINARY_NAME="agentry"
fi

# Build options
BUILD_TAGS=""
OUTPUT_DIR="."
VERBOSE=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --tools)
            BUILD_TAGS="-tags tools"
            shift
            ;;
        --output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        --verbose)
            VERBOSE=true
            shift
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo "Options:"
            echo "  --tools     Build with tools tag (includes plugin, tool, cost commands)"
            echo "  --output    Output directory (default: current directory)"
            echo "  --verbose   Verbose output"
            echo "  --help      Show this help"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Build command
BUILD_CMD="go build $BUILD_TAGS -o $OUTPUT_DIR/$BINARY_NAME ./cmd/agentry"

if [ "$VERBOSE" = true ]; then
    echo "Building Agentry..."
    echo "Command: $BUILD_CMD"
    echo "Binary name: $BINARY_NAME"
    echo "Output directory: $OUTPUT_DIR"
fi

# Execute build
eval $BUILD_CMD

if [ $? -eq 0 ]; then
    echo "✅ Build successful: $OUTPUT_DIR/$BINARY_NAME"
else
    echo "❌ Build failed"
    exit 1
fi
