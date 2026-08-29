#!/usr/bin/env bash
set -eu
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"
"$ROOT_DIR/scripts/stop.sh" || true
"$ROOT_DIR/scripts/start.sh"
