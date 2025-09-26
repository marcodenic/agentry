#!/bin/bash

# Agentry Build Script with Go 1.25 Optimizations by Default
# This script ensures you always get the optimized build without setting environment variables

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Go 1.25 optimizations enabled by default
export GOEXPERIMENT=jsonv2,greenteagc

echo -e "${BLUE}🚀 Agentry Builder with Go 1.25 Optimizations${NC}"
echo -e "${YELLOW}   Experimental features: JSON v2 + Green Tea GC${NC}"
echo ""

# Check Go version
GO_VERSION=$(go version)
echo -e "${BLUE}Go Version:${NC} $GO_VERSION"

# Check if we have Go 1.25+
if [[ $GO_VERSION == *"go1.25"* ]] || [[ $GO_VERSION == *"go1.26"* ]] || [[ $GO_VERSION == *"go1.27"* ]]; then
    echo -e "${GREEN}✅ Go 1.25+ detected - optimizations available${NC}"
else
    echo -e "${YELLOW}⚠️  Go 1.25+ recommended for best performance${NC}"
fi

echo ""

# Default action
ACTION=${1:-build}

case $ACTION in
    "build" | "")
        echo -e "${BLUE}Building optimized Agentry...${NC}"
        go build -o agentry ./cmd/agentry
        echo -e "${GREEN}✅ Build complete: ./agentry${NC}"
        echo -e "${YELLOW}💡 Run with: ./agentry --version${NC}"
        ;;
    
    "release")
        echo -e "${BLUE}Building release version...${NC}"
        go build -ldflags="-w -s" -trimpath -o agentry ./cmd/agentry
        echo -e "${GREEN}✅ Release build complete: ./agentry${NC}"
        ;;
        
    "install")
        echo -e "${BLUE}Installing optimized Agentry...${NC}"
        go install ./cmd/agentry
        echo -e "${GREEN}✅ Agentry installed to $(go env GOPATH)/bin/agentry${NC}"
        ;;
        
    "test")
        echo -e "${BLUE}Running tests with optimizations...${NC}"
        go test -v ./...
        ;;
        
    "benchmark")
        echo -e "${BLUE}Running performance benchmark...${NC}"
        go run ./cmd/benchmark 5000
        ;;
        
    "clean")
        echo -e "${BLUE}Cleaning build artifacts...${NC}"
        rm -f agentry agentry.exe agentry-*
        echo -e "${GREEN}✅ Clean complete${NC}"
        ;;
        
    "info")
        echo -e "${BLUE}Build Information:${NC}"
        echo "  Go Version: $GO_VERSION"
        echo "  Experiments: $GOEXPERIMENT"
        echo "  GOPATH: $(go env GOPATH)"
        echo "  GOROOT: $(go env GOROOT)"
        echo "  OS/Arch: $(go env GOOS)/$(go env GOARCH)"
        ;;
        
    "help" | "-h" | "--help")
        echo -e "${BLUE}Agentry Build Script${NC}"
        echo ""
        echo "Usage: $0 [command]"
        echo ""
        echo "Commands:"
        echo "  build     - Build optimized binary (default)"
        echo "  release   - Build fully optimized release binary"
        echo "  install   - Install to Go bin directory"
        echo "  test      - Run tests with optimizations"
        echo "  benchmark - Run performance benchmarks"
        echo "  clean     - Clean build artifacts"
        echo "  info      - Show build information"
        echo "  help      - Show this help"
        echo ""
        echo "Examples:"
        echo "  $0              # Build optimized binary"
        echo "  $0 build        # Same as above"
        echo "  $0 release      # Build release version"
        echo "  $0 install      # Install to system"
        ;;
        
    *)
        echo -e "${RED}❌ Unknown command: $ACTION${NC}"
        echo "Use '$0 help' for available commands"
        exit 1
        ;;
esac
