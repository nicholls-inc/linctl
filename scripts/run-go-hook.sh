#!/bin/bash
set -e

# Go Hook Runner Script
# Runs Go commands across all relevant modules based on changed files

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

usage() {
    echo "Usage: $0 <command> [changed_files...]"
    echo ""
    echo "Commands:"
    echo "  fmt       - Run go fmt"
    echo "  vet       - Run go vet"
    echo "  build     - Run go build"
    echo "  mod-tidy  - Run go mod tidy"
    echo "  mod-verify - Run go mod verify"
    echo "  staticcheck - Run staticcheck"
    echo ""
    echo "The script will determine which Go modules need the command based on changed files."
}

# Get all Go modules
get_go_modules() {
    "$SCRIPT_DIR/discover-go-modules.sh"
}

# Determine which modules are affected by the changed files
get_affected_modules() {
    local changed_files=("$@")
    local affected_modules=()

    # Get all Go modules
    local all_modules
    mapfile -t all_modules < <(get_go_modules)

    # For each module, check if any changed files are in that module
    for module in "${all_modules[@]}"; do
        for file in "${changed_files[@]}"; do
            # Check if file is in this module's directory
            if [[ "$file" == "$module"/* ]] || [[ "$file" == "$module" ]]; then
                affected_modules+=("$module")
                break
            fi
        done
    done

    # If no specific files provided or no modules matched, run on all modules
    if [[ ${#affected_modules[@]} -eq 0 ]]; then
        affected_modules=("${all_modules[@]}")
    fi

    printf '%s\n' "${affected_modules[@]}"
}

run_go_fmt() {
    local modules=("$@")
    local exit_code=0

    for module in "${modules[@]}"; do
        echo -e "${BLUE}Running go fmt in $module${NC}"
        cd "$REPO_ROOT/$module"
        if ! gofmt -s -w . || ! git diff --exit-code; then
            echo -e "${RED}go fmt failed or found formatting issues in $module${NC}"
            exit_code=1
        fi
    done

    return $exit_code
}

run_go_vet() {
    local modules=("$@")
    local exit_code=0

    for module in "${modules[@]}"; do
        echo -e "${BLUE}Running go vet in $module${NC}"
        cd "$REPO_ROOT/$module"
        if ! go vet ./...; then
            echo -e "${RED}go vet failed in $module${NC}"
            exit_code=1
        fi
    done

    return $exit_code
}

run_go_build() {
    local modules=("$@")
    local exit_code=0

    for module in "${modules[@]}"; do
        echo -e "${BLUE}Running go build in $module${NC}"
        cd "$REPO_ROOT/$module"

        # Try to build all packages, which handles both main packages and libraries
        if ! go build ./... 2>/dev/null; then
            echo -e "${RED}go build failed in $module${NC}"
            exit_code=1
        fi
    done

    return $exit_code
}

run_go_mod_tidy() {
    local modules=("$@")
    local exit_code=0

    for module in "${modules[@]}"; do
        echo -e "${BLUE}Running go mod tidy in $module${NC}"
        cd "$REPO_ROOT/$module"
        if ! go mod tidy || ! git diff --exit-code go.mod go.sum; then
            echo -e "${RED}go mod tidy failed or found changes in $module${NC}"
            exit_code=1
        fi
    done

    return $exit_code
}

run_go_mod_verify() {
    local modules=("$@")
    local exit_code=0

    for module in "${modules[@]}"; do
        echo -e "${BLUE}Running go mod verify in $module${NC}"
        cd "$REPO_ROOT/$module"
        if ! go mod verify; then
            echo -e "${RED}go mod verify failed in $module${NC}"
            exit_code=1
        fi
    done

    return $exit_code
}

run_staticcheck() {
    local modules=("$@")
    local exit_code=0

    for module in "${modules[@]}"; do
        echo -e "${BLUE}Running staticcheck in $module${NC}"
        cd "$REPO_ROOT/$module"
        if ! staticcheck ./...; then
            echo -e "${RED}staticcheck failed in $module${NC}"
            exit_code=1
        fi
    done

    return $exit_code
}

# Main execution
COMMAND="${1:-}"
shift || true

if [[ -z "$COMMAND" ]]; then
    usage
    exit 1
fi

# Get affected modules based on changed files (if any provided)
mapfile -t affected_modules < <(get_affected_modules "$@")

if [[ ${#affected_modules[@]} -eq 0 ]]; then
    echo -e "${GREEN}No Go modules affected by changes${NC}"
    exit 0
fi

echo -e "${BLUE}Affected modules: ${affected_modules[*]}${NC}"

case "$COMMAND" in
    fmt)
        run_go_fmt "${affected_modules[@]}"
        ;;
    vet)
        run_go_vet "${affected_modules[@]}"
        ;;
    build)
        run_go_build "${affected_modules[@]}"
        ;;
    mod-tidy)
        run_go_mod_tidy "${affected_modules[@]}"
        ;;
    mod-verify)
        run_go_mod_verify "${affected_modules[@]}"
        ;;
    staticcheck)
        run_staticcheck "${affected_modules[@]}"
        ;;
    *)
        echo "Unknown command: $COMMAND"
        usage
        exit 1
        ;;
esac
