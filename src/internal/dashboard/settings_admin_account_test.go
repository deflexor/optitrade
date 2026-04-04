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

func TestSettingsPut_accountStatusAdminOnly(t *testing.T) {
	dir := t.TempDir()
	db, err := sqlite.Open(filepath.Join(dir, "admin.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	crypto := testOperatorSettingsCrypto(t)
	st := NewOperatorSettingsStore(db)
	ctx := context.Background()
	prov := "deribit"
	mainnet := false
	seed := func(username string) {
		t.Helper()
		_, err := st.Put(ctx, username, crypto, OperatorSettingsPatch{
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
	}
	seed("op1")
	seed("admin1")

	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	auth := &DashboardAuthFile{
		Version: "1",
		Users: []AuthUserRecord{
			{Username: "op1", PasswordHash: string(hash)},
			{Username: "admin1", PasswordHash: string(hash)},
		},
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := NewServer(Options{
		Logger:         log,
		Auth:           auth,
		Sessions:       NewSessionStore(db),
		SettingsCrypto: crypto,
		Settings:       st,
	})
	h := srv.Handler()

	login := func(user string) *http.Cookie {
		t.Helper()
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader([]byte(
			`{"username":"`+user+`","password":"secret"}`,
		)))
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("login %s: %d %s", user, rec.Code, rec.Body.String())
		}
		for _, c := range rec.Result().Cookies() {
			if c.Name == sessionCookieName {
				return c
			}
		}
		t.Fatal("missing session cookie")
		return nil
	}

	putAccount := func(cookie *http.Cookie, status string) *httptest.ResponseRecorder {
		t.Helper()
		body, _ := json.Marshal(map[string]string{"account_status": status})
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/api/v1/settings", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(cookie)
		h.ServeHTTP(rec, req)
		return rec
	}

	t.Run("forbidden_when_admin_env_unset", func(t *testing.T) {
		t.Setenv("OPTITRADE_DASHBOARD_ADMIN_USER", "")
		c := login("op1")
		rec := putAccount(c, "disabled")
		if rec.Code != http.StatusForbidden {
			t.Fatalf("want 403 got %d %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("forbidden_for_non_admin", func(t *testing.T) {
		t.Setenv("OPTITRADE_DASHBOARD_ADMIN_USER", "admin1")
		c := login("op1")
		rec := putAccount(c, "disabled")
		if rec.Code != http.StatusForbidden {
			t.Fatalf("want 403 got %d %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("ok_for_admin", func(t *testing.T) {
		t.Setenv("OPTITRADE_DASHBOARD_ADMIN_USER", "admin1")
		c := login("admin1")
		rec := putAccount(c, "disabled")
		if rec.Code != http.StatusOK {
			t.Fatalf("want 200 got %d %s", rec.Code, rec.Body.String())
		}
		row, err := st.GetDecrypting(ctx, "admin1", crypto)
		if err != nil || row == nil {
			t.Fatalf("get: %v %v", row, err)
		}
		if row.AccountStatus != "disabled" {
			t.Fatalf("account_status: %q", row.AccountStatus)
		}
	})
}
