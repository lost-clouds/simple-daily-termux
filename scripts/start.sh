#!/bin/bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PIDFILE="$ROOT/simple-daily-termux.pid"
LOGFILE="$ROOT/simple-daily-termux.log"
CFGFILE="${1:-$ROOT/config.json}"

cd "$ROOT"

if [ -f "$PIDFILE" ]; then
    pid="$(cat "$PIDFILE")"
    if kill -0 "$pid" 2>/dev/null; then
        echo "Already running (pid=$pid)"
        exit 0
    fi
    rm "$PIDFILE"
fi

nohup "$ROOT/simple-daily-termux" "$CFGFILE" >> "$LOGFILE" 2>&1 &
echo $! > "$PIDFILE"
echo "Started simple-daily-termux (pid=$!)"
