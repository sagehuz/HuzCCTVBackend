#!/usr/bin/env bash
set -eu
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"
if [ ! -f .env ]; then
  cp .env.example .env
fi
BIN="./dist/huzbackend-linux-amd64"
if [ ! -f "$BIN" ]; then
  echo "Binary not found; building..."
  make build-linux >/dev/null 2>&1 || true
fi
if [ ! -f "$BIN" ]; then
  echo "Build failed: $BIN missing" >&2
  exit 1
fi
PID_FILE=".huzbackend.pid"
if [ -f "$PID_FILE" ] && kill -0 "$(cat "$PID_FILE")" 2>/dev/null; then
  echo "Already running with PID $(cat "$PID_FILE")"
  exit 0
fi
nohup "$BIN" > .huzbackend.log 2>&1 &
echo $! > "$PID_FILE"
echo "Huz CCTV server started on http://127.0.0.1:$(grep '^PORT=' .env 2>/dev/null | cut -d= -f2- || echo 3300)"
