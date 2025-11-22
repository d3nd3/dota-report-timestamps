#!/bin/bash

set -e

# Suppress protobuf registration conflicts between manta and go-dota2 (if they were in same binary)
# Still useful to keep if any other deps conflict, though we split them.
export GOLANG_PROTOBUF_REGISTRATION_CONFLICT=warn

echo "Cleaning old binaries..."
rm -f bot server

echo "Building bot..."
go build -o bot ./cmd/bot

echo "Building server..."
go build -o server ./cmd/server

if [ $? -eq 0 ]; then
    echo "Build successful!"

    # Function to kill background processes on exit
    cleanup() {
        echo "Stopping processes..."
        if [ -n "$BOT_PID" ]; then
            kill $BOT_PID 2>/dev/null
        fi
    }
    trap cleanup EXIT SIGINT SIGTERM

    echo "Starting bot on port 8082..."
    export BOT_PORT=8082
    ./bot &
    BOT_PID=$!

    # Wait a sec for bot to be ready
    sleep 1

    echo "Starting server on http://localhost:8081"
    echo ""
    ./server
else
    echo "Build failed!"
    exit 1
fi
