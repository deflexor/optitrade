package dashboard

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"time"
)

// SessionStore persists opaque dashboard sessions (token hashed at rest).
type SessionStore struct {
	db *sql.DB
}

func NewSessionStore(db *sql.DB) *SessionStore {
	return &SessionStore{db: db}
}

func hashSessionToken(cookieToken string) string {
	sum := sha256.Sum256([]byte(cookieToken))
	return hex.EncodeToString(sum[:])
}

// Create persists a new session row. cookieToken is the raw cookie value (stored hashed).
func (s *SessionStore) Create(ctx context.Context, username, cookieToken, userAgent string) error {
	th := hashSessionToken(cookieToken)
	now := time.Now().UnixMilli()
	var ua any
	if userAgent != "" {
		ua = userAgent
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO dashboard_session(username, token_hash, created_at, user_agent) VALUES(?,?,?,?)`,
		username, th, now, ua,
	)
	return err
}

// DeleteByCookieToken removes the session for the given raw cookie value.
func (s *SessionStore) DeleteByCookieToken(ctx context.Context, cookieToken string) error {
	if cookieToken == "" {
		return nil
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM dashboard_session WHERE token_hash = ?`, hashSessionToken(cookieToken))
	return err
}

// LookupUsername resolves the session to a username.
func (s *SessionStore) LookupUsername(ctx context.Context, cookieToken string) (string, error) {
	if cookieToken == "" {
		return "", sql.ErrNoRows
	}
	var u string
	err := s.db.QueryRowContext(ctx,
		`SELECT username FROM dashboard_session WHERE token_hash = ?`,
		hashSessionToken(cookieToken),
	).Scan(&u)
	return u, err
}
