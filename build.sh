#!/usr/bin/env bash

# Simple build script for grafterm without Docker
set -e

# Get version from git
VERSION=$(git describe --tags --always 2>/dev/null || echo "dev")

# Set build variables
SRC="./cmd/grafterm"
OUTPUT_DIR="./bin"
BINARY_NAME="grafterm"

# Create output directory
mkdir -p "${OUTPUT_DIR}"

# Detect OS and set appropriate binary name
OS="$(uname -s)"
case "${OS}" in
    Linux*)     OS_NAME="linux"; EXTENSION="";;
    Darwin*)    OS_NAME="darwin"; EXTENSION="";;
    CYGWIN*)    OS_NAME="windows"; EXTENSION=".exe";;
    MINGW*)     OS_NAME="windows"; EXTENSION=".exe";;
    *)          OS_NAME="unknown"; EXTENSION="";;
esac

# Set final output path
FINAL_OUTPUT="${OUTPUT_DIR}/${BINARY_NAME}${EXTENSION}"

# Build flags
LDFLAGS="-w -extldflags '-static'"
VERSION_FLAG="-X main.Version=${VERSION}"

echo "Building grafterm version: ${VERSION}"
echo "Target OS: ${OS_NAME}"
echo "Output: ${FINAL_OUTPUT}"

# Build the binary
go build \
    -ldflags "${LDFLAGS} ${VERSION_FLAG}" \
    -o "${FINAL_OUTPUT}" \
    "${SRC}"

echo "Build completed successfully!"
echo "Binary location: $(realpath "${FINAL_OUTPUT}")"

# Make binary executable (except on Windows)
if [ "${EXTENSION}" != ".exe" ]; then
    chmod +x "${FINAL_OUTPUT}"
fi