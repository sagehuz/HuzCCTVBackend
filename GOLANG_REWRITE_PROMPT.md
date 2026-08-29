
Bạn là kỹ sư backend Go senior. Nhiệm vụ: **Tạo project bằng golang, đây là backend** — tên project: "Huz CCTV Server"
Mục tiêu chính:

1. có thể biên dịch trên **Linux, macOS, Windows**.
2. **Cross-compile dễ dàng**: dùng `CGO_ENABLED=0` (thuần Go), không cần toolchain C của máy đích.
3. Máy đích **không cần cài bất kỳ runtime nào**.
4. Nhúng toàn bộ tài nguyên tĩnh (frontend) và bảng OUI vào binary bằng `go:embed`.

### Tổng quan hệ thống (phải giữ nguyên)

Huz CCTV Server là backend của hệ thống CCTV biến điện thoại Android cũ (app **HuzHome**) thành camera:

1. **HTTP server** phục vụ trang web tĩnh (đăng nhập, xem camera, danh sách thiết bị) + REST API.
2. **WebRTC signaling server** tại `ws://<host>:<port>/ws/signal` — chỉ **relay** tin nhắn JSON
   (SDP/ICE) giữa thiết bị (`role: "device"`, app Android) và người xem (`role: "viewer"`,
   trình duyệt đã đăng nhập). Server KHÔNG xử lý/lưu trữ luồng video — video đi **peer-to-peer**.
3. **Quét thiết bị LAN**: ping sweep + đọc bảng ARP (`ip neighbor show` / fallback `arp -a`) + làm
   giàu thông tin (vendor OUI, reverse DNS).
4. **Xác thực**: SQLite + session cookie httpOnly, hash mật khẩu bằng scrypt, rate-limit chống
   brute-force, tài khoản admin tự tạo lần đầu. 

### Công nghệ Go bắt buộc dùng

| Thành phần | Lựa chọn | Lý do |
|---|---|---|
| Go | 1.22+ | `net/http.ServeMux` hỗ trợ method + path pattern, giảm dependency |
| HTTP router | stdlib `net/http` (hoặc `github.com/go-chi/chi/v5`) | thuần Go |
| WebSocket | `github.com/gorilla/websocket` | thuần Go, đầy đủ ping/pong + close code |
| SQLite | `modernc.org/sqlite` (**bắt buộc**) | **thuần Go**, chạy được với `CGO_ENABLED=0` → cross-compile mọi nền tảng. **KHÔNG** dùng `mattn/go-sqlite3` (cần CGO) |
| scrypt | `golang.org/x/crypto/scrypt` | phải cùng tham số với Node (xem mục Auth) để đọc DB/mật khẩu cũ |
| .env | tự viết nhỏ hoặc `github.com/joho/godotenv` | env thật ghi đè file `.env` |
| UUID | tự sinh bằng `crypto/rand` (hoặc `github.com/google/uuid`) | `clientId` của WebSocket |
| OUI | nhúng bảng OUI qua `go:embed` | tra vendor offline |

### Cấu trúc thư mục đề xuất

```
huzbackend-go/
├── cmd/huzbackend/main.go     # entry: config, seed admin, HTTP + WS, graceful shutdown
├── internal/
│   ├── config/config.go       # đọc .env + biến môi trường
│   ├── store/store.go         # SQLite: schema, users, sessions
│   ├── auth/auth.go           # scrypt, session, cookie, middleware, rate limit
│   ├── signal/signal.go       # WebSocket signaling /ws/signal
│   ├── scan/scan.go           # quét LAN, ping sweep, parse ARP
│   ├── oui/                   # go:embed bảng OUI + hàm lookup
│   └── web/web.go             # embed frontend + page-auth middleware
├── public/                    # frontend hiện tại (giữ nguyên file)
├── go.mod / go.sum
├── Makefile                   # build cross-platform
├── .env.example
├── scripts/                   # start/stop/restart cho macOS/Linux/Windows
└── README.md
```

