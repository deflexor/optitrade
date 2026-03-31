package dashboard

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/dfr/optitrade/src/internal/state/sqlite"
	"golang.org/x/crypto/bcrypt"
)

func TestPositionsHandlersNilExchange(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	db, err := sqlite.Open(filepath.Join(dir, "s.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	hash, _ := bcrypt.GenerateFromPassword([]byte("x"), bcrypt.MinCost)
	auth := &DashboardAuthFile{Version: "1", Users: []AuthUserRecord{{Username: "op", PasswordHash: string(hash)}}}
	srv := NewServer(Options{
		Logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
		Auth:     auth,
		Sessions: NewSessionStore(db),
		TestExchange: nil,
	})
	h := srv.Handler()

	jar := make(map[string]string)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"username":"op","password":"x"}`))
	h.ServeHTTP(rec, req)
	for _, c := range rec.Result().Cookies() {
		if c.Name == sessionCookieName {
			jar[sessionCookieName] = c.Value
		}
	}
	if jar[sessionCookieName] == "" {
		t.Fatal("no session cookie")
	}

	for _, path := range []string{"/api/v1/positions/open", "/api/v1/positions/closed"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: jar[sessionCookieName]})
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("%s: status %d body %s", path, rec.Code, rec.Body.String())
		}
		var body struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		if body.Error != "exchange_unavailable" || body.Message == "" {
			t.Fatalf("%s: %+v", path, body)
		}
	}
}
