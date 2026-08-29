package auth

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"huzbackend-go/internal/config"
	"huzbackend-go/internal/store"
)

func TestLoginAndSessionFlow(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "app.db")
	cfg := &config.Config{Port: "3300", AdminUsername: "admin", AdminPassword: "onemilusd", CookieSecure: false, SessionPersistent: true, DBPath: dbPath}
	st, err := store.NewStore(dbPath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer st.Close()

	a, err := NewAuthService(st, cfg)
	if err != nil {
		t.Fatalf("new auth: %v", err)
	}
	if err := a.EnsureAdmin(); err != nil {
		t.Fatalf("ensure admin: %v", err)
	}

	user, err := a.Login("admin", "onemilusd", "127.0.0.1")
	if err != nil {
		t.Fatalf("login admin: %v", err)
	}
	if user == nil || user.Username != "admin" {
		t.Fatal("admin login should succeed")
	}

	token, err := a.IssueSession(user.ID)
	if err != nil {
		t.Fatalf("issue session: %v", err)
	}
	req := httptest.NewRequest("GET", "/api/auth/me", nil)
	req.AddCookie(&http.Cookie{Name: "huz_session", Value: token})
	if _, _, err := a.CurrentUserFromRequest(req); err != nil {
		t.Fatalf("current user from request: %v", err)
	}
}

func TestRateLimitFivePer15Minutes(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "app.db")
	cfg := &config.Config{Port: "3300", AdminUsername: "admin", AdminPassword: "onemilusd", CookieSecure: false, SessionPersistent: true, DBPath: dbPath}
	st, err := store.NewStore(dbPath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer st.Close()

	a, err := NewAuthService(st, cfg)
	if err != nil {
		t.Fatalf("new auth: %v", err)
	}
	if err := a.EnsureAdmin(); err != nil {
		t.Fatalf("ensure admin: %v", err)
	}

	for i := 0; i < 5; i++ {
		_, err := a.Login("admin", "wrongpass", "10.0.0.1")
		if err == nil {
			t.Fatalf("login attempt %d unexpectedly succeeded", i+1)
		}
	}
	_, err = a.Login("admin", "onemilusd", "10.0.0.1")
	if err == nil {
		t.Fatal("rate limit should block after five wrong attempts")
	}
}