### Cấu hình (.env — giữ nguyên tên biến và ý nghĩa)

| Biến | Mặc định | Ghi chú |
|---|---|---|
| `PORT` | `3300` | bind `0.0.0.0:PORT` |
| `ADMIN_USERNAME` | `admin` | tài khoản admin tự tạo lần đầu |
| `ADMIN_PASSWORD` | `onemilusd` | |
| `COOKIE_SECURE` | `false` | `true` khi chạy sau HTTPS |
| `SESSION_PERSISTENT` | `true` | chỉ `"false"` là tắt (phiên hết hạn sau 7 ngày) |
| `DB_PATH` | `data/app.db` | tương đối với thư mục làm việc; tự tạo thư mục cha |

- Không có `.env` vẫn chạy với mặc định. In log khi khởi động: `Server đang chạy tại port <PORT>`.
- **Schema SQLite giữ NGUYÊN bản cũ** để đọc được `data/app.db` hiện có. Pragmas:
  `journal_mode = WAL`, `foreign_keys = ON`. DDL chính xác:

```sql
CREATE TABLE IF NOT EXISTS users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  username TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  salt TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS sessions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  token_hash TEXT NOT NULL UNIQUE,
  user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  expires_at TEXT NOT NULL,
  last_used_at TEXT
);

CREATE INDEX IF NOT EXISTS idx_sessions_token_hash ON sessions(token_hash);
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
```


**Session:**
- Token: 32 bytes ngẫu nhiên → hex (64 ký tự); DB chỉ lưu `sha256(token)` dạng hex.
- Cookie `huz_session`: `HttpOnly`, `SameSite=Lax`, `Path=/`, `Secure` khi `COOKIE_SECURE=true`.
- `Max-Age` (Go tính bằng **giây**): persistent → `315360000` (10 năm); ngược lại → `604800` (7 ngày).
- `expires_at`: persistent → `'9999-12-31T23:59:59.999Z'`; ngược lại → `now + 7 ngày` (ISO8601 UTC).
- Khi khởi động với `SESSION_PERSISTENT=true`: chạy `UPDATE sessions SET expires_at =
  '9999-12-31T23:59:59.999Z' WHERE expires_at < '9999-12-31T23:59:59.999Z'` (gia hạn phiên như bản cũ).
- Kiểm tra token: join `sessions` + `users` theo `token_hash`; hết hạn → xóa session, trả null;
  `last_used_at` chỉ cập nhật khi null hoặc cũ hơn ~5 phút (throttle ghi DB).

**Rate limit đăng nhập:** lưu trong bộ nhớ, thread-safe; key `"<clientIP>:<username>"`; tối đa **5
lần / 15 phút**; `clientIP` = mục đầu của `X-Forwarded-For` (nếu có) hoặc `RemoteAddr`.

**Seed admin:** nếu chưa tồn tại user trùng `ADMIN_USERNAME` → tạo bằng `ADMIN_PASSWORD`. Chỉ 1 lần.

### REST API — giữ nguyên route, method, status code và toàn bộ chuỗi message tiếng Việt

| Method | Path | Auth | Mô tả |
|---|---|---|---|
| GET | `/api/health` | Không | `{"message":"..."}` |
| POST | `/api/auth/login` | Không | body `{username, password}` |
| POST | `/api/auth/logout` | Không | xóa session + cookie |
| GET | `/api/auth/me` | Có | `{"user":{"id","username"}}` |
| POST | `/api/auth/change-password` | Có | body `{currentPassword, newPassword}` |
| GET | `/api/network-devices` | Có | `{"count","devices"}` |

