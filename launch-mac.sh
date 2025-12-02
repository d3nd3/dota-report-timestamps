#!/bin/bash

set -e

if [ ! -f "bot" ] && [ ! -f "bot-arm64" ]; then
    echo "ERROR: bot executable not found!"
    echo "Please make sure you extracted all files from the zip."
    exit 1
fi

if [ ! -f "server" ] && [ ! -f "server-arm64" ]; then
    echo "ERROR: server executable not found!"
    echo "Please make sure you extracted all files from the zip."
    exit 1
fi

BOT_EXE="bot"
SERVER_EXE="server"

if [[ $(uname -m) == "arm64" ]] && [ -f "bot-arm64" ] && [ -f "server-arm64" ]; then
    BOT_EXE="bot-arm64"
    SERVER_EXE="server-arm64"
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
./"$BOT_EXE" &
BOT_PID=$!

sleep 2

echo "Starting server on http://localhost:8081"
echo ""
echo "IMPORTANT: Keep this window open while using the tool!"
echo "Press Ctrl+C to stop the server."
echo ""

open http://localhost:8081 2>/dev/null || true

./"$SERVER_EXE"

cleanup

