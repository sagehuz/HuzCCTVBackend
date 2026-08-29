# Huz CCTV Server

Go backend for Huz CCTV that serves the embedded web app, exposes REST APIs, handles WebSocket signaling, and scans local LAN devices.

## Quick start

1. Copy `.env.example` to `.env` if needed.
2. Build binaries:
   - `make build-all`
3. Run:
   - `./dist/huzbackend-linux-amd64` or `./scripts/start.sh`
4. Open `http://localhost:3300`

## Environment variables

- `PORT` default `3300`
- `ADMIN_USERNAME` default `admin`
- `ADMIN_PASSWORD` default `onemilusd`
- `COOKIE_SECURE` default `false`
- `SESSION_PERSISTENT` default `true`
- `DB_PATH` default `data/app.db`

## Build targets

- `make build-linux`
- `make build-macos`
- `make build-windows`
- `make build-all`

## Notes

- Build uses `CGO_ENABLED=0` and is compatible with cross-platform targets.
- Static frontend is embedded into the binary via `go:embed`.
- OUI data is embedded from `internal/oui/oui.txt`.