- `POST /api/auth/login`: thiếu field → `400` `Vui lòng nhập tên đăng nhập và mật khẩu`; sai → `401`
  `Tên đăng nhập hoặc mật khẩu không đúng`; vượt rate limit → `429` `Quá nhiều lần đăng nhập sai,
  vui lòng thử lại sau 15 phút`; đúng → set cookie + `200` `{"message":"Đăng nhập thành công",
  "user":{"id":...,"username":...}}`. Username trim + cắt 64 ký tự; password phải là chuỗi.
- `POST /api/auth/logout`: xóa session theo token trong cookie, xóa cookie → `200` `{"message":"Đã
  đăng xuất"}`.
- `GET /api/auth/me`: chưa đăng nhập → `401` `Chưa đăng nhập hoặc phiên đăng nhập đã hết hạn`;
  đã đăng nhập → `200` `{"user":{"id","username"}}`.
- `POST /api/auth/change-password`: thiếu field → `400` `Thiếu thông tin mật khẩu`; `newPassword`
  < 8 ký tự → `400` `Mật khẩu mới phải có ít nhất 8 ký tự`; sai mật khẩu hiện tại → `401` `Mật khẩu
  hiện tại không đúng`; thành công → tạo salt + hash mới, **thu hồi mọi session khác của user (giữ
  session hiện tại)** → `200` `{"message":"Đã đổi mật khẩu thành công"}`.
- `GET /api/network-devices`: chi tiết ở phần "Quét LAN". Không có NIC → `500` `Không tìm thấy card
  mạng nào đang hoạt động`; lỗi khác → `500` `Không thể lấy danh sách thiết bị mạng` (kèm `error`).

### Static pages (frontend)

- Nhúng `public/` qua `go:embed`, phục vụ tại `/` (index.html), `/login.html`, `/camera.html`,
  `/devices.html`, kèm MIME type đúng.
- `/camera.html` và `/devices.html` **yêu cầu đăng nhập** → nếu chưa: redirect `302` tới
  `/login.html?next=<url-encoded path gốc>`. Middleware này chạy TRƯỚC khi phục vụ static.

### WebSocket signaling — `/ws/signal` (protocol BẤT BIẾN, app Android HuzHome đang phụ thuộc)

**Vòng đời kết nối:**
- Khi có kết nối: sinh `clientId` (UUID v4). Đọc cookie `huz_session` từ header handshake →
  `authToken` (nhớ `decodeURIComponent`).
- Trạng thái client: `{ws, role: null|"device"|"viewer", name, authToken, user, deviceId}`.
- Heartbeat: cứ **10 giây** gửi ping; nếu `isAlive == false` → terminate; có pong → đánh dấu alive
  (dọn kết nối zombie khi mạng rớt đột ngột).

**Xử lý message JSON từ client:**

1. `{type:"register", role:"device"|"viewer", name?, deviceId?}`
   - `role` không hợp lệ → bỏ qua.
   - `viewer`: phải có session hợp lệ (qua `authToken`). Không hợp lệ → gửi `{type:"error",
     message:"Bạn cần đăng nhập để xem camera"}` rồi đóng sau 200ms với close code `4001` (reason
     `Unauthorized`).
   - `device`: `deviceId` = `msg.deviceId` (string, trim, cắt 128 ký tự; rỗng → `"d_"+clientId`).
     Nếu `devicesById[deviceId]` trỏ tới client khác còn sống → gửi tới client cũ `{type:"error",
     message:"Thiết bị đã kết nối lại ở phiên mới, đóng kết nối cũ"}`, xóa mapping, đóng client cũ
     code `4002` (reason `replaced`). Sau đó set `devicesById[deviceId] = clientId`.
   - `name`: cắt 64 ký tự; rỗng → `"Thiết bị " + 6 ký tự đầu của clientId`.
   - Gửi lại `{type:"registered", id: clientId}`; gọi `broadcastDeviceList()`.

