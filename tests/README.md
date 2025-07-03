# Agentry Testing

This directory contains the test suite for Agentry.

## Test Structure

```
tests/
├── README.md              # This file
├── coordination/           # Agent coordination and delegation tests
├── archive/               # Archived/historical tests
├── bash-tool/             # Shell tool specific tests
├── *.go                   # Go unit tests
├── *.json                 # Test suites and configuration
└── testutil.go            # Test utilities
```

## Running Tests

### Quick Start
```bash
# Run all tests (recommended)
./scripts/test.sh

# Run specific test categories
./scripts/test.sh unit          # Go unit tests only
./scripts/test.sh coordination  # Agent coordination tests
./scripts/test.sh build         # Build binary only
```

### Manual Testing
```bash
# Build binary
make build
# or
./scripts/build.sh --tools

# Run Go unit tests
go test ./...

# Run specific coordination test
cd /tmp/agentry-ai-sandbox
bash /path/to/agentry/tests/coordination/test_agent_execution.sh
```

## Test Categories

### Unit Tests (*.go)
Standard Go unit tests for internal packages and components.

### Coordination Tests (coordination/*.sh)
Integration tests that verify multi-agent coordination:
- Agent spawning and delegation
- Tool restrictions and role enforcement  
- Real-world coordination scenarios
- Agent communication and status tracking

### Key Test Files
- `test_agent_execution.sh` - Verifies agents execute delegated tasks
- `test_agent_0_tool_restrictions.sh` - Verifies Agent 0 role restrictions
- `validate_agent_0.sh` - Basic Agent 0 functionality validation
- `test_fixed_delegation.sh` - Delegation workflow testing

## Test Environment

All integration tests run in isolated sandboxes:
- **Workspace**: `/tmp/agentry-ai-sandbox` 
- **Isolation**: No access to project source code
- **Config**: Uses `.agentry.yaml` and `.env.local` from project root
- **Binary**: Uses locally built `agentry` executable

## Cross-Platform Support

Tests handle binary naming automatically:
- **Unix/Linux/macOS**: `agentry`
- **Windows**: `agentry.exe`

The test helpers in `scripts/test-helpers.sh` provide cross-platform utilities for test setup and execution.

## Adding New Tests

1. For Go unit tests: Add `*_test.go` files in appropriate packages
2. For coordination tests: Add shell scripts to `coordination/` directory
3. Use the test helper functions from `scripts/test-helpers.sh`
4. Follow the existing test naming convention

### Test Script Template
```bash
#!/bin/bash
set -e

# Source test helpers
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
source "$(dirname "$(dirname "$SCRIPT_DIR")")/scripts/test-helpers.sh"

echo "=== YOUR TEST NAME ==="

# Setup test workspace
SANDBOX_DIR="/tmp/agentry-your-test"
setup_test_workspace "$SANDBOX_DIR"

# Your test logic here...
run_agentry 60s chat < input.txt > output.txt
```
