#!/usr/bin/env bash

# Quick build fix to verify all import issues are resolved
set -e

echo "=== Testing build after import fixes ==="

# Try to build the main binary
echo "Building grafterm..."
if go build -o ./bin/grafterm ./cmd/grafterm; then
    echo "✅ Build successful!"
    echo "Binary created at: ./bin/grafterm"
    echo ""
    echo "Testing binary..."
    ./bin/grafterm --help
else
    echo "❌ Build failed. Checking specific packages..."
    
    # Test individual problematic packages
    echo "Testing prometheus package..."
    go build ./internal/service/metric/prometheus/
    
    echo "Testing graph widget package..."
    go build ./internal/view/page/widget/
    
    echo "Testing app package..."
    go build ./internal/view/
fi