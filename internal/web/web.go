package web

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"huzbackend-go/internal/auth"
	"huzbackend-go/internal/config"
	"huzbackend-go/internal/scan"
	"huzbackend-go/internal/signal"
	"huzbackend-go/internal/store"
)

type Handler struct {
	cfg    *config.Config
	auth   *auth.AuthService
	scanner *scan.Scanner
	hub    *signal.Hub
	static fs.FS
}

func NewHandler(cfg *config.Config, authSvc *auth.AuthService, scanner *scan.Scanner, hub *signal.Hub, static fs.FS) *Handler {
	return &Handler{cfg: cfg, auth: authSvc, scanner: scanner, hub: hub, static: static}
}

func (h *Handler) Mux() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", h.health)
	mux.HandleFunc("/api/auth/login", h.login)
	mux.HandleFunc("/api/auth/logout", h.logout)
	mux.HandleFunc("/api/auth/me", h.requireAuth(h.me))
	mux.HandleFunc("/api/auth/change-password", h.requireAuth(h.changePassword))
	mux.HandleFunc("/api/network-devices", h.requireAuth(h.networkDevices))
	mux.HandleFunc("/ws/signal", h.hub.HandleWS)
	mux.HandleFunc("/", h.serveStatic)
	return mux
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Server Node.js đang chạy thành công trên Ubuntu!"})
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var body map[string]string
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Vui lòng nhập tên đăng nhập và mật khẩu"})
		return
	}
	code, resp := h.auth.LoginJSON(body["username"], body["password"], auth.ClientIP(r))
	if code == http.StatusOK {
		token := fmt.Sprint(resp["token"])
		h.auth.SetSessionCookie(w, token)
		delete(resp, "token")
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	h.auth.Logout(r, w)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Đã đăng xuất"})
}

func (h *Handler) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _, err := h.auth.CurrentUserFromRequest(r)
		if err != nil || user == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "Chưa đăng nhập hoặc phiên đăng nhập đã hết hạn"})
			return
		}
		r = r.WithContext(contextWithUser(r.Context(), user))
		next(w, r)
	}
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	user := userFromContext(r.Context())
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"user": map[string]any{"id": user.ID, "username": user.Username}})
}

func (h *Handler) changePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var body map[string]string
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Thiếu thông tin mật khẩu"})
		return
	}
	if body["currentPassword"] == "" || body["newPassword"] == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Thiếu thông tin mật khẩu"})
		return
	}
	if len(body["newPassword"]) < 8 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Mật khẩu mới phải có ít nhất 8 ký tự"})
		return
	}
	user := userFromContext(r.Context())
	if err := h.auth.ChangePassword(user.ID, body["currentPassword"], body["newPassword"]); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Mật khẩu hiện tại không đúng"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Đã đổi mật khẩu thành công"})
}

func (h *Handler) networkDevices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	result, err := h.scanner.NetworkDevices()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if strings.Contains(err.Error(), "no network interface") {
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "Không tìm thấy card mạng nào đang hoạt động"})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"message": "Không thể lấy danh sách thiết bị mạng", "error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

func (h *Handler) serveStatic(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		path = "index.html"
	}
	if path == "camera.html" || path == "devices.html" {
		if _, _, err := h.auth.CurrentUserFromRequest(r); err != nil {
			redirectPath := url.QueryEscape(r.URL.Path)
			http.Redirect(w, r, "/login.html?next="+redirectPath, http.StatusFound)
			return
		}
	}
	file, err := fs.ReadFile(h.static, path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", mimeType(path))
	_, _ = w.Write(file)
}

func mimeType(path string) string {
	switch filepath.Ext(path) {
	case ".html":
		return "text/html; charset=utf-8"
	case ".js":
		return "application/javascript; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".json":
		return "application/json; charset=utf-8"
	default:
		return "application/octet-stream"
	}
}

type userKey struct{}

func contextWithUser(ctx context.Context, user *store.User) context.Context {
	return context.WithValue(ctx, userKey{}, user)
}

func userFromContext(ctx context.Context) *store.User {
	user, _ := ctx.Value(userKey{}).(*store.User)
	return user
}

func init() {
	_ = os.Stdout
	_ = embed.FS{}
	_ = log.Println
	_ = time.Now
}
