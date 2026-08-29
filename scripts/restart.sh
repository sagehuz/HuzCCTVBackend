#!/usr/bin/env bash
set -eu
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="$ROOT_DIR/dist/huzbackend-linux-amd64"
if [ ! -f "$BIN" ]; then
  echo "Binary not found: $BIN" >&2
  exit 1
fi
"$BIN" restart
