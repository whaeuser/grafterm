#!/usr/bin/env bash

# Simple integration test script without Docker
set -e

echo "Running integration tests..."

# Run integration tests with integration tag
echo "=== Integration Tests ==="
go test $(go list ./... | grep -v vendor) -v -tags='integration'

echo "Integration tests completed successfully!"