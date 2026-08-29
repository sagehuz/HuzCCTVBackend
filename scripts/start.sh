#!/usr/bin/env bash
set -eu
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="$ROOT_DIR/dist/huzbackend-linux-amd64"
if [ ! -f "$BIN" ]; then
  echo "Binary not found; building..."
  make -C "$ROOT_DIR" build-linux >/dev/null 2>&1 || true
fi
if [ ! -f "$BIN" ]; then
  echo "Build failed: $BIN missing" >&2
  exit 1
fi
# The CLI resolves .env relative to the binary directory → ensure one exists next to it.
if [ ! -f "$ROOT_DIR/dist/.env" ]; then
  if [ -f "$ROOT_DIR/.env" ]; then
    cp "$ROOT_DIR/.env" "$ROOT_DIR/dist/.env"
  elif [ -f "$ROOT_DIR/.env.example" ]; then
    cp "$ROOT_DIR/.env.example" "$ROOT_DIR/dist/.env"
  fi
fi
"$BIN" start
