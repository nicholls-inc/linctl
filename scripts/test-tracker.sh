#!/bin/bash
set -e

# Test Tracker Script
# Manages test verification hashes to ensure tests are run before commits

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"

# Default to pr-agent module for backward compatibility
DEFAULT_MODULE_PATH=".github/actions/pr-agent"
MODULE_PATH=""
TEST_TRACKER_FILE=""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

usage() {
    echo "Usage: $0 [--module-path <path>] {update|verify|status|clean}"
    echo ""
    echo "Options:"
    echo "  --module-path <path>  - Specify Go module path (default: $DEFAULT_MODULE_PATH)"
    echo ""
    echo "Commands:"
    echo "  update  - Update test tracker after running tests (call this after 'go test')"
    echo "  verify  - Verify tests have been run for current code state"
    echo "  status  - Show current test tracker status"
    echo "  clean   - Remove test tracker file"
    echo ""
    echo "Example workflow:"
    echo "  1. Make code changes in a module"
    echo "  2. Run: cd <module-path> && go test -short -v ./..."
    echo "  3. Run: $0 --module-path <module-path> update"
    echo "  4. Commit (pre-commit hook will verify)"
}

init_module_path() {
    # Parse command line arguments for module path
    while [[ $# -gt 0 ]]; do
        case $1 in
            --module-path)
                MODULE_PATH="$2"
                shift 2
                ;;
            *)
                break
                ;;
        esac
    done

    # Use default if not specified
    if [ -z "$MODULE_PATH" ]; then
        MODULE_PATH="$DEFAULT_MODULE_PATH"
    fi

    # Set up paths
    PROJECT_DIR="$REPO_ROOT/$MODULE_PATH"
    MODULE_NAME=$(echo "$MODULE_PATH" | tr '/' '-')
    TEST_TRACKER_FILE="$PROJECT_DIR/.test-tracker-$MODULE_NAME"

    # Verify module exists and has go.mod
    if [ ! -f "$PROJECT_DIR/go.mod" ]; then
        echo -e "${RED}❌ No go.mod found in $MODULE_PATH${NC}"
        echo -e "${YELLOW}💡 Available Go modules:${NC}"
        "$SCRIPT_DIR/discover-go-modules.sh" | sed 's/^/   /'
        exit 1
    fi
}

calculate_test_hash() {
    cd "$PROJECT_DIR"

    # Hash all Go source files and test files in this module
    find . -name "*.go" -type f | sort | xargs sha256sum | sha256sum | cut -d' ' -f1
}

calculate_test_result_hash() {
    cd "$PROJECT_DIR"

    # Run tests with -short flag to skip integration/E2E tests (same as CI)
    # Also set SKIP_INTEGRATION environment variable for consistency with CI
    if SKIP_INTEGRATION=true go test -short -v ./... > /tmp/test_output.log 2>&1; then
        TEST_RESULT="PASSED"
        TEST_EXIT_CODE=0
    else
        TEST_RESULT="FAILED"
        TEST_EXIT_CODE=1
    fi

    # Hash the test output for verification
    OUTPUT_HASH=$(sha256sum /tmp/test_output.log | cut -d' ' -f1)
    rm -f /tmp/test_output.log

    echo "${TEST_RESULT}:${OUTPUT_HASH}"
    return $TEST_EXIT_CODE
}

update_tracker() {
    echo -e "${BLUE}🔄 Updating test tracker for module: $MODULE_PATH${NC}"

    cd "$PROJECT_DIR"

    # Calculate current code hash
    CODE_HASH=$(calculate_test_hash)
    echo -e "${BLUE}📊 Code hash: ${CODE_HASH}${NC}"

    # Run tests and get result hash
    echo -e "${BLUE}🧪 Running tests to verify current state...${NC}"
    if TEST_RESULT_INFO=$(calculate_test_result_hash); then
        echo -e "${GREEN}✅ Tests passed!${NC}"
    else
        echo -e "${RED}❌ Tests failed! Cannot update tracker with failing tests.${NC}"
        exit 1
    fi

    # Create tracker file
    TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    cat > "$TEST_TRACKER_FILE" << EOF
# Test Tracker File - DO NOT EDIT MANUALLY
# This file tracks when tests were last run successfully for module: $MODULE_PATH
# Generated: $TIMESTAMP

MODULE_PATH=$MODULE_PATH
CODE_HASH=$CODE_HASH
TEST_RESULT=$TEST_RESULT_INFO
TIMESTAMP=$TIMESTAMP
COMMIT_HASH=$(git rev-parse HEAD 2>/dev/null || echo "unknown")
EOF

    echo -e "${GREEN}✅ Test tracker updated successfully!${NC}"
    echo -e "${YELLOW}💡 You can now commit your changes.${NC}"
}

