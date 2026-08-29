package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/scrypt"

	"huzbackend-go/internal/config"
	"huzbackend-go/internal/store"
)

type AuthService struct {
	store *store.Store
	cfg   *config.Config
	mu    sync.Mutex
	loginAttempts map[string][]time.Time
}

type UserDTO struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

func NewAuthService(db *store.Store, cfg *config.Config) (*AuthService, error) {
	return &AuthService{store: db, cfg: cfg, loginAttempts: map[string][]time.Time{}}, nil
}

func (a *AuthService) EnsureAdmin() error {
	_, err := a.store.FindUserByUsername(a.cfg.AdminUsername)
	if err == nil {
		return nil
	}
	if !strings.Contains(err.Error(), "no rows") {
		return err
	}
	passwordHash, salt, err := hashPassword(a.cfg.AdminPassword, generateSalt())
	if err != nil {
		return err
	}
	_, err = a.store.CreateUser(a.cfg.AdminUsername, passwordHash, salt)
	return err
}

func (a *AuthService) Login(username, password, clientIP string) (*store.User, error) {
	trimUser := strings.TrimSpace(username)
	if trimUser == "" || strings.TrimSpace(password) == "" {
		return nil, errors.New("empty")
	}
	key := clientIP + ":" + trimUser
	if a.isRateLimited(key) {
		return nil, errRateLimit
	}
	user, err := a.store.FindUserByUsername(trimUser)
	if err != nil {
		return nil, errInvalidCreds
	}
	ok, err := verifyPassword(password, user.Salt, user.PasswordHash)
	if err != nil || !ok {
		a.recordFailure(key)
		return nil, errInvalidCreds
	}
	a.clearFailure(key)
	return user, nil
}

func (a *AuthService) CurrentUserFromRequest(r *http.Request) (*store.User, string, error) {
	cookie, err := r.Cookie("huz_session")
	if err != nil {
		return nil, "", errors.New("not authenticated")
	}
	token := strings.TrimSpace(cookie.Value)
	if token == "" {
		return nil, "", errors.New("not authenticated")
	}
	token = strings.Trim(token, "\"")
	if token == "" {
		return nil, "", errors.New("not authenticated")
	}
	user, err := a.store.ValidateSession(token)
	if err != nil || user == nil {
		return nil, "", errors.New("not authenticated")
	}
	return user, token, nil
}

func (a *AuthService) Logout(r *http.Request, w http.ResponseWriter) {
	if token, err := sessionTokenFromRequest(r); err == nil {
		_ = a.store.DeleteSessionByToken(hashToken(token))
	}
	clearSessionCookie(w)
}

func (a *AuthService) IssueSession(userID int64) (string, error) {
	token, err := randomToken()
	if err != nil {
		return "", err
	}
	tokenHash := hashToken(token)
	expiresAt := time.Now().UTC().Add(7 * 24 * time.Hour).Format(time.RFC3339Nano)
	if a.cfg.SessionPersistent {
		expiresAt = "9999-12-31T23:59:59.999Z"
		if err := a.store.EnsureSessionPersistent(); err != nil {
			return "", err
		}
	}
	if err := a.store.CreateSession(tokenHash, userID, expiresAt); err != nil {
		return "", err
	}
	return token, nil
}

func (a *AuthService) SetSessionCookie(w http.ResponseWriter, token string) {
	maxAge := 7 * 24 * 60 * 60
	if a.cfg.SessionPersistent {
		maxAge = 315360000
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "huz_session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   a.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   maxAge,
	})
}

func (a *AuthService) ChangePassword(userID int64, currentPassword, newPassword string) error {
	user, err := a.store.FindUserByID(userID)
	if err != nil {
		return err
	}
	ok, err := verifyPassword(currentPassword, user.Salt, user.PasswordHash)
	if err != nil || !ok {
		return errors.New("current password mismatch")
	}
	if len(strings.TrimSpace(newPassword)) < 8 {
		return errors.New("new password too short")
	}
	passwordHash, salt, err := hashPassword(newPassword, generateSalt())
	if err != nil {
		return err
	}
	if err := a.store.SetPassword(userID, passwordHash, salt); err != nil {
		return err
	}
	return nil
}

