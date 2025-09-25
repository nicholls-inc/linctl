#!/bin/bash
set -e

# Go Hook Runner Script
# Runs Go commands across all relevant modules based on changed files

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Validate environment
if ! git rev-parse --git-dir >/dev/null 2>&1; then
    echo -e "${RED}Error: Not in a git repository${NC}" >&2
    exit 1
fi

# Ensure we have required tools
if ! command -v go >/dev/null 2>&1; then
    echo -e "${RED}Error: Go is not installed or not in PATH${NC}" >&2
    exit 1
fi

# Check if timeout command is available, use fallback if not
TIMEOUT_CMD="timeout"
if ! command -v timeout >/dev/null 2>&1; then
    # On macOS, timeout might not be available, use gtimeout if available
    if command -v gtimeout >/dev/null 2>&1; then
        TIMEOUT_CMD="gtimeout"
    else
        echo -e "${YELLOW}Warning: timeout command not available, commands may hang${NC}" >&2
        TIMEOUT_CMD=""
    fi
fi

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

    # Get all Go modules (project-only)
    local all_modules
    mapfile -t all_modules < <(get_go_modules)

    # Filter out any system paths that might have slipped through
    local project_modules=()
    for module in "${all_modules[@]}"; do
        # Only include modules that are relative paths or current directory
        if [[ "$module" == "." ]] || [[ "$module" != /* ]]; then
            project_modules+=("$module")
        fi
    done

    # For each project module, check if any changed files are in that module
    for module in "${project_modules[@]}"; do
        for file in "${changed_files[@]}"; do
            # Check if file is in this module's directory
            if [[ "$file" == "$module"/* ]] || [[ "$file" == "$module" ]]; then
                affected_modules+=("$module")
                break
            fi
        done
    done

    # If no specific files provided or no modules matched, run on all project modules
    if [[ ${#affected_modules[@]} -eq 0 ]]; then
        affected_modules=("${project_modules[@]}")
    fi

    printf '%s\n' "${affected_modules[@]}"
}

run_go_fmt() {
    local modules=("$@")
    local exit_code=0

    for module in "${modules[@]}"; do
        echo -e "${BLUE}Running go fmt in $module${NC}"
        local module_path="$REPO_ROOT"
        if [[ "$module" != "." ]]; then
            module_path="$REPO_ROOT/$module"
        fi

        cd "$module_path"
        if [[ -n "$TIMEOUT_CMD" ]]; then
            timeout_gofmt="$TIMEOUT_CMD 30s"
        else
            timeout_gofmt=""
        fi
        if ! $timeout_gofmt gofmt -s -w . || ! git diff --exit-code; then
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
        local module_path="$REPO_ROOT"
        if [[ "$module" != "." ]]; then
            module_path="$REPO_ROOT/$module"
        fi

        cd "$module_path"
        if [[ -n "$TIMEOUT_CMD" ]]; then
            timeout_vet="$TIMEOUT_CMD 60s"
        else
            timeout_vet=""
        fi
        if ! $timeout_vet go vet ./...; then
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
        local module_path="$REPO_ROOT"
        if [[ "$module" != "." ]]; then
            module_path="$REPO_ROOT/$module"
        fi

        cd "$module_path"
        # Try to build all packages, which handles both main packages and libraries
        if [[ -n "$TIMEOUT_CMD" ]]; then
            timeout_build="$TIMEOUT_CMD 120s"
        else
            timeout_build=""
        fi
        if ! $timeout_build go build ./... 2>/dev/null; then
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
        local module_path="$REPO_ROOT"
        if [[ "$module" != "." ]]; then
            module_path="$REPO_ROOT/$module"
        fi

        cd "$module_path"
        if [[ -n "$TIMEOUT_CMD" ]]; then
            timeout_tidy="$TIMEOUT_CMD 60s"
        else
            timeout_tidy=""
        fi
        if ! $timeout_tidy go mod tidy || ! git diff --exit-code go.mod go.sum; then
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
        local module_path="$REPO_ROOT"
        if [[ "$module" != "." ]]; then
            module_path="$REPO_ROOT/$module"
        fi

        cd "$module_path"
        if [[ -n "$TIMEOUT_CMD" ]]; then
            timeout_verify="$TIMEOUT_CMD 30s"
        else
            timeout_verify=""
        fi
        if ! $timeout_verify go mod verify; then
            echo -e "${RED}go mod verify failed in $module${NC}"
            exit_code=1
        fi
    done

    return $exit_code
}

run_staticcheck() {
    local modules=("$@")
    local exit_code=0

    # Check if staticcheck is available
    if ! command -v staticcheck >/dev/null 2>&1; then
        echo -e "${YELLOW}Warning: staticcheck not found, attempting to install...${NC}"
        if ! go install honnef.co/go/tools/cmd/staticcheck@latest; then
            echo -e "${RED}Failed to install staticcheck${NC}"
            return 1
        fi
    fi

    for module in "${modules[@]}"; do
        echo -e "${BLUE}Running staticcheck in $module${NC}"
        local module_path="$REPO_ROOT"
        if [[ "$module" != "." ]]; then
            module_path="$REPO_ROOT/$module"
        fi

        cd "$module_path"
        if [[ -n "$TIMEOUT_CMD" ]]; then
            timeout_staticcheck="$TIMEOUT_CMD 120s"
        else
            timeout_staticcheck=""
        fi
        if ! $timeout_staticcheck staticcheck ./...; then
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
