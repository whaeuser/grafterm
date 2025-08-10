#!/usr/bin/env bash

# Simple test script without Docker
set -e

echo "Running unit tests..."

# Run all unit tests excluding vendor directory
echo "=== All Unit Tests ==="
go test $(go list ./... | grep -v vendor) -v

echo "Unit tests completed successfully!"