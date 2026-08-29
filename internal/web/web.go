package web

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
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
	startedAt time.Time
}

func NewHandler(cfg *config.Config, authSvc *auth.AuthService, scanner *scan.Scanner, hub *signal.Hub, static fs.FS) *Handler {
	return &Handler{cfg: cfg, auth: authSvc, scanner: scanner, hub: hub, static: static, startedAt: time.Now()}
}

func (h *Handler) Mux() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", h.health)
	mux.HandleFunc("/api/auth/login", h.login)
	mux.HandleFunc("/api/auth/logout", h.logout)
	mux.HandleFunc("/api/auth/me", h.requireAuth(h.me))
	mux.HandleFunc("/api/auth/change-password", h.requireAuth(h.changePassword))
	mux.HandleFunc("/api/network-devices", h.requireAuth(h.networkDevices))
	mux.HandleFunc("/api/server-info", h.requireAuth(h.serverInfo))
	mux.HandleFunc("/ws/signal", h.hub.HandleWS)
	mux.HandleFunc("/", h.serveStatic)
	return mux
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"code": "ok", "message": "Server is running successfully"})
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"code": "invalid_input", "message": "Please enter username and password"})
		return
	}
	username, _ := body["username"].(string)
	password, _ := body["password"].(string)
	remember := true
	if v, ok := body["remember"].(bool); ok {
		remember = v
	}
	code, resp := h.auth.LoginJSON(username, password, auth.ClientIP(r))
	if code == http.StatusOK {
		token := fmt.Sprint(resp["token"])
		h.auth.SetSessionCookie(w, token, remember)
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
	_ = json.NewEncoder(w).Encode(map[string]string{"code": "ok", "message": "Signed out"})
}

func (h *Handler) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _, err := h.auth.CurrentUserFromRequest(r)
		if err != nil || user == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "unauthorized", "message": "Not signed in or session expired"})
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
		_ = json.NewEncoder(w).Encode(map[string]string{"code": "missing_password", "message": "Missing password information"})
		return
	}
	if body["currentPassword"] == "" || body["newPassword"] == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"code": "missing_password", "message": "Missing password information"})
		return
	}
	if len(body["newPassword"]) < 8 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"code": "password_too_short", "message": "New password must be at least 8 characters"})
		return
	}
	user := userFromContext(r.Context())
	if err := h.auth.ChangePassword(user.ID, body["currentPassword"], body["newPassword"]); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"code": "wrong_current_password", "message": "Current password is incorrect"})
		return
	}
	if token, err := r.Cookie("huz_session"); err == nil && token.Value != "" {
		if err := h.auth.RevokeOtherSessions(user.ID, token.Value); err != nil {
			log.Printf("revoke sessions: %v", err)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"code": "ok", "message": "Password changed successfully"})
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
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "no_network_iface", "message": "No active network interface found"})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"code": "scan_failed", "message": "Could not get the network device list", "error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

func (h *Handler) serverInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	hostname, _ := os.Hostname()
	ips, primaryIP := localIPv4s()
	info := map[string]any{
		"hostname":   hostname,
		"ip":         primaryIP,
		"ips":        ips,
		"port":       h.cfg.Port,
		"uptime":     int64(time.Since(h.startedAt).Seconds()),
		"started_at": h.startedAt.UTC().Format(time.RFC3339),
		"now":        time.Now().UTC().Format(time.RFC3339),
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
		"go_version": runtime.Version(),
		"num_cpu":    runtime.NumCPU(),
		"version":    "2026.1",
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(info)
}

func localIPv4s() ([]string, string) {
	ifs, err := net.Interfaces()
	if err != nil {
		return nil, ""
	}
	var ips []string
	var primary string
	for _, i := range ifs {
		if i.Flags&net.FlagUp == 0 || i.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			ipnet, ok := a.(*net.IPNet)
			if !ok || ipnet == nil || ipnet.IP == nil {
				continue
			}
			v4 := ipnet.IP.To4()
			if v4 == nil {
				continue
			}
			ip := v4.String()
			if ip == "" {
				continue
			}
			ips = append(ips, ip)
			if primary == "" {
				primary = ip
			}
		}
	}
	return ips, primary
}

func (h *Handler) serveStatic(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		path = "index.html"
	}
	// Pages that require sign-in.
	if path == "index.html" || path == "camera.html" || path == "devices.html" {
		if _, _, err := h.auth.CurrentUserFromRequest(r); err != nil {
			redirectPath := url.QueryEscape(r.URL.Path)
			http.Redirect(w, r, "/login.html?next="+redirectPath, http.StatusFound)
			return
		}
	}
	// Signed-in users visiting the login page are sent to the dashboard.
	if path == "login.html" {
		if _, _, err := h.auth.CurrentUserFromRequest(r); err == nil {
			http.Redirect(w, r, "/index.html", http.StatusFound)
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
