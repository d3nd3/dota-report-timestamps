#!/bin/bash

set -e

cd "$(dirname "$0")"

if [ ! -f "go.mod" ]; then
    echo "ERROR: go.mod not found!"
    echo "Please run this script from the repository root directory."
    exit 1
fi

export GOLANG_PROTOBUF_REGISTRATION_CONFLICT=warn

echo "Building release binaries for all platforms..."
echo ""

DIST_DIR="dist"
rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"/{windows,mac,linux}

echo "Building Windows binaries (amd64)..."
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o "$DIST_DIR/windows/bot.exe" ./cmd/bot
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o "$DIST_DIR/windows/server.exe" ./cmd/server

echo "Building Mac binaries (amd64)..."
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o "$DIST_DIR/mac/bot" ./cmd/bot
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o "$DIST_DIR/mac/server" ./cmd/server

echo "Building Mac binaries (arm64)..."
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o "$DIST_DIR/mac/bot-arm64" ./cmd/bot
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o "$DIST_DIR/mac/server-arm64" ./cmd/server

echo "Building Linux binaries (amd64)..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o "$DIST_DIR/linux/bot" ./cmd/bot
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o "$DIST_DIR/linux/server" ./cmd/server

echo ""
echo "Build complete! Binaries are in the dist/ directory."

