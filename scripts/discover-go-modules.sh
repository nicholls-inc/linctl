#!/bin/bash
set -e

# Go Module Discovery Script
# Finds all Go modules in the repository and returns their paths

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Validate that we're in a git repository
if ! git rev-parse --git-dir >/dev/null 2>&1; then
    echo "Error: Not in a git repository" >&2
    exit 1
fi

# Get the actual git repository root to ensure we stay within project bounds
GIT_ROOT="$(git rev-parse --show-toplevel)"

# Ensure REPO_ROOT is within or equal to GIT_ROOT for security
if [[ "$REPO_ROOT" != "$GIT_ROOT"* ]]; then
    echo "Error: Repository root outside git repository" >&2
    exit 1
fi

# Find all go.mod files within the project directory only
# Use maxdepth to prevent deep traversal and exclude hidden directories
find "$REPO_ROOT" -maxdepth 3 -name "go.mod" -type f -not -path "*/.*" -exec dirname {} \; | \
    sed "s|^$REPO_ROOT/||" | \
    sed "s|^$REPO_ROOT$|.|" | \
    grep -E "^(\.|[^/])" | \
    sort
