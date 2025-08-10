#!/usr/bin/env bash

# Fix build errors by cleaning up imports and dependencies
set -e

echo "=== Fixing build errors ==="

# Remove unused imports automatically
echo "Removing unused imports..."
goimports -w ./internal/ 2>/dev/null || echo "goimports not found, skipping"

# Fix imports manually for problematic files
echo "Checking for import issues..."

# Run go mod tidy to ensure dependencies are correct
go mod tidy

# Try to build to see if issues are resolved
echo "Testing build..."
if go build ./cmd/grafterm; then
    echo "✅ Build successful!"
    echo "You can now run: ./build.sh"
else
    echo "❌ Build still has issues. Let me check specific errors..."
    
    # Try to get more specific error information
    echo "Checking specific package errors..."
    go build ./internal/view/app.go
    go build ./internal/view/page/widget/graph.go
    go build ./internal/service/metric/prometheus/prometheus.go
fi

echo ""
echo "If issues persist, try running 'go fmt ./...' to format code"
echo "or check the specific error messages above."