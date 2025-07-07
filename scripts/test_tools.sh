#!/bin/bash

# Tool Testing Script for Agentry
# This script runs comprehensive tests for all built-in tools

set -e

# Configuration
AGENTRY_CONFIG="config/smart-config.yaml"
TEST_DIR="tests/tool_tests"
RESULTS_DIR="tests/results"
LOG_FILE="tests/tool_test_results.log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Create test directories
mkdir -p "$TEST_DIR" "$RESULTS_DIR"

# Initialize log
echo "=== Agentry Tool Testing Suite ===" > "$LOG_FILE"
echo "Started at: $(date)" >> "$LOG_FILE"
echo "" >> "$LOG_FILE"

# Function to run a test
run_test() {
    local test_name="$1"
    local prompt="$2"
    local expected_success="$3"
    
    echo -e "${YELLOW}Running test: $test_name${NC}"
    echo "Test: $test_name" >> "$LOG_FILE"
    echo "Prompt: $prompt" >> "$LOG_FILE"
    
    # Create a temporary test file for the prompt
    local temp_prompt_file="$TEST_DIR/temp_prompt.txt"
    echo "$prompt" > "$temp_prompt_file"
    
    # Run agentry with the prompt
    if ./agentry chat -c "$AGENTRY_CONFIG" -f "$temp_prompt_file" > "$RESULTS_DIR/${test_name}_output.txt" 2>&1; then
        echo -e "${GREEN}✓ PASS: $test_name${NC}"
        echo "Result: PASS" >> "$LOG_FILE"
    else
        echo -e "${RED}✗ FAIL: $test_name${NC}"
        echo "Result: FAIL" >> "$LOG_FILE"
        echo "Error output:" >> "$LOG_FILE"
        cat "$RESULTS_DIR/${test_name}_output.txt" >> "$LOG_FILE"
    fi
    
    echo "" >> "$LOG_FILE"
    rm -f "$temp_prompt_file"
}

# Function to setup test environment
setup_test_env() {
    echo -e "${YELLOW}Setting up test environment...${NC}"
    
    # Create test files for file operations
    mkdir -p "$TEST_DIR/sample_files"
    
    # Create sample text file
    cat > "$TEST_DIR/sample_files/test.txt" << 'EOF'
Line 1: This is a test file
Line 2: With multiple lines
Line 3: For testing file operations
Line 4: Each line has different content
Line 5: Including special characters: @#$%^&*()
Line 6: And unicode: 你好世界 🌍
Line 7: Numbers: 123456789
Line 8: Mixed content: abc123XYZ
Line 9: Almost at the end
Line 10: This is the last line
EOF

    # Create sample config file
    cat > "$TEST_DIR/sample_files/config.yaml" << 'EOF'
name: test_config
version: 1.0
settings:
  debug: true
  timeout: 30
  retries: 3
features:
  - feature1
  - feature2
  - feature3
EOF

    # Create sample JSON file
    cat > "$TEST_DIR/sample_files/data.json" << 'EOF'
{
  "users": [
    {"id": 1, "name": "Alice", "email": "alice@example.com"},
    {"id": 2, "name": "Bob", "email": "bob@example.com"}
  ],
  "settings": {
    "theme": "dark",
    "notifications": true
  }
}
EOF

    echo -e "${GREEN}Test environment setup complete${NC}"
}

# Function to cleanup test environment
cleanup_test_env() {
    echo -e "${YELLOW}Cleaning up test environment...${NC}"
    # Remove temporary test files created during testing
    find "$TEST_DIR" -name "temp_*" -delete 2>/dev/null || true
    find "$TEST_DIR" -name "test_created_*" -delete 2>/dev/null || true
    echo -e "${GREEN}Cleanup complete${NC}"
}

# Build agentry if not exists
if [ ! -f "./agentry" ]; then
    echo -e "${YELLOW}Building agentry...${NC}"
    make build-tools
fi

# Setup test environment
setup_test_env

echo -e "${GREEN}Starting comprehensive tool testing...${NC}"
echo ""

# Core built-in tools tests
echo -e "${YELLOW}=== Testing Core Built-in Tools ===${NC}"

run_test "echo_basic" "Use the echo tool to repeat the text 'Hello, World!'"
run_test "echo_multiline" "Echo a multi-line string: 'Line 1\nLine 2\nSpecial chars: @#\$%'"
run_test "echo_empty" "Echo an empty string using the echo tool"

run_test "ping_google" "Use the ping tool to check connectivity to google.com"
run_test "ping_localhost" "Ping localhost to verify local connectivity"

run_test "fetch_httpbin" "Use the fetch tool to download content from https://httpbin.org/get"
run_test "fetch_json" "Fetch a JSON response from https://httpbin.org/json"

run_test "sysinfo_basic" "Use the sysinfo tool to get comprehensive system information"

# File operation tools tests
echo -e "${YELLOW}=== Testing File Operation Tools ===${NC}"

run_test "view_test_file" "Use the view tool to display the file tests/tool_tests/sample_files/test.txt with line numbers"
run_test "read_lines_range" "Use the read_lines tool to read lines 3-7 from tests/tool_tests/sample_files/test.txt"
run_test "fileinfo_test" "Use the fileinfo tool to get detailed information about tests/tool_tests/sample_files/test.txt"

