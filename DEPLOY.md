# Deploy Huz CCTV Server lên Ubuntu Server (Production)

Hướng dẫn từng bước deploy file `dist/huzbackend-linux-amd64` lên một máy
**Ubuntu Server 20.04 / 22.04 / 24.04 (x86_64)**. Máy local dùng để build/upload
là macOS (đã có sẵn binary trong `dist/`).

---

## 0. Kiến trúc — bạn đang deploy cái gì?

| Đặc điểm | Giá trị |
|---|---|
| Binary | `dist/huzbackend-linux-amd64` (~10 MB) |
| Loại file | ELF 64-bit, **statically linked** (`CGO_ENABLED=0`) |
| Phụ thuộc runtime | **Không có** — không cần cài Go, glibc, node... |
| Phục vụ | Web dashboard (embedded), REST API, WebSocket `/ws/signal`, scan mạng LAN |
| Database | SQLite thuần Go, mặc định `data/app.db` (tự tạo thư mục `data/`) |
| Port mặc định | `3300` (bind `0.0.0.0`) |
| Cấu hình | File `.env` nằm **cạnh binary** (binary tự `chdir` về thư mục của nó) |
| Log | Chạy `serve` dưới systemd → log vào journald |

> ⚠️ **Không dùng lệnh `autostart on` tích hợp trên server.** Lệnh đó tạo
> systemd **user** unit (cần phiên đăng nhập mới chạy). Trên server production
> ta dùng **system unit** ở Bước 4 để chạy độc lập, tự khởi động lại khi reboot.

---

## 1. (Trên Mac) Build & kiểm tra binary

Đảm bảo binary mới nhất đã có trong `dist/`:

```bash
cd /Users/trungkhiem/Documents/sagehuz/HuzCCTV
make build-linux
file dist/huzbackend-linux-amd64      # kỳ vọng: ELF 64-bit, statically linked
shasum -a 256 dist/huzbackend-linux-amd64   # ghi lại hash để đối chiếu sau upload
```

---

## 2. (Trên Mac) Upload binary lên server

```bash
scp dist/huzbackend-linux-amd64 huzuser@SERVER_IP:/tmp/
```

Thay `huzuser` bằng user SSH của bạn, `SERVER_IP` bằng IP của server.
Nếu dùng key không phải mặc định: `scp -i ~/.ssh/your_key ...`

---

## 3. (SSH vào server) Chuẩn bị hệ thống

```bash
ssh huzuser@SERVER_IP
```

### 3.1 Cập nhật hệ thống

```bash
sudo apt update && sudo apt upgrade -y
```

### 3.2 Tạo user hệ thống chạy service (không đăng nhập được)

```bash
sudo useradd --system --home-dir /opt/huzcctv --shell /usr/sbin/nologin huzcctv
```

### 3.3 Tạo thư mục deploy & cài binary

```bash
sudo mkdir -p /opt/huzcctv
sudo mv /tmp/huzbackend-linux-amd64 /opt/huzcctv/
sudo chown -R huzcctv:huzcctv /opt/huzcctv   # QUAN TRỌNG: chown cả THƯ MỤC, không chỉ binary
sudo chmod 755 /opt/huzcctv/huzbackend-linux-amd64
```

> ⚠️ Nếu bỏ qua `chown -R` thư mục, service sẽ chạy được binary nhưng fail ngay với
> exit code 1 vì user `huzcctv` không tạo được `data/` bên trong thư mục root sở hữu.

Kiểm tra binary chạy được (chỉ in version, không khởi động server):

```bash
sudo -u huzcctv /opt/huzcctv/huzbackend-linux-amd64 version
```

### 3.4 Tạo file `.env` cấu hình production

