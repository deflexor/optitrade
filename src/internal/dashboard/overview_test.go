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
	"strings"
	"testing"

	"github.com/dfr/optitrade/src/internal/deribit"
	"github.com/dfr/optitrade/src/internal/state/sqlite"
	"golang.org/x/crypto/bcrypt"
)

type stubExchange struct{}

func (stubExchange) GetAccountSummaries(ctx context.Context, p *deribit.GetAccountSummariesParams) ([]deribit.AccountSummary, error) {
	eq, bal := 123.45, 100.0
	return []deribit.AccountSummary{{Currency: "BTC", Equity: &eq, Balance: &bal}}, nil
}

func (stubExchange) GetPositions(ctx context.Context, p *deribit.GetPositionsParams) ([]deribit.Position, error) {
	return nil, nil
}

func (stubExchange) GetUserTrades(ctx context.Context, p deribit.GetUserTradesParams) ([]deribit.UserTrade, error) {
	return nil, nil
}

func (stubExchange) GetServerTime(ctx context.Context) (int64, error) {
	return 0, nil
}

func TestOverviewJSONShape(t *testing.T) {
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
		Exchange: stubExchange{},
	})
	h := srv.Handler()

	jar := make(map[string]string)
	// login
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"username":"op","password":"x"}`))
	h.ServeHTTP(rec, req)
	for _, c := range rec.Result().Cookies() {
		if c.Name == sessionCookieName {
			jar[sessionCookieName] = c.Value
		}
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/overview", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: jar[sessionCookieName]})
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("overview: %d %s", rec.Code, rec.Body.String())
	}

	var body map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	var mood struct {
		Available   bool   `json:"available"`
		Explanation string `json:"explanation"`
	}
	if err := json.Unmarshal(body["market_mood"], &mood); err != nil {
		t.Fatal(err)
	}
	if mood.Available {
		t.Fatal("expected mood.available false for stub")
	}
	if strings.Contains(mood.Explanation, "strategy modules not wired") {
		t.Fatalf("market_mood.explanation must not use internal placeholder: %q", mood.Explanation)
	}
	if mood.Explanation == "" {
		t.Fatal("expected non-empty mood explanation when unavailable")
	}
	var strat struct {
		WinRateDefined bool `json:"win_rate_defined"`
		Available      bool `json:"available"`
	}
	if err := json.Unmarshal(body["strategy"], &strat); err != nil {
		t.Fatal(err)
	}
	if strat.Available || strat.WinRateDefined {
		t.Fatalf("strategy: %+v", strat)
	}
	var acct map[string]json.RawMessage
	if err := json.Unmarshal(body["account"], &acct); err != nil {
		t.Fatal(err)
	}
	var equity string
	if err := json.Unmarshal(acct["equity"], &equity); err != nil {
		t.Fatal(err)
	}
	if equity == "" {
		t.Fatal("expected equity string")
	}
}