2. Relay nguyên văn — các loại sau chỉ **thêm trường `from: clientId`** rồi chuyển tới đích
   `clients[msg.targetId]`: `watch`, `offer`, `answer`, `ice-candidate`, `control`, `capabilities`,
   `device-status`, `snapshot`.
   - Không tìm thấy đích → gửi người gửi `{type:"error", message:"Thiết bị đích không còn kết nối"}`.
   - Có đích → gửi `{...msg, from: clientId}`.

3. Loại message khác → bỏ qua.

**`broadcastDeviceList()`:** lấy mọi client `role=="device"` → `[{id, name, deviceId}]`; gửi
`{type:"device-list", devices:[...]}` tới **mọi client `role=="viewer"`**.

**Khi đóng kết nối:** xóa client; nếu `role=="device"` và `devicesById[deviceId]==clientId` → xóa
mapping (tránh xóa nhầm của kết nối mới sau reconnect); nếu `role=="device"` →
`broadcastDeviceList()`.

**Đồng bộ:** dùng `sync.RWMutex` (hoặc tương đương) bảo vệ `clients` và `devicesById`.

### Quét LAN — `GET /api/network-devices` (giữ nguyên thuật toán bản Node)

1. **Liệt kê subnet:** duyệt các interface mạng (dùng `net.Interfaces` + địa chỉ IPv4), lấy IPv4
   **không phải loopback** kèm netmask.
2. **Sinh dải host:** với mỗi subnet tính `network = ip & mask`, `broadcast = network | ^mask`; sinh
   mọi IP từ `network+1` đến `broadcast-1` (bỏ địa chỉ mạng + broadcast). Gộp lại, **giới hạn 1024
   host đầu tiên**.
3. **Ping sweep** (đánh thức bảng ARP, không dùng kết quả): concurrency **16**; Windows →
   `ping -n 1 -w 500 <ip>`, các OS khác → `ping -c 1 -W 1 <ip>`; timeout 2s/lệnh.
4. **Đọc bảng ARP** (retry 2 lần, timeout 10s): Linux thử `ip neighbor show` trước, lỗi thì fallback
   `arp -a`; macOS/Windows dùng `arp -a`.
5. **Parse**:
   - `ip neighbor show`: regex tương đương
     `^([\d.]+)\s+dev\s+(\S+)(?:\s+lladdr\s+([0-9a-fA-F:]+))?.*\s(\S+)$`; bỏ dòng thiếu `lladdr`
     hoặc state `FAILED` → `{ip, mac, hostname:null, iface, state}`.
   - `arp -a`: regex tương đương
     `^(\S+)?\s*\(([\d.]+)\)\s+at\s+([0-9a-fA-F:]{11,17})(?:.*on\s+(\S+))?`; bỏ mac chứa
     `incomplete`; hostname `?` → null → `{ip, mac, hostname, iface, state:null}`.
   - Lọc: bỏ nếu không có mac, mac `ff:ff:ff:ff:ff:ff`, hoặc octet đầu của IP thuộc **224–239**
     (multicast).
6. **Làm giàu** (concurrency 16):
   - `vendor`: mac bỏ `:` → hex in hoa; tra bảng OUI theo thứ tự ưu tiên **9 → 7 → 6** ký tự hex
     đầu; vendor = dòng đầu entry (hoặc null).
   - `hostname`: nếu chưa có → reverse DNS (`net.LookupAddr`) với timeout **1.5s** (dùng context);
     lỗi → null.
7. Trả về `{"count": n, "devices": [...]}` — field: `ip`, `mac`, `hostname`, `iface`, `state`,
   `vendor`.

**OUI database:** nhúng bảng OUI vào binary (nguồn: gói npm `oui-data` hoặc `oui.txt` IEEE). Ghi rõ
cách tái sinh dữ liệu trong README.

### Build cross-platform (trọng tâm của yêu cầu)

- Mọi dependency **thuần Go**; build luôn với `CGO_ENABLED=0`.
- Makefile với các target: `build-linux` (linux/amd64 + linux/arm64), `build-macos`
  (darwin/amd64 + darwin/arm64), `build-windows` (windows/amd64 → `huzbackend.exe`), `build-all`
  (xuất vào `dist/`). Có thể thêm `goreleaser` nếu muốn phát hành release.
