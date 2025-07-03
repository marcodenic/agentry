#!/bin/bash
# Main test runner for Agentry

set -e

# Source test helpers
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
source "$SCRIPT_DIR/test-helpers.sh"

# Test workspace
TEST_WORKSPACE="/tmp/agentry-ai-sandbox"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print colored output
print_status() {
    local color="$1"
    local message="$2"
    echo -e "${color}${message}${NC}"
}

# Test categories
run_unit_tests() {
    print_status "$BLUE" "🧪 Running Unit Tests..."
    cd "$PROJECT_DIR"
    go test ./... -v
}

run_integration_tests() {
    print_status "$BLUE" "🔧 Running Integration Tests..."
    cd "$PROJECT_DIR"
    go test -tags integration ./... -v
}

run_coordination_tests() {
    print_status "$BLUE" "🤖 Running Agent Coordination Tests..."
    
    # Setup workspace
    setup_test_workspace "$TEST_WORKSPACE"
    
    local test_dir="$PROJECT_DIR/tests/coordination"
    local passed=0
    local failed=0
    
    # Run key coordination tests
    local key_tests=(
        "test_agent_0_tool_restrictions.sh"
        "test_agent_execution.sh"
        "test_fixed_delegation.sh"
        "validate_agent_0.sh"
    )
    
    for test in "${key_tests[@]}"; do
        if [ -f "$test_dir/$test" ]; then
            print_status "$YELLOW" "Running $test..."
            if cd "$TEST_WORKSPACE" && bash "$test_dir/$test"; then
                print_status "$GREEN" "✅ $test passed"
                ((passed++))
            else
                print_status "$RED" "❌ $test failed"
                ((failed++))
            fi
        else
            print_status "$YELLOW" "⚠️  $test not found"
        fi
    done
    
    print_status "$BLUE" "Coordination Tests Summary: $passed passed, $failed failed"
}

build_and_test() {
    print_status "$BLUE" "🔨 Building Agentry..."
    cd "$PROJECT_DIR"
    
    # Build with tools for full functionality
    if ./scripts/build.sh --tools --verbose; then
        print_status "$GREEN" "✅ Build successful"
    else
        print_status "$RED" "❌ Build failed"
        exit 1
    fi
}

# Parse command line arguments
case "${1:-all}" in
    "unit")
        run_unit_tests
        ;;
    "integration")
        run_integration_tests
        ;;
    "coordination")
        build_and_test
        run_coordination_tests
        ;;
    "build")
        build_and_test
        ;;
    "all")
        build_and_test
        run_unit_tests
        run_coordination_tests
        ;;
    "help")
        echo "Usage: $0 [unit|integration|coordination|build|all|help]"
        echo ""
        echo "Commands:"
        echo "  unit         Run Go unit tests"
        echo "  integration  Run integration tests"
        echo "  coordination Run agent coordination tests"
        echo "  build        Build binary only"
        echo "  all          Run build + unit + coordination tests (default)"
        echo "  help         Show this help"
        ;;
    *)
        print_status "$RED" "Unknown command: $1"
        echo "Use '$0 help' for usage information"
        exit 1
        ;;
esac
