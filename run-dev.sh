#!/bin/bash

# Make the script executable
chmod +x run-dev.sh

# Set GOPATH if not set
export GOPATH="${GOPATH:-$HOME/go}"
export PATH="$PATH:$GOPATH/bin"

# Install air for Go hot-reloading if not installed
if ! command -v air &> /dev/null; then
    echo "Installing air for Go hot-reloading..."
    go install github.com/cosmtrek/air@latest
fi

# Install frontend dependencies
echo "Installing frontend dependencies..."
cd frontend && npm install

# Start the frontend server in the background
echo "Starting frontend server..."
npm run dev &
FRONTEND_PID=$!

# Register cleanup before blocking on air
trap 'kill $FRONTEND_PID 2>/dev/null' EXIT SIGINT SIGTERM

# Prepare backend
echo "Starting backend server..."
cd ../backend
pkill -f "./tmp/main" || true
rm -f tmp/main
go mod tidy
"$GOPATH/bin/air"