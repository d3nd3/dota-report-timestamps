#!/bin/bash

set -e

if [ ! -f "bot" ]; then
    echo "ERROR: bot executable not found!"
    echo "Please make sure you extracted all files from the zip."
    exit 1
fi

if [ ! -f "server" ]; then
    echo "ERROR: server executable not found!"
    echo "Please make sure you extracted all files from the zip."
    exit 1
fi

cleanup() {
    echo ""
    echo "Stopping bot..."
    if [ -n "$BOT_PID" ]; then
        kill $BOT_PID 2>/dev/null || true
    fi
    echo "Done!"
    exit 0
}

trap cleanup SIGINT SIGTERM

echo "Starting Dota 2 Report Timestamp Tool..."
echo ""

export BOT_PORT=8082
./bot &
BOT_PID=$!

sleep 2

echo "Starting server on http://localhost:8081"
echo ""
echo "IMPORTANT: Keep this window open while using the tool!"
echo "Press Ctrl+C to stop the server."
echo ""

xdg-open http://localhost:8081 2>/dev/null || true

./server

cleanup

