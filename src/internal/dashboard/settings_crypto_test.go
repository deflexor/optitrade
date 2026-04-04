package dashboard

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
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

func TestSettingsCryptoRoundTrip(t *testing.T) {
	t.Parallel()
	key := bytes.Repeat([]byte{'k'}, 32)
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatal(err)
	}
	c := &SettingsCrypto{gcm: gcm}

	plain := []byte(`{"deribit_client_id":"abc","deribit_client_secret":"s"}`)
	blob, err := c.Encrypt(plain)
	if err != nil {
		t.Fatal(err)
	}
	out, err := c.Decrypt(blob)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, plain) {
		t.Fatalf("decrypt mismatch: %q vs %q", out, plain)
	}
}

func TestSettingsAPIMaxLossEquityPct(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	db, err := sqlite.Open(filepath.Join(dir, "s.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	crypto := testOperatorSettingsCrypto(t)
	st := NewOperatorSettingsStore(db)
	ctx := context.Background()
	prov := "deribit"
	mainnet := false
	_, err = st.Put(ctx, "op1", crypto, OperatorSettingsPatch{
		Provider:          &prov,
		DeribitUseMainnet: &mainnet,
		Secrets: map[string]string{
			"deribit_client_id":     "cid",
			"deribit_client_secret": "csec",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	auth := &DashboardAuthFile{
		Version: "1",
		Users:   []AuthUserRecord{{Username: "op1", PasswordHash: string(hash)}},
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := NewServer(Options{
		Logger:         log,
		Auth:           auth,
		Sessions:       NewSessionStore(db),
		SettingsCrypto: crypto,
		Settings:       st,
		TestExchange:   nil,
	})
	h := srv.Handler()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader([]byte(`{"username":"op1","password":"secret"}`)))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("login: %d %s", rec.Code, rec.Body.String())
	}
	var cookie *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == sessionCookieName {
			cookie = c
			break
		}
	}
	if cookie == nil {
		t.Fatal("missing session cookie")
	}

	settingsGet := func() int {
		t.Helper()
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/settings", nil)
		req.AddCookie(cookie)
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("GET settings: %d %s", rec.Code, rec.Body.String())
		}
		var body struct {
			Values map[string]json.RawMessage `json:"values"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		raw, ok := body.Values["max_loss_equity_pct"]
		if !ok {
			t.Fatal("values.max_loss_equity_pct missing")
		}
		var n int
		if err := json.Unmarshal(raw, &n); err != nil {
			t.Fatalf("max_loss_equity_pct: %v", err)
		}
		return n
	}

	if got := settingsGet(); got != 10 {
		t.Fatalf("initial max_loss_equity_pct: got %d want 10", got)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/api/v1/settings", bytes.NewReader([]byte(`{"max_loss_equity_pct":25}`)))
	req.AddCookie(cookie)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("PUT settings: %d %s", rec.Code, rec.Body.String())
	}
	var putBody struct {
		Values map[string]json.RawMessage `json:"values"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &putBody); err != nil {
		t.Fatal(err)
	}
	raw, ok := putBody.Values["max_loss_equity_pct"]
	if !ok {
		t.Fatal("PUT response: max_loss_equity_pct missing")
	}
	var afterPut int
	if err := json.Unmarshal(raw, &afterPut); err != nil {
		t.Fatal(err)
	}
	if afterPut != 25 {
		t.Fatalf("PUT response max_loss_equity_pct: got %d want 25", afterPut)
	}

	if got := settingsGet(); got != 25 {
		t.Fatalf("GET after PUT max_loss_equity_pct: got %d want 25", got)
	}
}
