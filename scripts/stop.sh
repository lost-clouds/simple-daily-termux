#!/bin/bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PIDFILE="$ROOT/simple-daily-termux.pid"
TIMEOUT=10

if [ ! -f "$PIDFILE" ]; then
    echo "Not running (no pidfile)"
    exit 0
fi

pid="$(cat "$PIDFILE")"

if ! kill -0 "$pid" 2>/dev/null; then
    echo "Process not running (stale pidfile)"
    rm "$PIDFILE"
    exit 0
fi

# Send SIGTERM for graceful shutdown
kill "$pid" 2>/dev/null || true

# Wait for process to exit
elapsed=0
while kill -0 "$pid" 2>/dev/null && [ "$elapsed" -lt "$TIMEOUT" ]; do
    sleep 1
    elapsed=$((elapsed + 1))
done

# Force kill if still running
if kill -0 "$pid" 2>/dev/null; then
    echo "Timed out waiting for graceful shutdown, sending SIGKILL"
    kill -9 "$pid" 2>/dev/null || true
    sleep 1
fi

rm -f "$PIDFILE"
echo "Stopped simple-daily-termux"
