package dashboard

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

type loginBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	s.authMu.RLock()
	auth := s.auth
	s.authMu.RUnlock()
	if auth == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "auth_unconfigured", "dashboard auth is not configured")
		return
	}

	var body loginBody
	raw, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_body", "could not read body")
		return
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_body", "expected json")
		return
	}
	user := strings.TrimSpace(body.Username)
	pass := body.Password
	rec := auth.Lookup(user)
	if rec == nil {
		writeAPIError(w, http.StatusUnauthorized, "invalid_credentials", "invalid credentials")
		return
	}
	if err := verifyPasswordRecord(rec.PasswordHash, pass); err != nil {
		// Same response for unknown user vs bad password (avoid account enumeration).
		writeAPIError(w, http.StatusUnauthorized, "invalid_credentials", "invalid credentials")
		return
	}
	rawTok := make([]byte, 32)
	if _, err := rand.Read(rawTok); err != nil {
		s.log.Error("session token", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "server_error", "could not create session")
		return
	}
	tok := hex.EncodeToString(rawTok)
	if err := s.sessions.Create(r.Context(), rec.Username, tok, r.UserAgent()); err != nil {
		logHandlerError(s.log, "session create", err)
		writeAPIError(w, http.StatusInternalServerError, "server_error", "could not create session")
		return
	}
	s.setSessionCookie(w, r, tok)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	c, err := r.Cookie(sessionCookieName)
	if err == nil && c != nil && c.Value != "" {
		if err := s.sessions.DeleteByCookieToken(r.Context(), c.Value); err != nil {
			logHandlerError(s.log, "session delete", err)
		}
	}
	s.clearSessionCookie(w, r)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	u, ok := requestUser(r.Context())
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "sign in required")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"username": u})
}
