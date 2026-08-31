[![Buy Me a Coffee](https://img.shields.io/badge/Buy%20Me%20a%20Coffee-ffdd00?style=for-the-badge&logo=buy-me-a-coffee&logoColor=black)](https://www.buymeacoffee.com/khiemtrung)

# Huz CCTV Server

> 📱 **Android client:** the Huz CCTV Android app is available at
> **[HuzCCTV-Android](https://github.com/sagehuz/HuzCCTV-Android)**.

Go backend for Huz CCTV that serves the embedded web app, exposes REST APIs, handles WebSocket signaling, and scans local LAN devices.

The web dashboard is **multilingual** (English by default, Vietnamese available) with a language switcher in the header. The CLI is fully in English.

> 📖 **For end users:** a complete English user guide (installation, downloading from SourceForge.net, first sign-in, camera setup, CLI reference, troubleshooting…) is available in **[USER_GUIDE.md](USER_GUIDE.md)**.

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

## Frontend development (one source of truth)

The web frontend lives in **`public/`** at the repository root — this is the
single place to edit HTML/CSS/JS.

`cmd/huzbackend/public/` is a **generated mirror** of `public/`. It is what
`go:embed` actually packages into the binary (`cmd/huzbackend/main.go`), but it
is **never edited by hand**:

- `make build-*` automatically syncs `public/` → `cmd/huzbackend/public/`
  before compiling, so the binary always embeds the latest frontend.
- If you compile directly with `go build ./cmd/huzbackend` instead, run
  `make sync-public` first.
- `make check-public` verifies the two directories match, and `go test ./...`
  also fails if they drift (see `cmd/huzbackend/main_test.go`).

## Camera controls (zoom & PTZ)

The camera page (`/camera.html`) now ships with per-camera controls:

- **Digital zoom (client-side)** — works on any browser, no device support needed:
  - `＋` / `−` / reset buttons and the percentage badge in the top-right of each video.
  - Mouse wheel zoom, double-click to zoom (double-click again to reset).
  - Pinch-to-zoom and drag-to-pan on touch screens.
- **Device control bar** — below each video, wired to the camera device (app **HuzHome**):
  - **Device zoom** slider (sensor zoom up to `maxZoom`) + reset, and **EV** exposure slider + reset.
  - **Pan/tilt** D-pad (hold to keep moving) + home button to re-center the frame.
  - **Torch**, **switch camera**, **snapshot** (opens the JPEG), **motion detection** toggle, and **focus** reset.
  - **Tap the video** to focus at that point (single tap; double-tap is digital zoom).
  - **Battery** level + **motion** badge from live `device-status` telemetry.

The server relays these messages verbatim (`signal.go` already forwards the
`control` type). The viewer asks each device for its `capabilities` and uses
`device-status` / `snapshot` replies. Protocol used by the viewer:

```json
{ "type": "control", "targetId": "<device-client-id>", "action": "zoom", "value": 2.0 }
{ "type": "control", "targetId": "<device-client-id>", "action": "zoom-reset" }
{ "type": "control", "targetId": "<device-client-id>", "action": "torch", "value": true }
{ "type": "control", "targetId": "<device-client-id>", "action": "switch-camera" }
{ "type": "control", "targetId": "<device-client-id>", "action": "capabilities" }
{ "type": "control", "targetId": "<device-client-id>", "action": "snapshot" }
{ "type": "control", "targetId": "<device-client-id>", "action": "focus", "valueX": 0.5, "valueY": 0.5 }
{ "type": "control", "targetId": "<device-client-id>", "action": "focus-reset" }
{ "type": "control", "targetId": "<device-client-id>", "action": "exposure", "value": -2 }
{ "type": "control", "targetId": "<device-client-id>", "action": "pan", "valueX": 0.4, "valueY": 0.2 }
{ "type": "control", "targetId": "<device-client-id>", "action": "pan-reset" }
{ "type": "control", "targetId": "<device-client-id>", "action": "motion", "value": false }
```

The camera device (HuzHome app) handles all of these actions and replies with
`capabilities`, `device-status`, and `snapshot` messages.

## Phone remote — manage & control an Android phone (`/phone.html`)

The **Phone Remote** page (`/phone.html`) lets you select a connected Android
device, watch its **screen** (WebRTC, like the camera page) and **control it**
from the browser. All commands travel inside the existing `control` message, so
the server just relays them — no new message types.

### Workflow

1. The Android app registers with an optional `kind: "phone"` and `model`
   (the server now forwards `kind`/`model` in `device-list`).
2. Pick the device in the dropdown → the page sends `watch` and
   `control {action:"capabilities"}`.
3. Click **Start screen share** → the app captures the screen via
   **MediaProjection** (one-time consent dialog on the phone) and pushes the
   screen track over the existing WebRTC connection.
4. Control the phone: tap / drag / hold directly on the preview, or use the
   virtual buttons.

### Remote control protocol (implemented in the HuzHome app)

```jsonc
// Screen sharing
{ "type": "control", "targetId": "<device-client-id>", "action": "screen-start" }
{ "type": "control", "targetId": "<device-client-id>", "action": "screen-stop" }

// Turn the phone screen on / off (no root needed)
//   screen-on  → PowerManager.wakeUp()   (or FLAG_TURN_SCREEN_ON fallback)
//   screen-off → AccessibilityService GLOBAL_ACTION_LOCK_SCREEN
{ "type": "control", "targetId": "<device-client-id>", "action": "screen-on" }
{ "type": "control", "targetId": "<device-client-id>", "action": "screen-off" }

// Gestures — coordinates are normalized 0..1 relative to the phone screen.
// Android injects them with AccessibilityService.dispatchGesture().
{ "type": "control", "targetId": "<device-client-id>", "action": "tap", "valueX": 0.5, "valueY": 0.3 }
{ "type": "control", "targetId": "<device-client-id>", "action": "long-press", "valueX": 0.5, "valueY": 0.3, "valueDurationMs": 700 }
{ "type": "control", "targetId": "<device-client-id>", "action": "swipe",
  "valueX1": 0.5, "valueY1": 0.8, "valueX2": 0.5, "valueY2": 0.2, "valueDurationMs": 350 }

// Pinch — two fingers; (x1,y1)/(x2,y2) are the finger start points and
// (x3,y3)/(x4,y4) the finger end points (both strokes run for valueDurationMs).
{ "type": "control", "targetId": "<device-client-id>", "action": "pinch",
  "valueX1": 0.3, "valueY1": 0.4, "valueX2": 0.7, "valueY2": 0.4,
  "valueX3": 0.2, "valueY3": 0.4, "valueX4": 0.8, "valueY4": 0.4,
  "valueDurationMs": 400 }

// Hardware keys
{ "type": "control", "targetId": "<device-client-id>", "action": "key", "key": "back" }
// key: back | home | recent | menu | volume-up | volume-down
```

### Replies the app should send

- `capabilities` — reply to `control {action:"capabilities"}` with extended data:
  `{ "model": "...", "brand": "...", "screenSize": {"width":1080,"height":2400},
     "supportsTouch": true, "supportsScreenShare": true }`
- `device-status` — push periodically with `{ "screenOn": true, "screenStreaming": true,
  "batteryLevel": 78, "batteryCharging": true }`. The page uses it for the
  battery badge and the **"Screen is off"** warning, and to enable/disable the
  screen-on/off buttons.

> **Android side notes** — the phone does **not** need root. Screen capture uses
> MediaProjection (the user taps "Start" once), touch/swipe/pinch use an
> Accessibility Service with `canPerformGestures`. Share **one** MediaProjection
> screen track across all connected viewers. If the device has a PIN/pattern,
> `screen-on` only wakes the lock screen — keep the camera phone unlocked for a
> smooth remote experience.

