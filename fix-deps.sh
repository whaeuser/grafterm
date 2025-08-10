#!/usr/bin/env bash

# Fix dependencies and update go.sum
set -e

echo "=== Fixing Go dependencies ==="

# Download dependencies and update go.sum
echo "Running 'go mod tidy' to fix missing dependencies..."
go mod tidy

echo "Updating go.sum with 'go mod download'..."
go mod download

echo "Verifying dependencies..."
go mod verify

echo ""
echo "✅ Dependencies fixed!"
echo ""
echo "Now try building again:"
echo "./build.sh"