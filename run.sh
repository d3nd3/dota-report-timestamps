#!/bin/bash

set -e

# Suppress protobuf registration conflicts
export GOLANG_PROTOBUF_REGISTRATION_CONFLICT=warn

# Support updating dependencies: ./run.sh --update or UPDATE=1 ./run.sh
if [ "$1" == "--update" ] || [ "$1" == "-u" ] || [ "$UPDATE" == "1" ]; then
    echo "Updating Dota 2 replay parser (github.com/dotabuff/manta)..."
    # Only update the replay parser to prevent breaking third-party Steam/GC APIs
    go get -u github.com/dotabuff/manta
    go mod tidy
    if [ -d "vendor" ]; then
        echo "Updating vendor directory..."
        go mod vendor
    fi
    echo "Replay parser updated."
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

echo "Cleaning old binaries..."
rm -f bot server

echo "Building bot..."
build_binary bot ./cmd/bot

echo "Building server..."
build_binary server ./cmd/server

echo "Build successful!"

# Function to kill background processes on exit
cleanup() {
    echo ""
    echo "Stopping background processes..."
    if [ -n "$BOT_PID" ]; then
        kill "$BOT_PID" 2>/dev/null || true
    fi
    echo "Done!"
}
trap cleanup EXIT INT TERM

echo "Starting bot on port 8082..."
export BOT_PORT=8082
./bot &
BOT_PID=$!

sleep 1

echo "Starting server on http://localhost:8081"
echo ""
./server
