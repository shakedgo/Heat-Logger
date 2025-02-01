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

# Clean up and rebuild backend
echo "Cleaning up and rebuilding backend..."
cd ../backend
pkill -f "./tmp/main" || true
rm -f tmp/main
go build -o tmp/main cmd/server/main.go

# Install backend dependencies and start the server
echo "Starting backend server..."
go mod tidy
"$GOPATH/bin/air"

# Cleanup on script termination
trap 'kill $FRONTEND_PID' EXIT 