```bash
sudo tee /opt/huzcctv/.env > /dev/null <<'EOF'
PORT=3300
ADMIN_USERNAME=admin
ADMIN_PASSWORD=THAY_DOI_PASSWORD_THANH_CAI_KHO
COOKIE_SECURE=false
SESSION_PERSISTENT=true
DB_PATH=data/app.db
EOF
sudo chown huzcctv:huzcctv /opt/huzcctv/.env
sudo chmod 600 /opt/huzcctv/.env
```

| Biến | Ý nghĩa | Lưu ý |
|---|---|---|
| `PORT` | Cổng HTTP | Mặc định `3300` |
| `ADMIN_USERNAME` / `ADMIN_PASSWORD` | Tài khoản đăng nhập dashboard | **Phải đổi password** khỏi giá trị mặc định `onemilusd` |
| `COOKIE_SECURE` | Bật cờ `Secure` cho cookie `huz_session` | `false` khi mới dùng HTTP; **đổi `true` sau khi bật HTTPS** (Bước 6) |
| `SESSION_PERSISTENT` | Session lưu trong DB, sống qua các lần restart | Giữ `true` |
| `DB_PATH` | Đường dẫn file SQLite | Tương đối với thư mục binary |

---

## 4. Cài systemd service (chạy nền, tự khởi động cùng máy)

Tạo unit `/etc/systemd/system/huzcctv.service`:

```bash
sudo tee /etc/systemd/system/huzcctv.service > /dev/null <<'EOF'
[Unit]
Description=Huz CCTV Server
Wants=network-online.target
After=network-online.target

[Service]
Type=simple
User=huzcctv
Group=huzcctv
WorkingDirectory=/opt/huzcctv
ExecStart=/opt/huzcctv/huzbackend-linux-amd64 serve
Restart=always
RestartSec=3
LimitNOFILE=65536
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=full
ProtectHome=true
ReadWritePaths=/opt/huzcctv

[Install]
WantedBy=multi-user.target
EOF
```

> File mẫu đầy đủ kèm repo: `deploy/huzcctv.service` — có thể `scp` lên server rồi
> `sudo cp deploy/huzcctv.service /etc/systemd/system/`.

Bật và khởi động:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now huzcctv
```

Kiểm tra:

```bash
sudo systemctl status huzcctv --no-pager
curl http://127.0.0.1:3300/api/health   # kỳ vọng {"code":"ok","message":"Server is running successfully"}
sudo journalctl -u huzcctv -f            # theo dõi log (Ctrl+C để thoát)
```

---

## 5. Mở firewall (UFW)

```bash
sudo ufw allow OpenSSH
```

**Trường hợp A — truy cập thẳng qua cổng 3300** (không cần domain/nginx):

```bash
sudo ufw allow 3300/tcp
sudo ufw enable
```

**Trường hợp B — dùng nginx + HTTPS (khuyến nghị)** → chỉ mở web chuẩn:

```bash
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
```

Sau đó truy cập dashboard:

- Không nginx: `http://SERVER_IP:3300`
- Có nginx: `http://cctv.example.com`

Đăng nhập bằng `ADMIN_USERNAME` / `ADMIN_PASSWORD` đã đặt ở Bước 3.4.


---

## 6. (Khuyến nghị) Nginx reverse proxy + HTTPS

Nếu bạn có domain trỏ về server, dùng nginx để có HTTPS và WebSocket ổn định.

```bash
sudo apt install -y nginx
# Chép template từ repo lên server trước, rồi:
sudo cp deploy/nginx-huzcctv.conf /etc/nginx/sites-available/huzcctv
sudo ln -s /etc/nginx/sites-available/huzcctv /etc/nginx/sites-enabled/huzcctv
sudo nginx -t && sudo systemctl reload nginx
```

Template nginx (bắt buộc có phần WebSocket `Upgrade` cho `/ws/signal`):

```nginx
server {
    listen 80;
    listen [::]:80;
    server_name cctv.example.com;
    client_max_body_size 50m;

    location / {
        proxy_pass http://127.0.0.1:3300;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
    }
}
```

Cấp chứng chỉ Let's Encrypt (tự cấu hình nginx + redirect HTTPS):