verify_tracker() {
    if [ ! -f "$TEST_TRACKER_FILE" ]; then
        echo -e "${RED}❌ No test tracker found for module: $MODULE_PATH${NC}"
        echo -e "${YELLOW}📝 You must run tests before committing:${NC}"
        echo -e "   1. Run: ${BLUE}cd $MODULE_PATH && go test -short -v ./...${NC}"
        echo -e "   2. Run: ${BLUE}$0 --module-path $MODULE_PATH update${NC}"
        echo -e "   3. Then commit your changes"
        return 1
    fi

    # Load tracker data
    source "$TEST_TRACKER_FILE"

    # Calculate current code hash
    CURRENT_HASH=$(calculate_test_hash)

    if [ "$CODE_HASH" != "$CURRENT_HASH" ]; then
        echo -e "${RED}❌ Code has changed since tests were last run!${NC}"
        echo -e "${YELLOW}📝 Current code hash: ${CURRENT_HASH}${NC}"
        echo -e "${YELLOW}📝 Tracked code hash: ${CODE_HASH}${NC}"
        echo -e "${YELLOW}📝 You must run tests for the current code state:${NC}"
        echo -e "   1. Run: ${BLUE}cd $MODULE_PATH && go test -short -v ./...${NC}"
        echo -e "   2. Run: ${BLUE}$0 --module-path $MODULE_PATH update${NC}"
        echo -e "   3. Then commit your changes"
        return 1
    fi

    # Check if tests passed
    if [[ "$TEST_RESULT" != "PASSED:"* ]]; then
        echo -e "${RED}❌ Last test run did not pass!${NC}"
        echo -e "${YELLOW}📝 You must ensure tests pass before committing:${NC}"
        echo -e "   1. Fix failing tests"
        echo -e "   2. Run: ${BLUE}cd $MODULE_PATH && go test -short -v ./...${NC}"
        echo -e "   3. Run: ${BLUE}$0 --module-path $MODULE_PATH update${NC}"
        return 1
    fi

    echo -e "${GREEN}✅ Tests have been run and passed for current code state!${NC}"
    echo -e "${BLUE}📦 Module: ${MODULE_PATH}${NC}"
    echo -e "${BLUE}📊 Code hash: ${CODE_HASH}${NC}"
    echo -e "${BLUE}🕐 Last run: ${TIMESTAMP}${NC}"
    return 0
}

show_status() {
    if [ ! -f "$TEST_TRACKER_FILE" ]; then
        echo -e "${YELLOW}📝 No test tracker found for module: $MODULE_PATH${NC}"
        echo -e "${YELLOW}💡 Run tests and then: $0 --module-path $MODULE_PATH update${NC}"
        return
    fi

    source "$TEST_TRACKER_FILE"
    CURRENT_HASH=$(calculate_test_hash)

    echo -e "${BLUE}📊 Test Tracker Status${NC}"
    echo -e "   Module: ${MODULE_PATH}"
    echo -e "   Tracked code hash: ${CODE_HASH}"
    echo -e "   Current code hash: ${CURRENT_HASH}"
    echo -e "   Test result: ${TEST_RESULT}"
    echo -e "   Last updated: ${TIMESTAMP}"
    echo -e "   Commit: ${COMMIT_HASH}"

    if [ "$CODE_HASH" = "$CURRENT_HASH" ]; then
        echo -e "${GREEN}✅ Code matches tracked state${NC}"
    else
        echo -e "${YELLOW}⚠️  Code has changed since last test run${NC}"
    fi
}

clean_tracker() {
    if [ -f "$TEST_TRACKER_FILE" ]; then
        rm "$TEST_TRACKER_FILE"
        echo -e "${GREEN}✅ Test tracker cleaned${NC}"
    else
        echo -e "${YELLOW}📝 No test tracker to clean${NC}"
    fi
}

# Initialize module path and parse arguments
init_module_path "$@"

# Skip parsed arguments to get to the command
while [[ $# -gt 0 ]]; do
    case $1 in
        --module-path)
            shift 2
            ;;
        *)
            break
            ;;
    esac
done

# Main command handling
case "${1:-}" in
    update)
        update_tracker
        ;;
    verify)
        verify_tracker
        ;;
    status)
        show_status
        ;;
    clean)
        clean_tracker
        ;;
    *)
        usage
        exit 1
        ;;
esac
