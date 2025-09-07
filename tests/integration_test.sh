#!/bin/bash

# Integration tests for gh-issue-dependency CLI
# This script tests the CLI end-to-end functionality

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Function to print colored output
print_status() {
    local status=$1
    local message=$2
    case $status in
        "PASS")
            echo -e "${GREEN}[PASS]${NC} $message"
            ((TESTS_PASSED++))
            ;;
        "FAIL")
            echo -e "${RED}[FAIL]${NC} $message"
            ((TESTS_FAILED++))
            ;;
        "INFO")
            echo -e "${BLUE}[INFO]${NC} $message"
            ;;
        "WARN")
            echo -e "${YELLOW}[WARN]${NC} $message"
            ;;
    esac
}

# Function to run a test
run_test() {
    local test_name=$1
    local expected_exit_code=${2:-0}
    shift 2
    local cmd=("$@")
    
    ((TESTS_RUN++))
    print_status "INFO" "Running test: $test_name"
    
    # Capture both stdout and stderr
    local output
    local exit_code
    if output=$("${cmd[@]}" 2>&1); then
        exit_code=0
    else
        exit_code=$?
    fi
    
    if [ $exit_code -eq $expected_exit_code ]; then
        print_status "PASS" "$test_name"
        return 0
    else
        print_status "FAIL" "$test_name (expected exit code $expected_exit_code, got $exit_code)"
        echo "Command: ${cmd[*]}"
        echo "Output: $output"
        return 1
    fi
}

# Function to test output contains expected text
test_output_contains() {
    local test_name=$1
    local expected_text=$2
    shift 2
    local cmd=("$@")
    
    ((TESTS_RUN++))
    print_status "INFO" "Running test: $test_name"
    
    local output
    if output=$("${cmd[@]}" 2>&1); then
        if echo "$output" | grep -q "$expected_text"; then
            print_status "PASS" "$test_name"
            return 0
        else
            print_status "FAIL" "$test_name (output does not contain '$expected_text')"
            echo "Command: ${cmd[*]}"
            echo "Output: $output"
            return 1
        fi
    else
        local exit_code=$?
        print_status "FAIL" "$test_name (command failed with exit code $exit_code)"
        echo "Command: ${cmd[*]}"
        echo "Output: $output"
        return 1
    fi
}

# Function to test that output does NOT contain text
test_output_not_contains() {
    local test_name=$1
    local unexpected_text=$2
    shift 2
    local cmd=("$@")
    
    ((TESTS_RUN++))
    print_status "INFO" "Running test: $test_name"
    
    local output
    if output=$("${cmd[@]}" 2>&1); then
        if ! echo "$output" | grep -q "$unexpected_text"; then
            print_status "PASS" "$test_name"
            return 0
        else
            print_status "FAIL" "$test_name (output contains unexpected '$unexpected_text')"
            echo "Command: ${cmd[*]}"
            echo "Output: $output"
            return 1
        fi
    else
        local exit_code=$?
        print_status "FAIL" "$test_name (command failed with exit code $exit_code)"
        echo "Command: ${cmd[*]}"
        echo "Output: $output"
        return 1
    fi
}

