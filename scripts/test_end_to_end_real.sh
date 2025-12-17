#!/bin/bash
set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

PROJECT_DIR="TEST_PROJECT"
AGENTRY_BIN="./agentry"

echo -e "${BLUE}=== Agentry Real-World End-to-End Test ===${NC}"

# 1. Setup
echo -e "${BLUE}[1/4] Setting up test environment...${NC}"
if [ -d "$PROJECT_DIR" ]; then
    echo "Cleaning up previous $PROJECT_DIR..."
    rm -rf "$PROJECT_DIR"
fi
mkdir -p "$PROJECT_DIR"

if [ ! -f "$AGENTRY_BIN" ]; then
    echo -e "${RED}Error: agentry binary not found. Run ./build.sh first.${NC}"
    exit 1
fi

# 2. Execution
echo -e "${BLUE}[2/4] Running Agentry to create a Go web server project...${NC}"
echo "Prompt: Create a simple Go HTTP server in $PROJECT_DIR/main.go that listens on port 8090 and returns 'Hello from Agentry'. Also create a README.md in the same folder explaining how to run it. Use the Coder agent for code and Writer agent for docs."

# We use the --debug flag to see more details if needed, but standard output is fine.
# We pass the prompt as a single argument.
$AGENTRY_BIN "Create a simple Go HTTP server in $PROJECT_DIR/main.go that listens on port 8090 and returns 'Hello from Agentry'. Also create a README.md in the same folder explaining how to run it. Use the Coder agent for code and Writer agent for docs."

# 3. Verification
echo -e "${BLUE}[3/4] Verifying generated artifacts...${NC}"

FAILED=0

# Check main.go
if [ -f "$PROJECT_DIR/main.go" ]; then
    echo -e "${GREEN}✅ $PROJECT_DIR/main.go exists${NC}"
    if grep -q "8090" "$PROJECT_DIR/main.go"; then
        echo -e "${GREEN}✅ Port 8090 found in main.go${NC}"
    else
        echo -e "${RED}❌ Port 8090 NOT found in main.go${NC}"
        FAILED=1
    fi
else
    echo -e "${RED}❌ $PROJECT_DIR/main.go missing${NC}"
    FAILED=1
fi

# Check README.md
if [ -f "$PROJECT_DIR/README.md" ]; then
    echo -e "${GREEN}✅ $PROJECT_DIR/README.md exists${NC}"
else
    echo -e "${RED}❌ $PROJECT_DIR/README.md missing${NC}"
    FAILED=1
fi

# 4. Functional Test (Optional: Try to build the generated code)
echo -e "${BLUE}[4/4] Attempting to build generated code...${NC}"
if [ -f "$PROJECT_DIR/main.go" ]; then
    if go build -o "$PROJECT_DIR/server" "$PROJECT_DIR/main.go"; then
        echo -e "${GREEN}✅ Generated code compiles successfully${NC}"
        rm "$PROJECT_DIR/server"
    else
        echo -e "${RED}❌ Generated code failed to compile${NC}"
        FAILED=1
    fi
fi

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}=== TEST PASSED: Agentry successfully created a working project ===${NC}"
    exit 0
else
    echo -e "${RED}=== TEST FAILED: Artifacts missing or incorrect ===${NC}"
    exit 1
fi
