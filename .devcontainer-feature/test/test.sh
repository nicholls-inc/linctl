#!/bin/bash
set -e

# Test script for linctl DevContainer Feature

echo "🧪 Testing linctl DevContainer Feature..."

# Test 1: Check if linctl is installed and in PATH
echo "Test 1: Checking if linctl is installed..."
if command -v linctl >/dev/null 2>&1; then
    echo "✅ linctl is installed and in PATH"
else
    echo "❌ linctl is not found in PATH"
    exit 1
fi

# Test 2: Check version command
echo "Test 2: Checking version command..."
if linctl --version >/dev/null 2>&1; then
    echo "✅ linctl --version works"
    linctl --version
else
    echo "❌ linctl --version failed"
    exit 1
fi

# Test 3: Check help command
echo "Test 3: Checking help command..."
if linctl --help >/dev/null 2>&1; then
    echo "✅ linctl --help works"
else
    echo "❌ linctl --help failed"
    exit 1
fi

# Test 4: Check binary permissions
echo "Test 4: Checking binary permissions..."
LINCTL_PATH=$(which linctl)
if [ -x "$LINCTL_PATH" ]; then
    echo "✅ linctl binary is executable"
    ls -la "$LINCTL_PATH"
else
    echo "❌ linctl binary is not executable"
    exit 1
fi

# Test 5: Check if binary is in expected location
echo "Test 5: Checking binary location..."
if [ -f "/usr/local/bin/linctl" ]; then
    echo "✅ linctl is installed in /usr/local/bin/"
else
    echo "⚠️  linctl is not in /usr/local/bin/ (may be in different location)"
fi

echo "🎉 All tests passed! linctl DevContainer Feature is working correctly."
