#!/bin/bash

set -e

# Suppress protobuf registration conflicts (critical for Dota 2 protobuf dependencies)
export GOLANG_PROTOBUF_REGISTRATION_CONFLICT=warn

# Support an optional update flag (e.g. ./launch-linux.sh --update)
if [ "$1" == "--update" ] || [ "$1" == "-u" ]; then
    if command -v go >/dev/null 2>&1 && [ -f "go.mod" ]; then
        echo "Updating Dota 2 replay parser (github.com/dotabuff/manta)..."
        # Only update the replay parser so we don't break Steam / GC client APIs
        go get -u github.com/dotabuff/manta
        go mod tidy
        if [ -d "vendor" ]; then
            echo "Updating vendor directory..."
            go mod vendor
        fi
        echo "Replay parser updated successfully."
    else
        echo "WARNING: Cannot update dependencies. Go is not installed or go.mod is missing."
    fi
fi

# Build function with automatic vendor sync recovery
build_binary() {
    local target="$1"
    local pkg="$2"
    local build_err

    if ! build_err=$(go build -o "$target" "$pkg" 2>&1); then
        if echo "$build_err" | grep -q "inconsistent vendoring"; then
            echo "Inconsistent vendoring detected in vendor/ directory."
            echo "Resyncing vendor directory with 'go mod vendor'..."
            go mod vendor
            go build -o "$target" "$pkg"
        else
            echo "$build_err" >&2
            return 1
        fi
    fi
}

# Auto-rebuild if source files exist and Go compiler is available
if command -v go >/dev/null 2>&1 && [ -f "go.mod" ] && [ -d "cmd" ]; then
    REBUILD=false

    # Rebuild if binaries are missing
    if [ ! -f "bot" ] || [ ! -f "server" ]; then
        REBUILD=true
    else
        # Rebuild if go.mod or any Go source file is newer than the binary
        if [ "go.mod" -nt "server" ]; then
            REBUILD=true
        else
            NEWEST_SRC=$(find . -maxdepth 3 -name "*.go" -newer "server" 2>/dev/null | head -n 1)
            if [ -n "$NEWEST_SRC" ]; then
                REBUILD=true
            fi
        fi
    fi

    if [ "$REBUILD" = true ]; then
        echo "Source changes or missing binaries detected. Rebuilding..."
        build_binary bot ./cmd/bot
        build_binary server ./cmd/server
        echo "Build completed successfully."
    fi
fi

# Verify binaries exist
if [ ! -f "bot" ]; then
    echo "ERROR: bot executable not found!"
    echo "Please ensure Go is installed to build it or extract the pre-built binary."
    exit 1
fi

if [ ! -f "server" ]; then
    echo "ERROR: server executable not found!"
    echo "Please ensure Go is installed to build it or extract the pre-built binary."
    exit 1
fi

# Free up ports if previous instances were left running
cleanup_stale_ports() {
    for port in 8081 8082; do
        PID=$(lsof -ti :$port 2>/dev/null || fuser $port/tcp 2>/dev/null || true)
        if [ -n "$PID" ]; then
            kill -9 $PID 2>/dev/null || true
        fi
    done
}
cleanup_stale_ports

# Process cleanup handler
cleanup() {
    echo ""
    echo "Stopping background services..."
    if [ -n "$BOT_PID" ]; then
        kill "$BOT_PID" 2>/dev/null || true
    fi
    echo "Done!"
}

trap cleanup EXIT INT TERM

echo "Starting Dota 2 Report Timestamp Tool..."
echo ""

export BOT_PORT=8082
./bot &
BOT_PID=$!

sleep 1

echo "Starting server on http://localhost:8081"
echo ""
echo "IMPORTANT: Keep this window open while using the tool!"
echo "Press Ctrl+C to stop the server."
echo ""

# Attempt to open browser in background
(sleep 1 && xdg-open http://localhost:8081 2>/dev/null || true) &

./server
