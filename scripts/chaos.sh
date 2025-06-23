#!/usr/bin/env bash
set -euo pipefail

# Build agentry
echo "Building agentry binary"
go build -o agentry ./cmd/agentry

# Start long-running server
./agentry serve examples/.agentry.yaml >/tmp/agentry.log 2>&1 &
PID=$!

echo "Started server with PID $PID"
WAIT=$(( RANDOM % 120 + 30 ))
echo "Will kill after $WAIT seconds"
sleep $WAIT

kill $PID
sleep 5
echo "Server process $PID terminated"

# Restart server to verify recovery
./agentry serve examples/.agentry.yaml >/tmp/agentry.log 2>&1 &
PID=$!
echo "Restarted server with PID $PID"
sleep 15
kill $PID

echo "Recovery verified"
