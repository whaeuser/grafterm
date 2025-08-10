#!/usr/bin/env bash

# Docker Build Alternative (wenn Go nicht installiert ist)
set -e

# Prüfe ob Docker verfügbar ist
if ! command -v docker &> /dev/null; then
    echo "❌ Docker ist nicht installiert. Bitte installiere Docker oder verwende:"
    echo "./install-go.sh"
    exit 1
fi

echo "✅ Docker gefunden - baue grafterm mit Docker..."

# Get version from git
VERSION=$(git describe --tags --always 2>/dev/null || echo "dev")

# Create output directory
mkdir -p ./bin

# Build with Docker
echo "Baue grafterm version: $VERSION"
docker run --rm \
    -v "$(pwd):/src" \
    -w /src \
    golang:latest \
    go build \
        -ldflags "-w -extldflags '-static' -X main.Version=${VERSION}" \
        -o /src/bin/grafterm \
        ./cmd/grafterm

# Make binary executable
chmod +x ./bin/grafterm

echo "✅ Build abgeschlossen!"
echo "Binary: $(realpath ./bin/grafterm)"
echo ""
echo "Teste Binary:"
./bin/grafterm --help