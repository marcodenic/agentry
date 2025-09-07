#!/bin/bash
# Cross-platform build script for Agentry with Go 1.25 optimizations by default

set -e

# Enable Go 1.25 optimizations by default
export GOEXPERIMENT=jsonv2,greenteagc

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

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
STANDARD_BUILD=false

echo -e "${BLUE}🚀 Agentry Build Script with Go 1.25 Optimizations${NC}"
echo -e "${YELLOW}   Experiments: $GOEXPERIMENT${NC}"

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
        --standard)
            STANDARD_BUILD=true
            unset GOEXPERIMENT
            echo -e "${YELLOW}⚠️  Building without experimental features${NC}"
            shift
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo "Options:"
            echo "  --tools     Build with tools tag (includes plugin, tool, cost commands)"
            echo "  --output    Output directory (default: current directory)"
            echo "  --verbose   Verbose output"
            echo "  --standard  Build without Go 1.25 experimental features"
            echo "  --help      Show this help"
            echo ""
            echo "By default, builds with Go 1.25 optimizations:"
            echo "  - JSON v2 for faster JSON operations"
            echo "  - Green Tea GC for reduced garbage collection overhead"
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
    echo "Go experiments: ${GOEXPERIMENT:-none}"
fi

# Execute build
eval $BUILD_CMD

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Build successful: $OUTPUT_DIR/$BINARY_NAME${NC}"
    if [ "$STANDARD_BUILD" = false ]; then
        echo -e "${GREEN}💡 Built with Go 1.25 optimizations for better performance${NC}"
    fi
else
    echo -e "${RED}❌ Build failed${NC}"
    exit 1
fi
