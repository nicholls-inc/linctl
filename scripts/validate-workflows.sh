#!/bin/bash
set -e

# Script to validate GitHub Actions workflows using actionlint
# This script ensures actionlint is available and runs validation

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
WORKFLOWS_DIR="$PROJECT_DIR/.github/workflows"

echo "🔍 Validating GitHub Actions workflows..."

# Validate environment
if ! git rev-parse --git-dir >/dev/null 2>&1; then
    echo "❌ Error: Not in a git repository" >&2
    exit 1
fi

if ! command -v go >/dev/null 2>&1; then
    echo "❌ Error: Go is not installed or not in PATH" >&2
    exit 1
fi

# Change to project directory to ensure go.mod context
cd "$PROJECT_DIR"

# Ensure actionlint is available
echo "📦 Ensuring actionlint is available..."
if ! command -v actionlint >/dev/null 2>&1; then
    echo "Installing actionlint..."
    if ! go install github.com/rhysd/actionlint/cmd/actionlint@latest; then
        echo "❌ Failed to install actionlint" >&2
        exit 1
    fi
fi

# Find actionlint binary with fallback paths
ACTIONLINT_BIN=""
if command -v actionlint >/dev/null 2>&1; then
    ACTIONLINT_BIN="actionlint"
else
    # Try common Go binary paths
    GOPATH_BIN="${GOPATH:-$HOME/go}/bin/actionlint"
    if [ -f "$GOPATH_BIN" ]; then
        ACTIONLINT_BIN="$GOPATH_BIN"
    elif [ -f "/go/bin/actionlint" ]; then
        ACTIONLINT_BIN="/go/bin/actionlint"
    else
        echo "❌ actionlint binary not found after installation" >&2
        echo "Tried paths: actionlint (PATH), $GOPATH_BIN, /go/bin/actionlint" >&2
        exit 1
    fi
fi

echo "✅ Using actionlint: $ACTIONLINT_BIN"

# Validate all workflow files
WORKFLOW_FILES=""
if [ -d "$WORKFLOWS_DIR" ]; then
    WORKFLOW_FILES=$(find "$WORKFLOWS_DIR" -name "*.yml" -o -name "*.yaml" 2>/dev/null || true)
fi

if [ -z "$WORKFLOW_FILES" ]; then
    echo "ℹ️  No workflow files found in $WORKFLOWS_DIR"
    exit 0
fi

echo "🔍 Validating workflow files:"
echo "$WORKFLOW_FILES" | while read -r file; do
    echo "  - $(basename "$file")"
done

# Run actionlint on all workflow files
if echo "$WORKFLOW_FILES" | xargs "$ACTIONLINT_BIN"; then
    echo "✅ All workflow files are valid!"
else
    echo "❌ Workflow validation failed!"
    exit 1
fi