run_test "create_new_file" "Use the create tool to make a new file called tests/tool_tests/test_created_file.txt with content: 'This is a test file\nCreated by agentry\nWith multiple lines'"

run_test "grep_search" "Use the grep tool to search for the word 'test' in tests/tool_tests/sample_files/test.txt"
run_test "grep_regex" "Use the grep tool to search for lines containing numbers in tests/tool_tests/sample_files/test.txt"

run_test "insert_at_beginning" "Use the insert_at tool to add 'NEW LINE AT START' at the beginning of tests/tool_tests/test_created_file.txt"
run_test "edit_range_replace" "Use the edit_range tool to replace lines 2-3 in tests/tool_tests/test_created_file.txt with 'REPLACED LINE 2\nREPLACED LINE 3'"

run_test "search_replace_basic" "Use the search_replace tool to replace all occurrences of 'test' with 'TEST' in tests/tool_tests/test_created_file.txt"

# Enhanced exploration tools tests
echo -e "${YELLOW}=== Testing Exploration Tools ===${NC}"

run_test "ls_current_dir" "Use the ls tool to list contents of the current directory"
run_test "ls_test_dir" "Use the ls tool to list contents of the tests/tool_tests/sample_files directory"

run_test "find_go_files" "Use the find tool to locate all .go files in the current directory and subdirectories"
run_test "find_yaml_files" "Use the find tool to locate all .yaml files in the current directory and subdirectories"

run_test "glob_yaml" "Use the glob tool to find all *.yaml files in the current directory"
run_test "glob_recursive" "Use the glob tool to find files matching the pattern **/*.go"

# Web tools tests
echo -e "${YELLOW}=== Testing Web Tools ===${NC}"

run_test "api_get_request" "Use the api tool to make a GET request to https://httpbin.org/get"
run_test "api_post_request" "Use the api tool to make a POST request to https://httpbin.org/post with JSON data: {\"test\": \"data\"}"

run_test "read_webpage_httpbin" "Use the read_webpage tool to extract content from https://httpbin.org"

run_test "download_small_file" "Use the download tool to download a small file from https://httpbin.org/json and save it as tests/tool_tests/downloaded_file.json"

# Shell tools tests (platform-specific)
echo -e "${YELLOW}=== Testing Shell Tools ===${NC}"

run_test "bash_ls" "Use the bash tool to run 'ls -la tests/tool_tests/sample_files' and show directory contents"
run_test "bash_echo" "Use the bash tool to run 'echo \"Hello from bash\"'"
run_test "bash_date" "Use the bash tool to run 'date' and show current date/time"

run_test "sh_basic" "Use the sh tool to run 'pwd' and show current directory"
run_test "sh_env" "Use the sh tool to run 'echo \$HOME' and show home directory"

# Integration workflow tests
echo -e "${YELLOW}=== Testing Integration Workflows ===${NC}"

run_test "file_workflow" "Perform a complete file manipulation workflow: 1) Use ls to explore tests/tool_tests/sample_files, 2) Use view to display test.txt, 3) Use fileinfo to get details, 4) Use grep to search for 'Line', 5) Use read_lines to read lines 5-7"

run_test "exploration_workflow" "Perform file discovery: 1) Use ls to list current directory, 2) Use find to locate .yaml files, 3) Use glob to match *.go files, 4) Use grep to search for 'package' in found Go files"

run_test "web_workflow" "Perform web interaction: 1) Use api to GET https://httpbin.org/json, 2) Use read_webpage to extract content from https://httpbin.org, 3) Use fetch to get https://httpbin.org/get"

# Generate summary report
echo -e "${YELLOW}Generating test summary...${NC}"

total_tests=$(grep -c "^Test:" "$LOG_FILE" || echo "0")
passed_tests=$(grep -c "Result: PASS" "$LOG_FILE" || echo "0")
failed_tests=$(grep -c "Result: FAIL" "$LOG_FILE" || echo "0")

echo "" >> "$LOG_FILE"
echo "=== TEST SUMMARY ===" >> "$LOG_FILE"
echo "Total Tests: $total_tests" >> "$LOG_FILE"
echo "Passed: $passed_tests" >> "$LOG_FILE"
echo "Failed: $failed_tests" >> "$LOG_FILE"
echo "Success Rate: $(( passed_tests * 100 / total_tests ))%" >> "$LOG_FILE"
echo "Completed at: $(date)" >> "$LOG_FILE"

echo ""
echo -e "${GREEN}=== TEST SUMMARY ===${NC}"
echo -e "Total Tests: ${YELLOW}$total_tests${NC}"
echo -e "Passed: ${GREEN}$passed_tests${NC}"
echo -e "Failed: ${RED}$failed_tests${NC}"
echo -e "Success Rate: ${YELLOW}$(( passed_tests * 100 / total_tests ))%${NC}"
echo ""
echo -e "Detailed results saved to: ${YELLOW}$LOG_FILE${NC}"
echo -e "Individual test outputs saved to: ${YELLOW}$RESULTS_DIR/${NC}"

# Cleanup
cleanup_test_env

if [ "$failed_tests" -eq 0 ]; then
    echo -e "${GREEN}All tests passed! 🎉${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed. Check the log for details.${NC}"
    exit 1
fi