```bash
sudo apt install -y certbot python3-certbot-nginx
sudo certbot --nginx -d cctv.example.com
```

Sau khi HTTPS hoạt động, bật cookie an toàn và restart:

```bash
sudo sed -i 's/^COOKIE_SECURE=.*/COOKIE_SECURE=true/' /opt/huzcctv/.env
sudo systemctl restart huzcctv
```

Truy cập: `https://cctv.example.com`

---

## 7. Cập nhật lên version mới (deploy lại)

Quy trình an toàn — luôn backup DB trước:

```bash
# Trên Mac: build lại rồi upload
cd /Users/trungkhiem/Documents/sagehuz/HuzCCTV
make build-linux
scp dist/huzbackend-linux-amd64 huzuser@SERVER_IP:/tmp/

# Trên server
sudo systemctl stop huzcctv
sudo cp /opt/huzcctv/data/app.db /opt/huzcctv/data/app.db.bak.$(date +%F)
sudo install -o huzcctv -g huzcctv -m 755 /tmp/huzbackend-linux-amd64 /opt/huzcctv/
sudo systemctl start huzcctv
sudo systemctl status huzcctv --no-pager
curl http://127.0.0.1:3300/api/version
```

---

## 8. Backup định kỳ

Dữ liệu quan trọng: `/opt/huzcctv/data/app.db` (SQLite) + `/opt/huzcctv/.env`.

Thêm cron (chạy 2h sáng mỗi ngày):

```bash
sudo mkdir -p /backup
sudo crontab -e
# Thêm dòng:
0 2 * * * sudo tar czf /backup/huzcctv-$(date +\%F).tar.gz -C /opt/huzcctv data .env && find /backup -name 'huzcctv-*' -mtime +14 -delete
```


---

## 9. Xử lý sự cố

| Triệu chứng | Kiểm tra |
|---|---|
| Service không chạy | `sudo systemctl status huzcctv --no-pager`, `sudo journalctl -u huzcctv -e` |
| `Active: activating (auto-restart)`, exit code 1 ngay khi khởi động | Thường do thiếu quyền ghi `/opt/huzcctv`. Fix: `sudo chown -R huzcctv:huzcctv /opt/huzcctv && sudo systemctl restart huzcctv`. Xem log `journalctl -u huzcctv -e` để xác nhận lỗi `permission denied` |
| Port 3300 bận | `sudo ss -tlnp \| grep 3300` |
| Truy cập ngoài không được | `sudo ufw status`, kiểm tra security group / firewall của nhà cung cấp cloud |
| Không đăng nhập dashboard được | Sửa `ADMIN_USERNAME`/`ADMIN_PASSWORD` trong `/opt/huzcctv/.env` rồi `sudo systemctl restart huzcctv` |
| Sau khi bật HTTPS không đăng nhập được | Cookie `Secure` chỉ gửi qua HTTPS — chắc chắn truy cập bằng `https://` và `COOKIE_SECURE=true` |
| WebSocket/live video không lên | Kiểm tra nginx có đủ 2 dòng `Upgrade` + `Connection "upgrade"` ở Bước 6 |
| Tính năng quét mạng LAN không thấy thiết bị | Chạy thử `ping`/`arp -a` thủ công trên server; trên Ubuntu các lệnh này chạy được với user thường |

---

## 10. Checklist bảo mật production

- [ ] Đổi `ADMIN_PASSWORD` khỏi giá trị mặc định, đặt password mạnh.
- [ ] Cấp quyền file chặt: `.env` là `600`, thuộc user `huzcctv`.
- [ ] Bật HTTPS + `COOKIE_SECURE=true`.
- [ ] Chỉ mở đúng port cần thiết trên UFW (3300 **hoặc** 80/443 + nginx).
- [ ] Backup DB định kỳ (Bước 8).
- [ ] Không chạy server bằng user `root` (systemd unit đã dùng `User=huzcctv`).

