#!/bin/bash
set -e

# Script to validate GitHub Actions workflows using actionlint
# This script ensures actionlint is available and runs validation

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
WORKFLOWS_DIR="$(cd "$PROJECT_DIR/../../workflows" && pwd)"

echo "🔍 Validating GitHub Actions workflows..."

# Change to project directory to ensure go.mod context
cd "$PROJECT_DIR"

# Ensure actionlint is available by installing it from our tools.go
echo "📦 Ensuring actionlint is available..."
go install github.com/rhysd/actionlint/cmd/actionlint

# Find actionlint binary
ACTIONLINT_BIN=""
if command -v actionlint >/dev/null 2>&1; then
    ACTIONLINT_BIN="actionlint"
elif [ -f "$GOPATH/bin/actionlint" ]; then
    ACTIONLINT_BIN="$GOPATH/bin/actionlint"
elif [ -f "$HOME/go/bin/actionlint" ]; then
    ACTIONLINT_BIN="$HOME/go/bin/actionlint"
elif [ -f "/go/bin/actionlint" ]; then
    ACTIONLINT_BIN="/go/bin/actionlint"
else
    echo "❌ actionlint binary not found after installation"
    exit 1
fi

echo "✅ Using actionlint: $ACTIONLINT_BIN"

# Validate all workflow files
WORKFLOW_FILES=$(find "$WORKFLOWS_DIR" -name "*.yml" -o -name "*.yaml" 2>/dev/null || true)

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
