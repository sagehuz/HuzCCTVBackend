package store

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type User struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
	Salt         string `json:"-"`
	CreatedAt    string `json:"created_at,omitempty"`
}

type Session struct {
	ID         int64
	TokenHash  string
	UserID     int64
	CreatedAt  string
	ExpiresAt  string
	LastUsedAt sql.NullString
}

type Store struct {
	db *sql.DB
}

func NewStore(dbPath string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	db.SetMaxOpenConns(1)

	st := &Store{db: db}
	if err := st.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return st, nil
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) DB() *sql.DB { return s.db }

func (s *Store) migrate() error {
	ddl := `
		PRAGMA journal_mode = WAL;
		PRAGMA foreign_keys = ON;
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
	`
	_, err := s.db.Exec(ddl)
	return err
}

func (s *Store) CreateUser(username, passwordHash, salt string) (int64, error) {
	res, err := s.db.Exec(`INSERT INTO users(username, password_hash, salt) VALUES (?, ?, ?)`, strings.TrimSpace(username), passwordHash, salt)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) FindUserByUsername(username string) (*User, error) {
	row := s.db.QueryRow(`SELECT id, username, password_hash, salt, created_at FROM users WHERE username = ?`, strings.TrimSpace(username))
	u := &User{}
	if err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Salt, &u.CreatedAt); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Store) FindUserByID(id int64) (*User, error) {
	row := s.db.QueryRow(`SELECT id, username, password_hash, salt, created_at FROM users WHERE id = ?`, id)
	u := &User{}
	if err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Salt, &u.CreatedAt); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Store) CreateSession(tokenHash string, userID int64, expiresAt string) error {
	_, err := s.db.Exec(`INSERT INTO sessions(token_hash, user_id, expires_at) VALUES (?, ?, ?)`, tokenHash, userID, expiresAt)
	return err
}

func (s *Store) DeleteSessionByToken(tokenHash string) error {
	_, err := s.db.Exec(`DELETE FROM sessions WHERE token_hash = ?`, tokenHash)
	return err
}

func (s *Store) DeleteSessionsForUserExcept(userID int64, keepTokenHash string) error {
	_, err := s.db.Exec(`DELETE FROM sessions WHERE user_id = ? AND token_hash != ?`, userID, keepTokenHash)
	return err
}

func (s *Store) ValidateSession(token string) (*User, error) {
	tokenHash := sha256Hex(token)
	row := s.db.QueryRow(`SELECT u.id, u.username, u.password_hash, u.salt, u.created_at FROM sessions s JOIN users u ON u.id = s.user_id WHERE s.token_hash = ? AND s.expires_at > ?`, tokenHash, time.Now().UTC().Format(time.RFC3339Nano))
	u := &User{}
	if err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Salt, &u.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			_ = s.DeleteSessionByToken(tokenHash)
			return nil, nil
		}
		return nil, err
	}
	if err := s.UpdateSessionLastUsed(tokenHash); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Store) UpdateSessionLastUsed(tokenHash string) error {
	_, err := s.db.Exec(`UPDATE sessions SET last_used_at = ? WHERE token_hash = ? AND (last_used_at IS NULL OR datetime(last_used_at) < datetime('now', '-5 minutes'))`, time.Now().UTC().Format(time.RFC3339Nano), tokenHash)
	return err
}

func (s *Store) SetPassword(userID int64, passwordHash, salt string) error {
	_, err := s.db.Exec(`UPDATE users SET password_hash = ?, salt = ? WHERE id = ?`, passwordHash, salt, userID)
	return err
}

func (s *Store) EnsureSessionPersistent() error {
	_, err := s.db.Exec(`UPDATE sessions SET expires_at = '9999-12-31T23:59:59.999Z' WHERE expires_at < '9999-12-31T23:59:59.999Z'`)
	return err
}

func sha256Hex(v string) string {
	sum := sha256.Sum256([]byte(v))
	return hex.EncodeToString(sum[:])
}

func sha256Sum(b []byte) []byte {
	sum := sha256.Sum256(b)
	return sum[:]
}
