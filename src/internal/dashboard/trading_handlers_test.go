package dashboard

import (
	"bytes"
	"context"
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

func TestTradingStatusAndMode(t *testing.T) {
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

	tradingGet := func() (accountStatus, botMode string, runnerRunning bool) {
		t.Helper()
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/trading/status", nil)
		req.AddCookie(cookie)
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("GET trading/status: %d %s", rec.Code, rec.Body.String())
		}
		var body struct {
			AccountStatus string `json:"account_status"`
			BotMode       string `json:"bot_mode"`
			RunnerRunning bool   `json:"runner_running"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		return body.AccountStatus, body.BotMode, body.RunnerRunning
	}

	a, m, r := tradingGet()
	if a != "active" || m != "manual" || r {
		t.Fatalf("initial status: account=%q bot=%q runner=%v want active, manual, false", a, m, r)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/api/v1/trading/mode", bytes.NewReader([]byte(`{"bot_mode":"paused"}`)))
	req.AddCookie(cookie)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("PUT trading/mode: %d %s", rec.Code, rec.Body.String())
	}
	var putBody struct {
		BotMode string `json:"bot_mode"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &putBody); err != nil {
		t.Fatal(err)
	}
	if putBody.BotMode != "paused" {
		t.Fatalf("PUT response bot_mode: got %q want paused", putBody.BotMode)
	}

	a, m, r = tradingGet()
	if a != "active" || m != "paused" || r {
		t.Fatalf("after PUT: account=%q bot=%q runner=%v want active, paused, false", a, m, r)
	}
}
