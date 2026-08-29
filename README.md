# Huz CCTV Server

Go backend for Huz CCTV that serves the embedded web app, exposes REST APIs, handles WebSocket signaling, and scans local LAN devices.

The web dashboard is **multilingual** (English by default, Vietnamese available) with a language switcher in the header. The CLI is fully in English.

## Quick start

1. Build binaries: `make build-all`
2. Run `./dist/huzbackend` (no arguments) → **the terminal opens an interactive menu**: Start/Stop server, configuration, auto-start...

> To run the server in the background without the menu: `./dist/huzbackend start`; open the web dashboard: `./dist/huzbackend open`.

## Managing the server from the CLI

The `huzbackend` binary has a built-in command set that works on macOS / Windows / Linux. **Running it without arguments from a terminal opens the interactive menu** with full functionality:

| Command | Description |
|---|---|
| `huzbackend` | **Open the interactive menu** (when run from a terminal) |
| `huzbackend start` | Start the server in the background, writes `.huzbackend.pid` + `.huzbackend.log` |
| `huzbackend stop` | Stop the server (graceful, falls back to force kill if needed) |
| `huzbackend restart` | Stop, then start again |
| `huzbackend status` | Show status: running/stopped, PID, uptime, port |
| `huzbackend logs [-n <lines>] [-f]` | Show logs (default 50 lines; `-f` follows like `tail -f`) |
| `huzbackend open` | Open the web dashboard in the browser |
| `huzbackend autostart on\|off\|status` | Start automatically at login |
| `huzbackend config list\|get\|set\|reset` | Read/write `.env` configuration |
| `huzbackend menu` | Interactive menu mode |
| `huzbackend version` | Show version |

Examples:

```bash
./dist/huzbackend            # open the interactive menu
./dist/huzbackend start      # start the server in the background
./dist/huzbackend status
./dist/huzbackend config set PORT 3301
```

Note: `.env`, `data/`, `.huzbackend.pid`, `.huzbackend.log` are resolved relative to the **directory containing the binary** — keeping the binary together with `.env` and `data/` is recommended.

## Autostart at login

`huzbackend autostart on` creates a login task, depending on the operating system:

- **macOS**: writes `~/Library/LaunchAgents/com.huzcctv.server.plist` and loads it via `launchctl` (`autostart off` removes the task).
- **Windows**: adds a `Huz CCTV Server` value to the registry `HKCU\...\CurrentVersion\Run` (runs in the background without a console window).
- **Linux**: creates + enables a systemd **user** unit `~/.config/systemd/user/huzcctv.service`.

## Environment variables

- `PORT` default `3300`
- `ADMIN_USERNAME` default `admin`
- `ADMIN_PASSWORD` default `onemilusd`
- `COOKIE_SECURE` default `false`
- `SESSION_PERSISTENT` default `true`
- `DB_PATH` default `data/app.db`

They can be edited directly with `huzbackend config set <KEY> <VALUE>`.

## Web app language

- The web interface defaults to **English**. The first time you visit, it follows your browser language (`vi` → Vietnamese).
- Use the **language dropdown in the header** (English / Tiếng Việt) to switch. The choice is remembered in `localStorage`.
- Static text is translated through `data-i18n` attributes; dynamic strings use the `I18N` dictionary in `public/js/i18n.js`.

## Build targets

- `make build-linux`
- `make build-macos`
- `make build-windows`
- `make build-all`

## Notes

- Build uses `CGO_ENABLED=0` and is compatible with cross-platform targets.
- Static frontend is embedded into the binary via `go:embed`.
- OUI data is embedded from `internal/oui/oui.txt`.

