#!/bin/bash

set -e

echo "Building server..."
go build -o server ./cmd/server

if [ $? -eq 0 ]; then
    echo "Build successful!"
    echo "Starting server on http://localhost:8081"
    echo ""
    ./server
else
    echo "Build failed!"
    exit 1
fi