func (a *AuthService) isRateLimited(key string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	now := time.Now()
	entries := a.loginAttempts[key]
	filtered := entries[:0]
	for _, v := range entries {
		if now.Sub(v) <= 15*time.Minute {
			filtered = append(filtered, v)
		}
	}
	if len(filtered) >= 5 {
		a.loginAttempts[key] = filtered
		return true
	}
	a.loginAttempts[key] = filtered
	return false
}

func (a *AuthService) recordFailure(key string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.loginAttempts[key] = append(a.loginAttempts[key], time.Now())
}

func (a *AuthService) clearFailure(key string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.loginAttempts, key)
}

func hashPassword(password, salt string) (string, string, error) {
	if salt == "" {
		salt = generateSalt()
	}
	dk, err := scrypt.Key([]byte(password), []byte(salt), 1<<14, 8, 1, 32)
	if err != nil {
		return "", "", err
	}
	return hex.EncodeToString(dk), salt, nil
}

func verifyPassword(password, salt, storedHash string) (bool, error) {
	if strings.TrimSpace(password) == "" || strings.TrimSpace(salt) == "" {
		return false, nil
	}
	derived, _, err := hashPassword(password, salt)
	if err != nil {
		return false, err
	}
	return storedHash == derived, nil
}

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func sessionTokenFromRequest(r *http.Request) (string, error) {
	cookie, err := r.Cookie("huz_session")
	if err != nil {
		return "", err
	}
	v := strings.TrimSpace(cookie.Value)
	if v == "" {
		return "", errors.New("empty")
	}
	return v, nil
}

func ClientIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		parts := strings.Split(forwarded, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	if r.RemoteAddr != "" {
		ip := strings.Split(r.RemoteAddr, ":")
		if len(ip) > 0 {
			return ip[0]
		}
	}
	return "unknown"
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "huz_session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func generateSalt() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

var (
	errInvalidCreds = errors.New("invalid credentials")
	errRateLimit    = errors.New("rate limit")
)

func (a *AuthService) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, _, err := a.CurrentUserFromRequest(r)
		if err != nil || user == nil {
			http.Error(w, "Chưa đăng nhập hoặc phiên đăng nhập đã hết hạn", http.StatusUnauthorized)
			return
		}
		r = r.WithContext(contextWithUser(r.Context(), user))
		next.ServeHTTP(w, r)
	})
}

func contextWithUser(ctx context.Context, user *store.User) context.Context {
	return context.WithValue(ctx, authUserKey{}, user)
}

type authUserKey struct{}

func UserFromContext(ctx context.Context) *store.User {
	v, ok := ctx.Value(authUserKey{}).(*store.User)
	if !ok {
		return nil
	}
	return v
}

func (a *AuthService) LoginJSON(username, password, clientIP string) (int, map[string]any) {
	if strings.TrimSpace(username) == "" || strings.TrimSpace(password) == "" {
		return http.StatusBadRequest, map[string]any{"message": "Vui lòng nhập tên đăng nhập và mật khẩu"}
	}
	user, err := a.Login(username, password, clientIP)
	if err != nil {
		if errors.Is(err, errRateLimit) {
			return http.StatusTooManyRequests, map[string]any{"message": "Quá nhiều lần đăng nhập sai, vui lòng thử lại sau 15 phút"}
		}
		return http.StatusUnauthorized, map[string]any{"message": "Tên đăng nhập hoặc mật khẩu không đúng"}
	}

	token, err := a.IssueSession(user.ID)
	if err != nil {
		return http.StatusInternalServerError, map[string]any{"message": "Không thể tạo phiên đăng nhập"}
	}
	return http.StatusOK, map[string]any{"message": "Đăng nhập thành công", "user": map[string]any{"id": user.ID, "username": user.Username}, "token": token}
}

func (a *AuthService) UserDTOFromUser(user *store.User) UserDTO {
	return UserDTO{ID: user.ID, Username: user.Username}
}

func parseInt64(v string) int64 {
	if i, err := strconv.ParseInt(v, 10, 64); err == nil {
		return i
	}
	return 0
}
