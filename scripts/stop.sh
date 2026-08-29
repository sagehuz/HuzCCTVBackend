#!/usr/bin/env bash
set -eu
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"
PID_FILE=".huzbackend.pid"
if [ ! -f "$PID_FILE" ]; then
  echo "No PID file found"
  exit 0
fi
PID="$(cat "$PID_FILE")"
if kill -0 "$PID" 2>/dev/null; then
  kill "$PID"
  echo "Stopped Huz CCTV server (PID $PID)"
else
  echo "Process not running; removing stale PID file"
fi
rm -f "$PID_FILE"