# Main test function
main() {
    echo -e "${BLUE}=== gh-issue-dependency Integration Tests ===${NC}"
    echo ""
    
    # Build the binary first
    print_status "INFO" "Building gh-issue-dependency binary..."
    if ! go build -o gh-issue-dependency .; then
        print_status "FAIL" "Failed to build binary"
        exit 1
    fi
    print_status "PASS" "Binary built successfully"
    
    # Test 1: Binary executes without errors
    run_test "Binary executes without errors" 0 ./gh-issue-dependency
    
    # Test 2: Help flag works
    test_output_contains "Help flag shows usage" "A GitHub CLI extension" ./gh-issue-dependency --help
    
    # Test 3: Version flag works
    test_output_contains "Version flag shows version" "gh-issue-dependency version" ./gh-issue-dependency --version
    
    # Test 4: Short help flag works
    test_output_contains "Short help flag works" "A GitHub CLI extension" ./gh-issue-dependency -h
    
    # Test 5: Root command shows help by default
    test_output_contains "Root command shows help" "Available Commands:" ./gh-issue-dependency
    
    # Test 6: Invalid flag shows error
    run_test "Invalid flag returns error" 2 ./gh-issue-dependency --invalid-flag
    
    # Test 7: Global repo flag is recognized
    test_output_not_contains "Global repo flag recognized" "unknown flag" ./gh-issue-dependency --repo owner/repo --help
    
    # Test 8: Short repo flag is recognized
    test_output_not_contains "Short repo flag recognized" "unknown flag" ./gh-issue-dependency -R owner/repo --help
    
    # Test 9: Subcommands are listed
    test_output_contains "Add subcommand listed" "add" ./gh-issue-dependency --help
    test_output_contains "List subcommand listed" "list" ./gh-issue-dependency --help  
    test_output_contains "Remove subcommand listed" "remove" ./gh-issue-dependency --help
    
    # Test 10: Help format is proper
    test_output_contains "Usage section present" "Usage:" ./gh-issue-dependency --help
    test_output_contains "Flags section present" "Flags:" ./gh-issue-dependency --help
    test_output_contains "Examples section present" "Examples:" ./gh-issue-dependency --help
    
    # Test 11: Version template format
    test_output_contains "Version format correct" "gh-issue-dependency version" ./gh-issue-dependency --version
    
    # Test 12: No panic on empty args
    run_test "No panic on empty args" 0 ./gh-issue-dependency
    
    # Test 13: Subcommand help works
    test_output_contains "Add command help" "Add dependency" ./gh-issue-dependency add --help
    test_output_contains "List command help" "List dependencies" ./gh-issue-dependency list --help
    test_output_contains "Remove command help" "Remove dependency" ./gh-issue-dependency remove --help
    
    # Test 14: Invalid subcommand shows error
    run_test "Invalid subcommand returns error" 2 ./gh-issue-dependency invalid-command
    
    # Test 15: Help text contains examples
    test_output_contains "Help contains examples" "gh issue-dependency list" ./gh-issue-dependency --help
    test_output_contains "Help contains add example" "gh issue-dependency add" ./gh-issue-dependency --help
    test_output_contains "Help contains remove example" "gh issue-dependency remove" ./gh-issue-dependency --help
    
    # Test 16: Repository flag validation (if applicable)
    # This test depends on the actual implementation but tests error handling
    run_test "Invalid repo format handled" 2 ./gh-issue-dependency -R "invalid" list || true
    
    # Test 17: Test build with different flags
    print_status "INFO" "Testing build with version flag..."
    if go build -ldflags "-X cmd.Version=test-1.0.0" -o gh-issue-dependency-versioned .; then
        test_output_contains "Custom version in binary" "test-1.0.0" ./gh-issue-dependency-versioned --version
        rm -f gh-issue-dependency-versioned
    else
        print_status "WARN" "Could not test custom version build"
    fi
    
    # Clean up
    rm -f gh-issue-dependency gh-issue-dependency.exe
    
    # Summary
    echo ""
    echo -e "${BLUE}=== Test Summary ===${NC}"
    echo "Tests run: $TESTS_RUN"
    echo "Tests passed: $TESTS_PASSED"  
    echo "Tests failed: $TESTS_FAILED"
    
    if [ $TESTS_FAILED -eq 0 ]; then
        echo -e "${GREEN}All tests passed!${NC}"
        exit 0
    else
        echo -e "${RED}$TESTS_FAILED test(s) failed!${NC}"
        exit 1
    fi
}

# Check if we're in the right directory
if [ ! -f "go.mod" ] || [ ! -f "main.go" ]; then
    echo -e "${RED}Error: This script must be run from the project root directory${NC}"
    echo "Expected files: go.mod, main.go"
    exit 1
fi

# Run main function
main "$@"