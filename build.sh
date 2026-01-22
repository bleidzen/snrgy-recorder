#!/bin/bash
set -e

echo "Building SNRGY Recorder..."

# Download dependencies
go mod tidy

# Detect architecture
ARCH=$(uname -m)

if [ "$ARCH" = "arm64" ]; then
    echo "Building for macOS (Apple Silicon)..."
    go build -ldflags="-s -w" -o snrgy-recorder .
else
    echo "Building for macOS (Intel)..."
    go build -ldflags="-s -w" -o snrgy-recorder .
fi

echo "Done! Output: snrgy-recorder"
echo ""
echo "To create a .app bundle, run: ./create-app-bundle.sh"
