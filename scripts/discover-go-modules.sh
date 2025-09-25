#!/bin/bash
set -e

# Go Module Discovery Script
# Finds all Go modules in the repository and returns their paths

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"

# Find all go.mod files and extract their directory paths
find "$REPO_ROOT" -name "go.mod" -type f -exec dirname {} \; | \
    sed "s|^$REPO_ROOT/||" | \
    sed "s|^$REPO_ROOT$|.|" | \
    sort
