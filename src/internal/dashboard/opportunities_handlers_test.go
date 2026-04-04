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
	"time"

	"github.com/dfr/optitrade/src/internal/opportunities"
	"github.com/dfr/optitrade/src/internal/state/sqlite"
	"golang.org/x/crypto/bcrypt"
)

func TestOpportunitiesGet_pausedEmpty(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	db, err := sqlite.Open(filepath.Join(dir, "o.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	crypto := testOperatorSettingsCrypto(t)
	st := NewOperatorSettingsStore(db)
	ctx := context.Background()
	prov := "deribit"
	mainnet := false
	paused := "paused"
	_, err = st.Put(ctx, "op1", crypto, OperatorSettingsPatch{
		Provider:          &prov,
		DeribitUseMainnet: &mainnet,
		BotMode:           &paused,
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
	opp := NewOpportunityStore(db)
	srv := NewServer(Options{
		Logger:         log,
		Auth:           auth,
		Sessions:       NewSessionStore(db),
		SettingsCrypto: crypto,
		Settings:       st,
		Opportunities:  opp,
		OpportunitySnapshot: func(string) opportunities.Snapshot {
			return opportunities.Snapshot{
				UpdatedAtMs: 99,
				Rows: []opportunities.Row{
					{ID: "cand-1", StrategyName: "credit_spread", Status: opportunities.StatusCandidate},
				},
			}
		},
	})
	h := srv.Handler()
	cookie := loginCookie(t, h, "op1", "secret")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/opportunities", nil)
	req.AddCookie(cookie)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET opportunities: %d %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Paused bool                      `json:"paused"`
		Rows   []json.RawMessage         `json:"rows"`
		Hint   string                    `json:"resume_hint"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if !body.Paused || len(body.Rows) != 0 {
		t.Fatalf("want paused empty, got paused=%v rows=%d", body.Paused, len(body.Rows))
	}
	if body.Hint == "" {
		t.Fatal("expected resume_hint")
	}
}

func TestOpportunitiesGet_disabled(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	db, err := sqlite.Open(filepath.Join(dir, "d.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	crypto := testOperatorSettingsCrypto(t)
	st := NewOperatorSettingsStore(db)
	ctx := context.Background()
	prov := "deribit"
	mainnet := false
	disabled := "disabled"
	_, err = st.Put(ctx, "op1", crypto, OperatorSettingsPatch{
		Provider:          &prov,
		DeribitUseMainnet: &mainnet,
		AccountStatus:     &disabled,
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
	})
	h := srv.Handler()
	cookie := loginCookie(t, h, "op1", "secret")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/opportunities", nil)
	req.AddCookie(cookie)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET opportunities: %d %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Disabled bool `json:"disabled"`
		Rows     []any
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if !body.Disabled || len(body.Rows) != 0 {
		t.Fatalf("want disabled empty, got %+v", body)
	}
}

func TestOpportunitiesGet_mergesDBAndSnapshot(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	db, err := sqlite.Open(filepath.Join(dir, "m.db"))
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

	opp := NewOpportunityStore(db)
	row := opportunities.Row{
		ID:           "db-opening-1",
		StrategyName: "credit_spread",
		Status:       opportunities.StatusOpening,
		Legs:         []opportunities.LegQuote{{Instrument: "BTC-USD-260327-95000-P", Bid: 1, Ask: 2}},
		MaxProfit:    "0.1",
		MaxLoss:      "1",
		Recommend:    "open",
		Rationale:    "in flight",
		ExpectedEdge: "0.05",
		EdgeAfter:    0.04,
	}
	legsJ, metaJ, err := encodeOpportunityRow(row)
	if err != nil {
		t.Fatal(err)
	}
	if err := opp.Upsert(ctx, &OpportunityRecord{
		ID:           row.ID,
		Username:     "op1",
		Status:       string(opportunities.StatusOpening),
		StrategyName: row.StrategyName,
		LegsJSON:     legsJ,
		MetaJSON:     metaJ,
	}); err != nil {
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
		Opportunities:  opp,
		OpportunitySnapshot: func(string) opportunities.Snapshot {
			return opportunities.Snapshot{
				UpdatedAtMs: 12345,
				Rows: []opportunities.Row{
					{
						ID: "cand-btc", StrategyName: "credit_spread", Status: opportunities.StatusCandidate,
						Recommend: "open", Rationale: "ok", ExpectedEdge: "0.02", EdgeAfter: 0.01,
					},
				},
			}
		},
	})
	h := srv.Handler()
	cookie := loginCookie(t, h, "op1", "secret")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/opportunities", nil)
	req.AddCookie(cookie)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET opportunities: %d %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Paused        bool                `json:"paused"`
		UpdatedAtMs   int64               `json:"updated_at_ms"`
		Rows          []opportunities.Row `json:"rows"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Paused || len(body.Rows) != 2 {
		t.Fatalf("want 2 rows got %d body=%s", len(body.Rows), rec.Body.String())
	}
	if body.UpdatedAtMs != 12345 {
		t.Fatalf("updated_at_ms: %d", body.UpdatedAtMs)
	}
	ids := map[string]struct{}{}
	for _, r := range body.Rows {
		ids[r.ID] = struct{}{}
	}
	if _, ok := ids["db-opening-1"]; !ok {
		t.Fatal("missing db row")
	}
	if _, ok := ids["cand-btc"]; !ok {
		t.Fatal("missing snapshot row")
	}
}

func TestOpportunitiesStream_firstChunkHasRows(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	db, err := sqlite.Open(filepath.Join(dir, "sse.db"))
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
		OpportunitySnapshot: func(string) opportunities.Snapshot {
			return opportunities.Snapshot{UpdatedAtMs: 1, Rows: []opportunities.Row{
				{ID: "s1", StrategyName: "credit_spread", Status: opportunities.StatusCandidate},
			}}
		},
	})
	h := srv.Handler()
	cookie := loginCookie(t, h, "op1", "secret")

	reqCtx, cancel := context.WithCancel(context.Background())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/opportunities/stream", nil).WithContext(reqCtx)
	req.AddCookie(cookie)

	done := make(chan struct{})
	go func() {
		h.ServeHTTP(rec, req)
		close(done)
	}()

	deadline := time.After(2 * time.Second)
	for {
		select {
		case <-deadline:
			cancel()
			<-done
			t.Fatalf("timeout; body=%q", rec.Body.String())
		default:
			body := rec.Body.String()
			if strings.Contains(body, "data: ") && strings.Contains(body, `"rows"`) {
				cancel()
				<-done
				if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "text/event-stream") {
					t.Fatalf("content-type %q", ct)
				}
				return
			}
			time.Sleep(15 * time.Millisecond)
		}
	}
}

func loginCookie(t *testing.T, h http.Handler, user, pass string) *http.Cookie {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader([]byte(`{"username":"`+user+`","password":"`+pass+`"}`)))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("login: %d %s", rec.Code, rec.Body.String())
	}
	for _, c := range rec.Result().Cookies() {
		if c.Name == sessionCookieName {
			return c
		}
	}
	t.Fatal("missing session cookie")
	return nil
}