- Trước khi bàn giao: `go build`, `go vet ./...`, `go test ./...` phải sạch.

### Scripts chạy (thay thế bản Node, giữ trải nghiệm người dùng)

- `scripts/macos/start.command`, `stop.command`, `restart.command` — double-click trong Finder; gọi
  các `.sh`; tự tạo `.env` từ `.env.example` nếu chưa có; in URL local + LAN; tự mở trình duyệt
  (macOS); chỉ stop đúng binary của project (dùng PID file, không kill nhầm).
- `scripts/start.sh`, `stop.sh`, `restart.sh` — chung macOS + Linux, chạy binary thay `node index.js`.
- `scripts/windows/start.bat`, `stop.bat`, `restart.bat` — dòng đầu `chcp 65001 >nul`; start chạy
  `huzbackend.exe` + ghi PID; stop đọc PID file (dùng `taskkill /PID`).
- `.env.example` giữ nguyên biến hiện tại. README viết lại cho bản Go (không cần Node; lệnh build,
  chạy, API, troubleshooting).

### Chất lượng chung

- **Graceful shutdown**: bắt `SIGINT`/`SIGTERM` → đóng WebSocket, dừng HTTP, đóng DB.
- Log bằng thư viện `log` chuẩn; giữ message khởi động như bản cũ.
- Không panic; mọi map dùng chung đều có lock.
- Unit test tối thiểu: parse ARP 2 định dạng, rate limit, vòng đời session, scrypt tương thích DB
  cũ, relay signaling, `getHostRange`.
- **Không đổi protocol WebSocket, không đổi tên field JSON, không đổi chuỗi message tiếng Việt** —
  để app HuzHome Android và frontend hiện tại hoạt động nguyên vẹn.

### Tiêu chí nghiệm thu (acceptance criteria)

1. `CGO_ENABLED=0 go build` thành công cho linux/amd64, linux/arm64, darwin/amd64, darwin/arm64,
   windows/amd64.
2. Binary chạy độc lập trên cả 3 OS, bind `0.0.0.0:PORT`, không cần Node/npm.
3. Mọi endpoint HTTP trả đúng status + JSON như bảng API ở trên.
4. Trang web tĩnh được phục vụ từ binary (không cần thư mục `public/` bên ngoài).
5. `/camera.html`, `/devices.html` chặn đăng nhập + redirect `?next=` đúng.
6. Đăng nhập được bằng DB/mật khẩu cũ (tạo từ bản Node); session persistent đúng; rate limit 5
   lần/15 phút; đổi mật khẩu thu hồi session khác.
7. WebSocket: device register → viewer nhận `device-list`; relay đúng đích kèm `from`; deviceId
   trùng → đá kết nối cũ (4002); viewer chưa đăng nhập → đóng 4001; heartbeat dọn kết nối chết.
8. `/api/network-devices` trả danh sách thiết bị thật trên LAN có vendor/hostname.
9. Ctrl+C dừng sạch sẽ (graceful shutdown).
10. `go vet ./...` và `go test ./...` sạch.

### Ghi chú quan trọng

- Server chỉ relay signaling, không lưu video.
- Rate limit trong bộ nhớ (mất khi restart — giống bản cũ, chấp nhận được).
- Giới hạn quét 1024 host.
- OUI: ưu tiên 9 → 7 → 6 ký tự hex.
- Reverse DNS timeout 1.5s — phải dùng context.
- Cookie `Max-Age` trong Go tính bằng **giây**.
- Nếu đổi nội dung message `/api/health` (bản cũ: `Server Node.js đang chạy thành công trên
  Ubuntu!`), phải cập nhật đồng bộ README và kiểm tra frontend hiển thị đúng.

