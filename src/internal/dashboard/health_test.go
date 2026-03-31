package dashboard

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/dfr/optitrade/src/internal/deribit"
	"github.com/dfr/optitrade/src/internal/state/sqlite"
	"golang.org/x/crypto/bcrypt"
)

type healthExchange struct {
	t   int64
	err error
}

func (h healthExchange) GetAccountSummaries(ctx context.Context, p *deribit.GetAccountSummariesParams) ([]deribit.AccountSummary, error) {
	return nil, nil
}

func (h healthExchange) GetPositions(ctx context.Context, p *deribit.GetPositionsParams) ([]deribit.Position, error) {
	return nil, nil
}

func (h healthExchange) GetUserTrades(ctx context.Context, p deribit.GetUserTradesParams) ([]deribit.UserTrade, error) {
	return nil, nil
}

func (h healthExchange) GetServerTime(ctx context.Context) (int64, error) {
	return h.t, h.err
}

func TestHealthAPI_unauthenticated_noExchange(t *testing.T) {
	t.Parallel()
	srv := NewServer(Options{
		Logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
		Auth:     &DashboardAuthFile{Version: "1"},
		Sessions: nil,
		TestExchange: nil,
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Trading map[string]any `json:"trading"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Trading["exchange_reachable"] != false {
		t.Fatalf("exchange_reachable: %+v", body.Trading["exchange_reachable"])
	}
	if body.Trading["detail"] != "exchange_not_configured" {
		t.Fatalf("detail: %+v", body.Trading["detail"])
	}
}

func TestHealthAPI_exchangeUnreachable(t *testing.T) {
	t.Parallel()
	srv := NewServer(Options{
		Logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
		Auth:     &DashboardAuthFile{Version: "1"},
		Sessions: nil,
		TestExchange: healthExchange{err: errHealthTest},
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d", rec.Code)
	}
	var body struct {
		Trading map[string]any `json:"trading"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Trading["exchange_reachable"] != false {
		t.Fatal("expected unreachable")
	}
	if body.Trading["detail"] != "exchange_unreachable" {
		t.Fatalf("detail: %+v", body.Trading["detail"])
	}
}

func TestHealthAPI_exchangeReachable(t *testing.T) {
	t.Parallel()
	srv := NewServer(Options{
		Logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
		Auth:     &DashboardAuthFile{Version: "1"},
		Sessions: nil,
		TestExchange: healthExchange{t: 123},
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d", rec.Code)
	}
	var body struct {
		Trading map[string]any `json:"trading"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Trading["exchange_reachable"] != true {
		t.Fatal("expected reachable")
	}
}

func TestHealthAPI_requiresAuthForOverview(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	db, err := sqlite.Open(filepath.Join(dir, "s.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	hash, _ := bcrypt.GenerateFromPassword([]byte("pw"), bcrypt.MinCost)
	auth := &DashboardAuthFile{Version: "1", Users: []AuthUserRecord{{Username: "u", PasswordHash: string(hash)}}}
	srv := NewServer(Options{
		Logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
		Auth:     auth,
		Sessions: NewSessionStore(db),
		TestExchange: nil,
	})
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/overview", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("overview without session: %d", rec.Code)
	}
}

var errHealthTest = errHealthMarker{}

type errHealthMarker struct{}

func (errHealthMarker) Error() string { return "health test exchange error" }
