# Huz CCTV — Complete User Guide

> **Turn your retired Android devices into a CCTV camera system.**
>
> **Huz CCTV Server** is a free, open-source, cross-platform backend for a DIY
> surveillance system. It turns old Android phones (running the companion
> **HuzHome** app) into network cameras and lets you watch them live from any
> modern web browser — all inside your own local network.

This guide explains everything you need: what the system is, how to download it
from **SourceForge.net**, how to install it on **Windows / macOS / Linux**, how to
sign in, how to use the web dashboard, how to control your cameras, how to manage
the server from the command line, and how to fix common problems.

---

## Table of contents

- [1. About Huz CCTV](#1-about-huz-cctv)
- [2. How it works](#2-how-it-works)
- [3. Features](#3-features)
- [4. System requirements](#4-system-requirements)
- [5. Downloading from SourceForge.net](#5-downloading-from-sourceforgenet)
- [6. Installation](#6-installation)
  - [6.1 Windows](#61-windows)
  - [6.2 macOS](#62-macos)
  - [6.3 Linux](#63-linux)
  - [6.4 What happens on first run](#64-what-happens-on-first-run)
- [7. First-time sign-in](#7-first-time-sign-in)
- [8. Using the web dashboard](#8-using-the-web-dashboard)
  - [8.1 Dashboard](#81-dashboard)
  - [8.2 Network Devices](#82-network-devices)
  - [8.3 Camera view](#83-camera-view)
  - [8.4 Changing the interface language](#84-changing-the-interface-language)
- [9. Setting up cameras (HuzHome Android app)](#9-setting-up-cameras-huzhome-android-app)
- [10. Camera controls reference](#10-camera-controls-reference)
- [11. Managing the server from the command line](#11-managing-the-server-from-the-command-line)
- [12. The interactive menu](#12-the-interactive-menu)
- [13. Configuration](#13-configuration)
- [14. Autostart at login](#14-autostart-at-login)
- [15. REST API reference](#15-rest-api-reference)
- [16. WebSocket signaling protocol](#16-websocket-signaling-protocol)
- [17. Building from source (optional)](#17-building-from-source-optional)
- [18. Updating to a newer version](#18-updating-to-a-newer-version)
- [19. Troubleshooting](#19-troubleshooting)
- [20. Frequently asked questions](#20-frequently-asked-questions)
- [21. Security notes](#21-security-notes)
- [22. License and support](#22-license-and-support)

---

## 1. About Huz CCTV

**Huz CCTV** is a DIY (do-it-yourself) CCTV system built on two components:

| Component | What it is |
|---|---|
| **Huz CCTV Server** | A small, single-file program written in **Go**. It runs on Windows, macOS, or Linux, serves the web dashboard, exposes a REST API, relays WebRTC signaling between cameras and viewers, and scans your local network to discover devices. |
| **HuzHome (Android app)** | The companion app (written in Kotlin) that you install on an old Android phone. It turns the phone's camera into a CCTV camera and connects it to the server. |

There is **no cloud involved**. Everything runs on your own machines and on your
own network.

The server was published as an open-source project on
[SourceForge.net](https://sourceforge.net/projects/huz-cctv/) under the
**MIT license**, and is free to download and use.

---

## 2. How it works

```
┌─────────────────────────┐          ┌──────────────────────────┐
│   Old Android phone     │          │   Your computer / server │
│   (camera device)       │          │   (Huz CCTV Server)      │
│   runs the HuzHome app  │          │                          │
│                         │  WebRTC  │  ┌────────────────────┐  │
│   ┌─────────────────┐   │  video   │  │  Web dashboard     │  │
│   │  camera + mic   │◄──┼──────────┼──┤  (browser)         │  │
│   └─────────────────┘   │  peer-to-│  │                    │  │
│                         │  peer    │  │  ┌──────────────┐  │  │
│   registers via         │          │  │  │  viewer      │  │  │
│   WebSocket             │          │  │  └──────────────┘  │  │
│   └────────────┴────────┼──►┌──────┴──┴────────────────────┘  │
│                         │   │  Signaling server  (relays only)│
│                         │   └─────────────────────────────────┘
└─────────────────────────┘          └──────────────────────────┘
```

- The **camera device** (HuzHome app) connects to the server over a WebSocket and
  "registers" itself.
- The **viewer** (your browser, signed in to the dashboard) connects to the same
  server and asks for the list of cameras.
- The server only **relays signaling messages** (WebRTC *offer*, *answer*, *ICE
  candidates*, and *control* commands) between the two. It **never stores or
  relays the video itself**.
- The live video stream travels **peer-to-peer** directly between the camera phone
  and your browser.

This design means the server stays lightweight and your video never passes
through a third-party service.

---

## 3. Features

- **Local web server** — a built-in, dark-themed web dashboard for signing in,
  watching live cameras, and browsing your network devices.
- **No runtime required** — the server is a single self-contained executable
  (pure Go, `CGO_ENABLED=0`). No Node.js, no Python, no Docker, and no separate
  database server are needed.
- **Cross-platform** — runs on **Windows**, **macOS** (Intel & Apple Silicon),
  and **Linux** (x86_64 & ARM64).
- **WebRTC peer-to-peer video** — live video flows directly between the camera
  phone and your browser; the server never stores video.
- **LAN device discovery** — scans your local network (ping sweep + ARP table),
  then enriches each device with the **vendor** (OUI lookup) and **hostname**
  (reverse DNS).
- **Built-in authentication** — SQLite database, password hashing with **scrypt**,
  httpOnly session cookies, and brute-force protection (rate limiting).
- **Multilingual web app** — **English** by default, **Vietnamese** available,
  switchable from the header.
- **Per-camera controls** — digital zoom, pan/tilt (PTZ), device sensor zoom,
  exposure, torch, snapshot, motion-detection toggle, and tap-to-focus.
- **Management CLI** — a full command set built into the same executable
  (`start`, `stop`, `status`, `logs`, `open`, `autostart`, `config`, …).
- **Autostart at login** — register the server to start automatically when you
  log in (macOS, Windows, and Linux).
- **Open source (MIT license)** — free to use, modify, and distribute.

---

## 4. System requirements

**Server side (runs Huz CCTV Server):**

| Requirement | Minimum |
|---|---|
| Operating system | Windows 10/11, macOS 11+, or Linux (kernel 4.x+) |
| Architecture | Windows: x86_64 · macOS: Intel (amd64) or Apple Silicon (arm64) · Linux: x86_64 or ARM64 |
| RAM | 256 MB free (1 GB recommended) |
| Disk space | ~50 MB free |
| Network | Any network interface with a private IPv4 address (e.g. `192.168.x.x`) |

> The server is tiny and happily runs 24/7 on an old laptop, a Raspberry Pi,
> a mini PC, or your main desktop.

**Viewer side (opens the web dashboard):**

- Any modern browser: **Chrome, Edge, Firefox, Safari** (latest version).
- WebRTC support is required for live video.
- A screen wide enough for the dashboard (desktop or tablet recommended).

**Camera side:**

- An Android phone (the older and less capable, the better — that is the point!)
- The **HuzHome** app installed on that phone.
- The phone must be on the **same local network** as the server (or reachable
  through a VPN / port forwarding if you use it remotely).

---

## 5. Downloading from SourceForge.net

The official download source for Huz CCTV is **SourceForge.net**.

### 5.1 Open the project page

1. Go to the project page:

   👉 **https://sourceforge.net/projects/huz-cctv/**

   The page is titled **"Huz CCTV — Make your retired android devices to be CCTV
   camera"** and is published by the project author `sagehuz`.

2. Click the **"Download"** button (or open the **Files** tab directly at
   **https://sourceforge.net/projects/huz-cctv/files/**).

> **Tip:** SourceForge picks the file it thinks is best for your platform
> automatically. You can still choose a different file manually from the list.

### 5.2 Choose the right file for your platform

The server is distributed as a **zip archive** containing the pre-built
executable (plus a sample `.env.example` file). Pick the file that matches your
computer:

| Operating system | Architecture | File to download |
|---|---|---|
| **Windows** | 64-bit (x86_64) | `huzcctv-windows-amd64.zip` → contains `huzbackend.exe` |
| **macOS** | Intel (x86_64) | `huzcctv-macos-amd64.zip` → contains `huzbackend-darwin-amd64` |
| **macOS** | Apple Silicon (M1/M2/M3…) | `huzcctv-macos-arm64.zip` → contains `huzbackend-darwin-arm64` |
| **Linux** | 64-bit (x86_64) | `huzcctv-linux-amd64.zip` → contains `huzbackend-linux-amd64` |
| **Linux** | ARM64 (e.g. Raspberry Pi 4/5, 64-bit) | `huzcctv-linux-arm64.zip` → contains `huzbackend-linux-arm64` |

> **Not sure which architecture you have?**
>
> - **Windows:** Open *Settings → System → About*. Look at "System type" — it
>   should say *64-bit operating system, x64-based processor*.
> - **macOS:** Click the Apple menu → *About This Mac*. If the chip is listed as
>   "Apple M1/M2/M3/…", download the **arm64** build; if it says "Intel", download
>   the **amd64** build.
> - **Linux:** Run `uname -m` in a terminal — `x86_64` → **amd64**;
>   `aarch64` → **arm64**.

### 5.3 After you download

The zip contains a folder with everything the server needs. You can now go to
[Section 6 — Installation](#6-installation).

> **Note for early adopters:** if the Files page does not show any release
> archives yet (the project is new), you have two options:
> 1. **Watch** the project on SourceForge ("Get an email when there's a new
>    version of Huz CCTV") so you are notified when the first release is
>    published, or
> 2. Build the server yourself from the source code — see
>    [Section 17 — Building from source](#17-building-from-source-optional).

---

## 6. Installation

The server needs no installer. "Installation" simply means **unzipping the
archive to a folder of your choice** and **running the executable**. Because the
binary resolves its data files (`.env`, `data/`, logs, PID) relative to its own
folder, it is recommended to keep everything together in one folder.

> **Recommended layout:**
>
> ```
> C:\HuzCCTV\          (Windows)   or   ~/huzcctv/     (macOS / Linux)
>   ├── huzbackend.exe / huzbackend-darwin-arm64 / ...
>   ├── .env.example                (sample configuration)
>   ├── .env                        (created automatically on first run)
>   ├── data\                       (SQLite database, created on first run)
>   └── .huzbackend.log             (server log, created on first run)
> ```

### 6.1 Windows

1. **Unzip** the downloaded archive (right-click → *Extract All…*).
2. Move the extracted folder somewhere permanent, for example
   `C:\HuzCCTV\`. Avoid system-protected folders like `C:\Program Files\`
   unless you are comfortable granting write permissions.
3. Open **Command Prompt** (press `Win+R`, type `cmd`, press `Enter`).
4. Navigate to the folder and start the server:
   ```bat
   cd C:\HuzCCTV
   huzbackend.exe start
   ```
   *(Windows SmartScreen may show a warning because the executable is not
   digitally signed — click **More info → Run anyway**.)*
5. Open the dashboard:
   ```bat
   huzbackend.exe open
   ```
   Or simply open your browser and go to `http://127.0.0.1:3300`.

> **Double-click option:** you can also double-click `huzbackend.exe`. If you run
> it from a terminal, the interactive menu opens (see
> [Section 12](#12-the-interactive-menu)); if you double-click it outside a
> terminal, the server starts directly in the background.

### 6.2 macOS

1. **Unzip** the downloaded archive (double-click the zip in Finder).
2. Move the extracted folder somewhere permanent, for example `~/huzcctv/`.
3. Open **Terminal** (`Cmd+Space`, type "Terminal", press `Enter`).
4. Navigate to the folder and make the binary executable:
   ```bash
   cd ~/huzcctv
   chmod +x huzbackend-darwin-*
   ```
5. Start the server:
   ```bash
   ./huzbackend-darwin-arm64 start          # Apple Silicon
   # or
   ./huzbackend-darwin-amd64 start          # Intel
   ```
6. Open the dashboard:
   ```bash
   ./huzbackend-darwin-arm64 open
   ```
   Or visit `http://127.0.0.1:3300` in your browser.

> **Gatekeeper note:** the first time you run the binary, macOS may say the app
> cannot be opened because it is from an unidentified developer. To allow it:
> **System Settings → Privacy & Security → scroll down → "Open Anyway"**, then
> confirm. Alternatively run `xattr -dr com.apple.quarantine huzbackend-darwin-*`
> in the terminal.

### 6.3 Linux

1. **Unzip** the downloaded archive:
   ```bash
   cd ~
   unzip huzcctv-linux-amd64.zip -d huzcctv        # x86_64
   # or
   unzip huzcctv-linux-arm64.zip -d huzcctv        # Raspberry Pi (64-bit)
   ```
2. Make the binary executable and start it:
   ```bash
   cd ~/huzcctv
   chmod +x huzbackend-linux-*
   ./huzbackend-linux-amd64 start
   ```
3. Open the dashboard:
   ```bash
   ./huzbackend-linux-amd64 open
   ```
   Or visit `http://127.0.0.1:3300` in your browser.

> The server listens on `0.0.0.0`, so once it is running you can also open the
> dashboard from **another computer** on the same network using the server's LAN
> IP address — for example `http://192.168.1.50:3300`. You can find your LAN IP
> with `./huzbackend-linux-amd64 status` or from the **Dashboard** page.

### 6.4 What happens on first run

When you start the server for the first time, it automatically:

1. Creates a **`.env`** configuration file next to the binary (if missing).
2. Creates a **`data/`** folder and the **SQLite database** (`data/app.db`).
3. **Creates the default administrator account** with the username and password
   from the configuration.
4. Writes a **`.huzbackend.log`** log file and a **`.huzbackend.pid`** PID file.

The default administrator credentials are:

| Setting | Default value |
|---|---|
| Username | `admin` |
| Password | `onemilusd` |

> ⚠️ **Change the default password immediately after your first sign-in** — see
> [Section 7](#7-first-time-sign-in). Anyone on your network can otherwise sign in
> with the default credentials.

---

## 7. First-time sign-in

1. Make sure the server is running (see [Section 6](#6-installation)).
2. Open `http://127.0.0.1:3300` in your browser — or run
   `huzbackend open` / `huzbackend.exe open` from a terminal.
3. If you are not signed in yet, you are redirected to the **Sign in** page.
4. Enter the default credentials:
   - **Username:** `admin`
   - **Password:** `onemilusd`
5. Tick **"Remember me on this browser"** if you want a long-lived session
   (this is on by default).
6. Click **Sign in**. You land on the **Dashboard**.

### 7.1 Change the default password (strongly recommended)

1. While signed in, open the **Dashboard** and find the **Account / Change
   password** section (or use the `huzbackend config set ADMIN_PASSWORD` command,
   see [Section 13](#13-configuration)).
2. Enter your current password, choose a new password of **at least 8
   characters**, and save.
3. When the password changes, all other sessions are signed out — sign in again
   with the new password.

> 🔒 **Security reminder:** the server binds to `0.0.0.0`, so the sign-in page is
> reachable from any device on your network. Always use a strong, unique password.
> For internet exposure, see [Section 21 — Security notes](#21-security-notes).

---

## 8. Using the web dashboard

The web dashboard has four pages, available from the navigation bar at the top:

| Page | URL | Purpose |
|---|---|---|
| **Dashboard** | `/index.html` | Server status and quick access |
| **Network Devices** | `/devices.html` | Devices discovered on your LAN |
| **Camera** | `/camera.html` | Live camera views and controls |
| **Sign in** | `/login.html` | Log-in page |

### 8.1 Dashboard

The Dashboard shows a snapshot of your system:

- **Server** card — running/stopped status, IP address, port, hostname, uptime,
  operating system, app version, Go/CPU info, and the time the server started.
- **Network devices** card — the number of devices discovered on your local
  network, updated every time the page loads.
- **Actions** — buttons to open **Network Devices** and **View cameras**.

### 8.2 Network Devices

This page shows the result of a **local network scan**:

- The server performs a **ping sweep** of the current subnet (up to 1024 hosts),
- reads the **ARP table** (`ip neighbor show`, with a fallback to `arp -a`),
- and enriches each entry with the **vendor** (OUI lookup) and **hostname**
  (reverse DNS lookup).

Each row shows the device's **IP address, MAC address, vendor, hostname,
interface, and state** (Online / Stale / Offline).

- Use the **"Scan again"** button to re-run the scan.
- The scan result is informative — it helps you identify which of your devices
  are active on the network (including your camera phones).

### 8.3 Camera view

The Camera page lists every **HuzHome device** currently registered with the
server. Each camera appears as its own card containing:

- a live **video player**,
- a **digital zoom** control (top-right of the video),
- a **PTZ pad** (bottom-right of the video),
- a **device control bar** below the video,
- a **battery** level and **motion** badge from live device telemetry.

For a full description of every control, see
[Section 10 — Camera controls reference](#10-camera-controls-reference).

### 8.4 Changing the interface language

The web interface is **multilingual**:

- It defaults to **English**. The first time you visit, it follows your
  browser's language (`vi` → Vietnamese).
- Use the **language dropdown in the header** (*English* / *Tiếng Việt*) to
  switch at any time.
- Your choice is remembered in your browser's `localStorage`, so it persists
  between visits.

---

## 9. Setting up cameras (HuzHome Android app)

The camera "device" is an Android phone running the **HuzHome** app. This is the
companion app to Huz CCTV Server.

### 9.1 Prepare the phone

1. Take an old Android phone.
2. Connect it to your **Wi-Fi network** — the same network the server is on.
3. Keep it **plugged into power** and keep the screen on (you can configure the
   screen timeout to "never" in the phone settings).
4. Install the **HuzHome** app on the phone (APK). The phone does not need to be
   rooted.

### 9.2 Connect the phone to the server

1. Open **HuzHome** on the phone.
2. Enter the **server address** — the server's LAN IP and port, for example:
   `ws://192.168.1.50:3300/ws/signal` (the HuzHome app normally asks only for
   `192.168.1.50` and the port `3300`).
3. The app registers itself with the server over a WebSocket. When registration
   succeeds, the server replies with a `registered` message containing the
   device's client ID.
4. If the phone app connects twice with the same device ID, the server closes the
   older connection and keeps the newest one (it sends a `replaced` error to the
   old connection).

### 9.3 Watch the camera

1. Open the **Camera** page in the web dashboard
   (`http://127.0.0.1:3300/camera.html`).
2. The camera appears in the list automatically the moment HuzHome registers.
3. Your browser and the phone now negotiate a **WebRTC** connection (offer /
   answer / ICE candidates relayed by the server). Once connected, live video
   appears and flows **peer-to-peer** — the server does not carry the video.

### 9.4 Multiple cameras

- Each phone running HuzHome is one camera.
- You can add **as many cameras as you want** — the Camera page shows them all,
  one per card.
- Use the **"switch camera"** control if a phone has more than one camera (front
  and rear).

---

## 10. Camera controls reference

Each camera card provides two layers of controls:

### 10.1 Digital zoom (client-side, works on any browser)

This zoom is applied in your browser — it does not require any device support.

| Control | How to use |
|---|---|
| **＋ / − / reset buttons** | Top-right of the video, with a percentage badge showing the zoom level |
| **Mouse wheel** | Scroll over the video to zoom in/out |
| **Double-click** | Zoom in at the cursor; double-click again to reset |
| **Touch pinch** | Pinch to zoom, drag to pan (on touch screens) |

### 10.2 Device control bar (below the video)

These controls are sent to the camera device (HuzHome app) and require the device
to support the action.

| Control | What it does |
|---|---|
| **Device zoom slider** | Sensor zoom up to the device's `maxZoom`; the **reset** button returns to 1× |
| **EV slider** | Exposure compensation (e.g. −2 … +2); **reset** restores auto exposure |
| **Pan/Tilt D-pad** | Hold a direction to keep moving the camera; the **home** button re-centers the frame |
| **Torch** | Toggle the phone's flashlight on/off |
| **Switch camera** | Toggle between the phone's front and rear cameras |
| **Snapshot** | Capture the current frame; the JPEG opens in a new tab |
| **Motion detection** | Toggle motion detection on/off |
| **Focus reset** | Restore automatic focus |
| **Tap to focus** | Tap the video once to focus at that point (a *double*-tap is digital zoom instead) |

### 10.3 Live telemetry

- **Battery** badge — the phone's remaining battery level.
- **Motion** badge — flashes when the device reports motion events.

These come from live `device-status` messages sent by the device over the
signaling WebSocket.

---

## 11. Managing the server from the command line

You do **not** need to memorise commands. The binary you downloaded from
SourceForge contains a built-in **interactive menu**: just run the file in a
terminal and choose what you want to do by typing a number (**1** to **9**) and
pressing **Enter**.

### 11.1 Run the binary you downloaded

Put the binary in a folder of your choice (for example `C:\HuzCCTV\` on Windows
or `~/huzcctv/` on macOS/Linux), open a terminal in that folder and run the file
**with no arguments**:

**Windows:**
```bat
cd C:\HuzCCTV
huzbackend.exe
```

**macOS (Apple Silicon or Intel):**
```bash
cd ~/huzcctv
./huzbackend-darwin-arm64       # Apple Silicon (M1/M2/M3…)
# or
./huzbackend-darwin-amd64       # Intel
```

**Linux (x86_64 or ARM64):**
```bash
cd ~/huzcctv
./huzbackend-linux-amd64        # x86_64
# or
./huzbackend-linux-arm64        # ARM64 (Raspberry Pi 64-bit)
```

The **interactive menu** appears right away, showing the current status and the
configured port at the top:

```
════════════════════════════════════════════════
   Huz CCTV Server 1.0.0 — management
   Status: 🔴 STOPPED | Port: 3300
════════════════════════════════════════════════
  1) Start server
  2) Stop server
  3) Restart server
  4) Detailed status
  5) View logs (last 50 lines)
  6) Open web dashboard
  7) Autostart at login
  8) Configuration (.env)
  9) Quit
Choose [1-9, q]:
```

Type the number of the action you want and press **Enter** — the menu does the
rest for you.

### 11.2 What each menu option does

| Menu option | What it does | Typical use |
|---|---|---|
| **1 — Start server** | Starts the server in the background (writes `.huzbackend.pid` and `.huzbackend.log`) | The first thing you do after installing |
| **2 — Stop server** | Stops the server cleanly | Before updating or turning the machine off |
| **3 — Restart server** | Stops, then starts it again | After changing the configuration (`.env`) |
| **4 — Detailed status** | Shows running/stopped, PID, uptime, port, dashboard URL | "Is my server up?" |
| **5 — View logs** | Shows the last 50 lines of `.huzbackend.log` | See what the server is doing / diagnose problems |
| **6 — Open web dashboard** | Opens `http://127.0.0.1:3300` in your browser | Go to the web app |
| **7 — Autostart at login** | Opens a sub-menu to enable/disable starting at login (see [Section 12](#12-the-interactive-menu)) | Run the server automatically at login |
| **8 — Configuration (.env)** | Opens a sub-menu to view/change settings such as the port and admin password (see [Section 12](#12-the-interactive-menu)) | Change the port, password, … |
| **9 — Quit** | Leaves the menu (a running server keeps running in the background) | Done for now |

### 11.3 Example session

```
Choose [1-9, q]: 1
(server starts in the background…)

Choose [1-9, q]: 4
Status: 🟢 RUNNING
PID: 1234
Uptime: 1h 02m 11s
Port: 3300
Dashboard: http://127.0.0.1:3300
Log: /Users/you/huzcctv/.huzbackend.log

Choose [1-9, q]: 6
Opening: http://127.0.0.1:3300

Choose [1-9, q]: 9
Goodbye!
```

> 💡 **Tip:** closing the menu does **not** stop the server — it keeps running in
> the background. Run the binary again whenever you need to manage it; the menu
> re-opens and shows the live status at the top.

### 11.4 Direct commands (optional, for advanced users)

Every menu option is also available as a one-line command — handy for scripts or
when you already know exactly what you want:

| Command | Description |
|---|---|
| `huzbackend` | Open the interactive menu (when run from a terminal) |
| `huzbackend start` | Start the server in the background |
| `huzbackend stop` | Stop the server |
| `huzbackend restart` | Stop, then start again |
| `huzbackend status` | Show status: running/stopped, PID, uptime, port |
| `huzbackend logs [-n <lines>] [-f]` | Show logs (default 50 lines; `-f` follows like `tail -f`) |
| `huzbackend open` | Open the web dashboard in your browser |
| `huzbackend autostart on\|off\|status` | Enable / disable / check autostart at login |
| `huzbackend config list\|get\|set\|reset` | Read/write the `.env` configuration |
| `huzbackend menu` | Open the interactive menu mode |
| `huzbackend version` | Show the server version |
| `huzbackend help` | Show the built-in help text |

Examples:

```bash
./huzbackend status                  # is it running? on which port?
./huzbackend logs -n 100             # show the last 100 log lines
./huzbackend config set PORT 3301    # change the port
./huzbackend autostart on            # start automatically at login
```

> **Windows** examples use `huzbackend.exe` instead: `huzbackend.exe status`, etc.

### 11.5 Where files live

`.env`, `data/`, `.huzbackend.pid` and `.huzbackend.log` are resolved relative to
the **directory that contains the binary** — not your current directory. Keep the
binary together with its `.env` and `data/` folder, and you can run the commands
from anywhere.

---

## 12. The interactive menu

The main menu is covered in [Section 11](#11-managing-the-server-from-the-command-line).
Two of its options open their own sub-menus, described below.

### 12.1 Autostart sub-menu (menu option 7)

```
  ── Autostart at login (current: DISABLED) ──
  1) Enable
  2) Disable
  3) Back
Choose [1-3]:
```

- **1 — Enable:** the server will start automatically every time you log in.
- **2 — Disable:** removes the autostart task.
- **3 — Back:** return to the main menu.

How autostart is implemented on each operating system (macOS LaunchAgent,
Windows registry, Linux systemd) is described in
[Section 14 — Autostart at login](#14-autostart-at-login).

### 12.2 Configuration sub-menu (menu option 8)

```
  ── Configuration (.env) ──
  1) Show current configuration
  2) Change PORT
  3) Change ADMIN_USERNAME
  4) Change ADMIN_PASSWORD
  5) Change DB_PATH
  6) Change COOKIE_SECURE (true/false)
  7) Change SESSION_PERSISTENT (true/false)
  8) Restore default configuration
  9) Back
Choose [1-9]:
```

- **1 — Show current configuration:** prints every setting (the admin password is
  masked).
- **2 – 7 — Change …:** prompts you for a new value and writes it to `.env`.
- **8 — Restore default configuration:** resets `.env` back to the defaults.
- **9 — Back:** return to the main menu.

> If the server is running, this sub-menu reminds you that configuration changes
> only take effect after a **Restart** (main-menu option **3**). The full list of
> settings is documented in [Section 13 — Configuration](#13-configuration).

---

## 13. Configuration

Configuration is stored in a plain **`.env`** file next to the binary. The server
runs with sensible defaults even if the file does not exist.

### 13.1 Configuration keys

| Key | Default | Description |
|---|---|---|
| `PORT` | `3300` | HTTP/WebSocket port the server binds to (`0.0.0.0:<PORT>`) |
| `ADMIN_USERNAME` | `admin` | Username of the auto-created administrator account |
| `ADMIN_PASSWORD` | `onemilusd` | Password of the auto-created administrator account |
| `COOKIE_SECURE` | `false` | Set to `true` when the server is served over HTTPS (Secure cookie flag) |
| `SESSION_PERSISTENT` | `true` | When `true`, sessions stay valid for a long time; when `false`, sessions expire after 7 days |
| `DB_PATH` | `data/app.db` | SQLite database path (relative to the binary folder) |

> The `.env` values are overridden by real **environment variables** of the same
> name if they are set (e.g. `PORT=8080 ./huzbackend start`).

### 13.2 Editing configuration from the CLI

The easiest way to change settings is the `config` command:

```bash
huzbackend config list                 # show all settings (password is masked)
huzbackend config get PORT             # show one value
huzbackend config set PORT 3301        # change the port
huzbackend config set ADMIN_PASSWORD 'My-Str0ng-Pass!'
huzbackend config reset                # restore the default configuration
```

> After changing configuration, **restart the server** for it to take effect
> (`huzbackend restart`).

### 13.3 Editing `.env` manually

You can also open `.env` in any text editor. It looks like this:

```ini
PORT=3300
ADMIN_USERNAME=admin
ADMIN_PASSWORD=onemilusd
COOKIE_SECURE=false
SESSION_PERSISTENT=true
DB_PATH=data/app.db
```

Save the file and restart the server.

> Changing `ADMIN_USERNAME`/`ADMIN_PASSWORD` in `.env` only affects the account
> **created on the first run**. To change the password of an existing account,
> use the **Change password** feature in the web app (or delete the database once
> — after backing it up — to re-seed the default admin).

---

## 14. Autostart at login

You can make the server start automatically every time you log in to your
computer:

```bash
huzbackend autostart on       # enable
huzbackend autostart off      # disable
huzbackend autostart status   # check
```

What happens behind the scenes depends on your operating system:

| OS | Mechanism |
|---|---|
| **macOS** | Writes `~/Library/LaunchAgents/com.huzcctv.server.plist` and loads it with `launchctl`. `autostart off` removes it. |
| **Windows** | Adds a `Huz CCTV Server` entry to the registry key `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`, so it starts in the background **without a console window**. |
| **Linux** | Creates and enables a **systemd user unit** at `~/.config/systemd/user/huzcctv.service` (runs `huzbackend serve`, restarts on failure). |

> The autostart task starts the server **in the background** (`serve` mode).
> Autostart is registered per user account.

---

## 15. REST API reference

The server exposes a small REST API. JSON is used everywhere.

| Method | Endpoint | Auth | Description |
|---|---|---|---|
| `GET` | `/api/health` | No | Health check — returns `{"code":"ok","message":"Server is running successfully"}` |
| `GET` | `/api/version` | No | Server version + OS/arch (`{"version":"1.0.0","os":"linux","arch":"amd64"}`) |
| `POST` | `/api/auth/login` | No | Sign in — body `{username, password, remember}`; sets the `huz_session` cookie |
| `POST` | `/api/auth/logout` | No | Sign out (invalidates the session) |
| `GET` | `/api/auth/me` | Yes | Current user `{id, username}` |
| `POST` | `/api/auth/change-password` | Yes | Change password (requires current password; new password ≥ 8 chars) |
| `GET` | `/api/network-devices` | Yes | LAN scan result `{count, devices:[{ip, mac, hostname, iface, state, vendor}]}` |
| `GET` | `/api/server-info` | Yes | Server info: hostname, IPs, port, uptime, OS, arch, Go version, CPU count, version |
| `WS` | `/ws/signal` | Viewers must be signed in | WebRTC signaling + device registration (see [Section 16](#16-websocket-signaling-protocol)) |

**Authentication** uses an **httpOnly session cookie** named `huz_session`.
Pages `index.html`, `camera.html`, `devices.html` require sign-in and redirect to
`/login.html?next=<page>` when the session is missing or expired. Signed-in users
visiting `/login.html` are redirected to the dashboard.

**Security behavior:**
- Passwords are hashed with **scrypt** and stored in SQLite.
- Failed logins are **rate-limited**: after too many failures for an
  IP+username pair, further attempts are blocked for **15 minutes**.
- Sessions expire after **7 days**, unless `SESSION_PERSISTENT=true` (the
  default), in which case they stay valid for a long time.

---

## 16. WebSocket signaling protocol

Endpoint: `ws://<host>:<port>/ws/signal` (or `wss://` behind HTTPS).

This endpoint **relays JSON messages only**. The server never interprets SDP or
ICE payloads — it just forwards them to the right peer, adding a `from` field.

### 16.1 Message types

| Type | Direction | Purpose |
|---|---|---|
| `register` | device & viewer → server | Announce your role: `{"type":"register","role":"device"}` or `{"type":"register","role":"viewer"}` |
| `registered` | server → client | Reply containing your client `id` |
| `device-list` | server → viewer | The list of registered devices `{type, devices:[{id, name, deviceId}]}` |
| `watch` | viewer → device | Ask a device to start streaming |
| `offer` / `answer` / `ice-candidate` | between peers | WebRTC signaling (relayed) |
| `control` | viewer → device | Device commands (zoom, torch, PTZ, snapshot, …) |
| `capabilities` | either direction | Query/report device capabilities |
| `device-status` | device → viewer | Live telemetry (battery, motion, …) |
| `snapshot` | device → viewer | A still image |
| `error` | server → client | Errors, e.g. `target_gone`, `unauthorized` |

### 16.2 Rules enforced by the server

- **Viewers must be signed in.** A viewer without a valid `huz_session` cookie is
  rejected (close code `4001 Unauthorized`).
- **Devices do not need a session** — they register with `role: "device"`.
- **Duplicate device IDs** — a device reconnecting with the same `deviceId` kicks
  the old connection (close code `4002 replaced`).
- **Heartbeat** — the server pings every 10 seconds and drops connections that
  stop responding, so stale devices disappear from the list.
- All `control` messages are forwarded **verbatim** to the target device, which
  executes the action and replies with `capabilities`, `device-status` or
  `snapshot` messages.

### 16.3 Example `control` messages (used by the Camera page)

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

---

## 17. Building from source (optional)

If you prefer, or if no pre-built release is available yet, you can build the
server yourself. You need **Go 1.22 or newer** installed on your machine
(https://go.dev/dl/).

### 17.1 Get the source

```bash
git clone https://github.com/sagehuz/HuzCCTVBackend.git
cd HuzCCTVBackend
```

*(The SourceForge project page also links to the source code under its
**Code** tab — https://sourceforge.net/projects/huz-cctv/.)*

### 17.2 Build

The included `Makefile` builds for every platform:

```bash
make build-all        # Linux (amd64 + arm64), macOS (amd64 + arm64), Windows (amd64)
make build-linux      # Linux only
make build-macos      # macOS only
make build-windows    # Windows only
```

The binaries are written to `dist/`:

| File | Platform |
|---|---|
| `dist/huzbackend-linux-amd64` | Linux x86_64 |
| `dist/huzbackend-linux-arm64` | Linux ARM64 |
| `dist/huzbackend-darwin-amd64` | macOS Intel |
| `dist/huzbackend-darwin-arm64` | macOS Apple Silicon |
| `dist/huzbackend.exe` | Windows x86_64 |

All builds use `CGO_ENABLED=0`, so they are pure-Go, statically linked binaries
with no runtime dependencies — exactly like the official downloads.

### 17.3 Quality checks

```bash
make test     # run unit tests
make vet      # static analysis
make fmt      # format the code
```

### 17.4 Helper scripts

The repository also includes convenience scripts that build (if needed), create a
`.env`, and start/stop/restart the server:

- macOS/Linux: `scripts/start.sh`, `scripts/stop.sh`, `scripts/restart.sh`
- Windows: `scripts/windows/start.bat`, `scripts/windows/stop.bat`,
  `scripts/windows/restart.bat`

---

## 18. Updating to a newer version

1. Stop the server:
   ```bash
   huzbackend stop
   ```
2. Download the **new release** from SourceForge
   (https://sourceforge.net/projects/huz-cctv/files/) for your platform.
3. Replace the old executable with the new one — **keep** the `.env` file and the
   `data/` folder where they are.
4. Start the server again:
   ```bash
   huzbackend start
   ```
5. Check the new version with `huzbackend version` (or look at the footer of the
   web dashboard).

> Your configuration (`.env`), database (`data/app.db`), and autostart settings
> are kept intact because they live next to the binary.

---

## 19. Troubleshooting

### The dashboard won't open

- Make sure the server is running: `huzbackend status` should say
  `Status: 🟢 RUNNING`.
- Check the port: the default is `3300`. If you changed it, use the new port in
  the URL (`http://127.0.0.1:<PORT>`).
- If `status` says another process is already on the port, stop that process or
  change `PORT` (`huzbackend config set PORT 3301`, then `huzbackend restart`).

### "Not signed in or session expired"

- Your session cookie expired or was cleared. Sign in again.
- If it keeps happening, `SESSION_PERSISTENT=false` limits sessions to 7 days;
  set it back to `true` (default) for long-lived sessions.

### "Too many failed login attempts"

- The server rate-limits failed sign-ins per IP + username for **15 minutes**.
- Wait 15 minutes, or restart the server (the counter is in-memory and resets on
  restart).

### Camera shows but no video

- Verify the phone and the computer are on the **same local network**.
- Check that your browser supports WebRTC (Chrome, Edge, Firefox, Safari).
- Corporate networks / strict Wi-Fi can block peer-to-peer UDP traffic; try a
  different network or enable port forwarding on the phone's port range.
- Make sure the phone screen is on and the HuzHome app is running.

### Camera not appearing on the Camera page

- Confirm HuzHome is connected: it should show a `registered` confirmation.
- Open the browser **Console** (F12) and look for WebSocket errors — you must be
  **signed in** to view cameras (viewers without a session get close code 4001).
- If a device with the same ID reconnected, the old connection is kicked
  (close code 4002) — restart the app and it should register again.

### "No active network interface found" on the Devices page

- The server needs at least one active, non-loopback IPv4 interface to scan.
- Check your network cable / Wi-Fi. VPNs that remove local interfaces can also
  cause this.

### macOS "cannot be opened because the developer cannot be verified"

- Right-click the file in Finder → **Open**, then click **Open** in the dialog.
- Or in **System Settings → Privacy & Security**, click **Open Anyway** next to
  the blocked app.
- Or clear the quarantine attribute:
  `xattr -dr com.apple.quarantine ./huzbackend-darwin-arm64`

### Windows SmartScreen "Windows protected your PC"

- Click **More info → Run anyway**. The binary is open-source and unsigned;
  you can verify the SHA-256 checksum published with each release before running.

### I forgot the admin password

- There is no built-in "forgot password" flow.
- If you have **terminal access** to the server machine, you can reset the
  configuration with `huzbackend config set ADMIN_PASSWORD '<new>'` — but this
  only seeds new databases. For an **existing** account, the practical reset is:
  1. Stop the server (`huzbackend stop`).
  2. Back up `data/app.db`, then delete it.
  3. Start the server again — a fresh admin account is created from `.env`
     defaults (change the password right away).

### How do I see what the server is doing?

- `huzbackend logs -n 200` shows the last 200 log lines.
- `huzbackend logs -f` follows new lines in real time (like `tail -f`).
- Logs are written to `.huzbackend.log` next to the binary.

---

## 20. Frequently asked questions

**Is Huz CCTV really free?**
Yes. The server is released under the **MIT license** and is free to download
from SourceForge.

**Does the video go through the server or the internet?**
No. Video is **peer-to-peer** over WebRTC. The server only relays tiny signaling
messages (offers, answers, ICE candidates, and control commands). Video never
leaves your local network unless you deliberately expose it.

**Does the server record or store video?**
No. It is a live-viewing system. There is no recording, no cloud storage, and no
video files on the server.

**Can I access my cameras when I'm away from home?**
Yes, with extra setup. You can use a **VPN** (recommended) or configure **port
forwarding** for the server port (3300) on your router — and you should also put
the server behind **HTTPS** and set `COOKIE_SECURE=true`. Never expose an
unsecured server directly to the internet.

**Can I use more than one camera?**
Yes — add as many old phones as you like. Each runs HuzHome and appears as its
own camera card.

**Do I need a powerful computer?**
No. A small, low-power machine is enough; the server is a few MB of RAM.

**Why would I use an old phone as a camera?**
It gives new life to retired hardware and lets you place cameras where you want,
powered by Wi-Fi and a charger, with no dedicated camera hardware required.

**Which languages does the dashboard support?**
English (default) and Vietnamese.

---

## 21. Security notes

- **Change the default password.** The default `admin` / `onemilusd` is public
  knowledge. Change it after your first sign-in.
- **The server binds to `0.0.0.0`.** Anyone on your LAN can reach the sign-in
  page. A strong password is your first line of defense.
- **For internet access, use HTTPS + `COOKIE_SECURE=true`.** Put the server
  behind a reverse proxy (Caddy, nginx, Traefik) or a VPN. Without HTTPS, session
  cookies and passwords can be sniffed in transit.
- **Brute-force protection is built in.** The rate limiter (per IP + username,
  15-minute block) slows down password guessing.
- **Don't delete or share `data/app.db`** — it contains the password hashes and
  sessions for your server.
- **Trust your network.** WebRTC peers connect directly; make sure your camera
  phones are on a network you control.
- **Keep the server updated.** Install new releases from SourceForge to get
  security and stability fixes.

---

## 22. License and support

- **License:** [MIT License](https://opensource.org/license/mit) — the server is
  free and open source.
- **Project page:** https://sourceforge.net/projects/huz-cctv/
- **Author:** `sagehuz`
- **Reporting issues:** use the project's **Support** / **Bugs** tracker on the
  SourceForge project page, and include the version (`huzbackend version`) and
  the last log lines (`huzbackend logs -n 100`).
- **Feedback:** post a review on SourceForge to help others find the project.

---

*Thank you for using Huz CCTV. Stay safe, secure, and reliable.*
