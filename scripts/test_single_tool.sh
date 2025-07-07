#!/bin/bash

# Simple tool tester for Agentry
# Usage: ./test_single_tool.sh <tool_name> [prompt]

set -e

AGENTRY_CONFIG="config/smart-config.yaml"
TEST_DIR="tests/tool_tests"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Create test directory
mkdir -p "$TEST_DIR"

# Function to test a single tool
test_tool() {
    local tool_name="$1"
    local prompt="$2"
    
    echo -e "${YELLOW}Testing tool: $tool_name${NC}"
    echo "Prompt: $prompt"
    echo ""
    
    # Build agentry if needed
    if [ ! -f "./agentry" ]; then
        echo -e "${YELLOW}Building agentry...${NC}"
        make build-tools
    fi
    
    # Run the test
    echo "$prompt" | ./agentry chat -c "$AGENTRY_CONFIG" -f -
    
    echo ""
    echo -e "${GREEN}Test completed for $tool_name${NC}"
}

# Predefined test prompts for each tool
case "$1" in
    "echo")
        test_tool "echo" "Use the echo tool to repeat the text 'Hello, World! This is a test of the echo tool.'"
        ;;
    "ping")
        test_tool "ping" "Use the ping tool to check connectivity to google.com"
        ;;
    "sysinfo")
        test_tool "sysinfo" "Use the sysinfo tool to get comprehensive system information"
        ;;
    "ls")
        test_tool "ls" "Use the ls tool to list contents of the current directory"
        ;;
    "find")
        test_tool "find" "Use the find tool to locate all .go files in the current directory"
        ;;
    "glob")
        test_tool "glob" "Use the glob tool to find all *.yaml files"
        ;;
    "view")
        # First create a test file
        cat > "$TEST_DIR/sample.txt" << 'EOF'
Line 1: This is a test file
Line 2: With multiple lines
Line 3: For testing the view tool
Line 4: Each line has different content
Line 5: This is the last line
EOF
        test_tool "view" "Use the view tool to display the file tests/tool_tests/sample.txt with line numbers"
        ;;
    "create")
        test_tool "create" "Use the create tool to make a new file called tests/tool_tests/created_file.txt with content: 'This is a test file\\nCreated by agentry\\nWith multiple lines'"
        ;;
    "fileinfo")
        # Create test file first
        echo "Test file for fileinfo" > "$TEST_DIR/fileinfo_test.txt"
        test_tool "fileinfo" "Use the fileinfo tool to get detailed information about tests/tool_tests/fileinfo_test.txt"
        ;;
    "grep")
        # Create test file first
        cat > "$TEST_DIR/grep_test.txt" << 'EOF'
This is a test file
It contains multiple lines
Some lines have the word test
Others have different content
The word test appears multiple times
EOF
        test_tool "grep" "Use the grep tool to search for the word 'test' in tests/tool_tests/grep_test.txt"
        ;;
    "read_lines")
        # Create test file first
        cat > "$TEST_DIR/read_lines_test.txt" << 'EOF'
Line 1
Line 2
Line 3
Line 4
Line 5
Line 6
Line 7
Line 8
Line 9
Line 10
EOF
        test_tool "read_lines" "Use the read_lines tool to read lines 3-7 from tests/tool_tests/read_lines_test.txt"
        ;;
    "fetch")
        test_tool "fetch" "Use the fetch tool to download content from https://httpbin.org/get"
        ;;
    "api")
        test_tool "api" "Use the api tool to make a GET request to https://httpbin.org/get"
        ;;
    "bash")
        test_tool "bash" "Use the bash tool to run 'echo \"Hello from bash\"'"
        ;;
    "sh")
        test_tool "sh" "Use the sh tool to run 'pwd' and show current directory"
        ;;
    "all")
        echo -e "${YELLOW}Running all basic tool tests...${NC}"
        echo ""
        
        # Run basic tests for all tools
        $0 echo
        echo ""; echo "---"; echo ""
        $0 ping
        echo ""; echo "---"; echo ""
        $0 sysinfo
        echo ""; echo "---"; echo ""
        $0 ls
        echo ""; echo "---"; echo ""
        $0 find
        echo ""; echo "---"; echo ""
        $0 glob
        echo ""; echo "---"; echo ""
        $0 view
        echo ""; echo "---"; echo ""
        $0 create
        echo ""; echo "---"; echo ""
        $0 fileinfo
        echo ""; echo "---"; echo ""
        $0 grep
        echo ""; echo "---"; echo ""
        $0 read_lines
        echo ""; echo "---"; echo ""
        $0 fetch
        echo ""; echo "---"; echo ""
        $0 api
        echo ""; echo "---"; echo ""
        $0 bash
        echo ""; echo "---"; echo ""
        $0 sh
        
        echo ""
        echo -e "${GREEN}All basic tool tests completed!${NC}"
        ;;
    *)
        if [ -n "$2" ]; then
            test_tool "$1" "$2"
        else
            echo "Usage: $0 <tool_name> [custom_prompt]"
            echo ""
            echo "Available predefined tests:"
            echo "  echo      - Test echo tool"
            echo "  ping      - Test ping tool"
            echo "  sysinfo   - Test system info tool"
            echo "  ls        - Test directory listing"
            echo "  find      - Test file finding"
            echo "  glob      - Test glob patterns"
            echo "  view      - Test file viewing"
            echo "  create    - Test file creation"
            echo "  fileinfo  - Test file info"
            echo "  grep      - Test content search"
            echo "  read_lines - Test line reading"
            echo "  fetch     - Test HTTP fetching"
            echo "  api       - Test API calls"
            echo "  bash      - Test bash commands"
            echo "  sh        - Test shell commands"
            echo "  all       - Run all basic tests"
            echo ""
            echo "Or provide a custom prompt:"
            echo "  $0 custom_tool \"Your custom prompt here\""
            exit 1
        fi
        ;;
esac
