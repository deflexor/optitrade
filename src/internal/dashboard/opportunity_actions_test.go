package dashboard

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dfr/optitrade/src/internal/opportunities"
	"github.com/dfr/optitrade/src/internal/state/sqlite"
	"golang.org/x/crypto/bcrypt"
)

func TestOpportunityOpen_paused(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	db, err := sqlite.Open(filepath.Join(dir, "p.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	crypto := testOperatorSettingsCrypto(t)
	st := NewOperatorSettingsStore(db)
	ctx := context.Background()
	prov := "okx"
	paused := "paused"
	_, err = st.Put(ctx, "op1", crypto, OperatorSettingsPatch{
		Provider: &prov,
		BotMode:  &paused,
		Secrets: map[string]string{
			"okx_api_key":        "k",
			"okx_secret_key":     strings.Repeat("s", 32),
			"okx_passphrase":     "p",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	auth := &DashboardAuthFile{Version: "1", Users: []AuthUserRecord{{Username: "op1", PasswordHash: string(hash)}}}
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
			return opportunities.Snapshot{Rows: []opportunities.Row{{
				ID: "c1", StrategyName: "credit_spread", Status: opportunities.StatusCandidate,
				Legs: []opportunities.LegQuote{
					{Instrument: "A-P"},
					{Instrument: "B-P"},
				},
				Recommend: "open", Rationale: "x", MaxProfit: "1", MaxLoss: "2", ExpectedEdge: "0", EdgeAfter: 0,
			}}}
		},
	})
	h := srv.Handler()
	cookie := loginCookie(t, h, "op1", "secret")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/opportunities/c1/open", nil)
	req.AddCookie(cookie)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("want 409 got %d %s", rec.Code, rec.Body.String())
	}
}

func TestOpportunityOpen_okxBatchPersistsOpening(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	db, err := sqlite.Open(filepath.Join(dir, "o.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	okxSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v5/trade/batch-orders" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":"0","data":[{"ordId":"o1","sCode":"0"},{"ordId":"o2","sCode":"0"}]}`))
	}))
	t.Cleanup(okxSrv.Close)

	crypto := testOperatorSettingsCrypto(t)
	st := NewOperatorSettingsStore(db)
	ctx := context.Background()
	prov := "okx"
	_, err = st.Put(ctx, "op1", crypto, OperatorSettingsPatch{
		Provider: &prov,
		Secrets: map[string]string{
			"okx_api_key":        "k",
			"okx_secret_key":     strings.Repeat("s", 32),
			"okx_passphrase":     "p",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	auth := &DashboardAuthFile{Version: "1", Users: []AuthUserRecord{{Username: "op1", PasswordHash: string(hash)}}}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	opp := NewOpportunityStore(db)
	srv := NewServer(Options{
		Logger:         log,
		Auth:           auth,
		Sessions:       NewSessionStore(db),
		SettingsCrypto: crypto,
		Settings:       st,
		Opportunities:  opp,
		OKXExecBaseURL: strings.TrimSuffix(okxSrv.URL, "/"),
		OKXExecHTTP:    okxSrv.Client(),
		OpportunitySnapshot: func(string) opportunities.Snapshot {
			return opportunities.Snapshot{Rows: []opportunities.Row{{
				ID: "sp1", StrategyName: "credit_spread", Status: opportunities.StatusCandidate,
				Legs: []opportunities.LegQuote{
					{Instrument: "BTC-USD-260327-90000-P"},
					{Instrument: "BTC-USD-260327-91000-P"},
				},
				Recommend: "open", Rationale: "x", MaxProfit: "1", MaxLoss: "2", ExpectedEdge: "0", EdgeAfter: 0,
			}}}
		},
	})
	h := srv.Handler()
	cookie := loginCookie(t, h, "op1", "secret")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/opportunities/sp1/open", nil)
	req.AddCookie(cookie)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("open: %d %s", rec.Code, rec.Body.String())
	}
	got, err := opp.Get(ctx, "sp1", "op1")
	if err != nil || got == nil {
		t.Fatalf("get: %v %v", got, err)
	}
	if got.Status != string(opportunities.StatusOpening) {
		t.Fatalf("status %q", got.Status)
	}
	var meta opportunityMetaJSON
	if err := json.Unmarshal([]byte(got.MetaJSON), &meta); err != nil {
		t.Fatal(err)
	}
	if len(meta.OrderIDs) != 2 || meta.OrderIDs[0] != "o1" {
		t.Fatalf("meta %+v", meta)
	}
}

func TestOpportunityCancel_wrongState(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	db, err := sqlite.Open(filepath.Join(dir, "c.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	crypto := testOperatorSettingsCrypto(t)
	st := NewOperatorSettingsStore(db)
	ctx := context.Background()
	prov := "okx"
	_, err = st.Put(ctx, "op1", crypto, OperatorSettingsPatch{
		Provider: &prov,
		Secrets: map[string]string{
			"okx_api_key":        "k",
			"okx_secret_key":     strings.Repeat("s", 32),
			"okx_passphrase":     "p",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	opp := NewOpportunityStore(db)
	row := opportunities.Row{
		ID: "x1", StrategyName: "credit_spread", Status: opportunities.StatusActive,
		Legs: []opportunities.LegQuote{
			{Instrument: "A-P"},
			{Instrument: "B-P"},
		},
		MaxProfit: "1", MaxLoss: "2", Recommend: "open", Rationale: "y", ExpectedEdge: "0", EdgeAfter: 0,
	}
	legsJ, metaJ, err := encodeOpportunityRowPersist(row, &opportunityExecPersist{
		OrderIDs: []string{"1", "2"},
		LegSides: []legSideMeta{{Instrument: "A-P", Side: "sell"}, {Instrument: "B-P", Side: "buy"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := opp.Upsert(ctx, &OpportunityRecord{
		ID: row.ID, Username: "op1", Status: string(opportunities.StatusActive),
		StrategyName: row.StrategyName, LegsJSON: legsJ, MetaJSON: metaJ,
	}); err != nil {
		t.Fatal(err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	auth := &DashboardAuthFile{Version: "1", Users: []AuthUserRecord{{Username: "op1", PasswordHash: string(hash)}}}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := NewServer(Options{
		Logger:         log,
		Auth:           auth,
		Sessions:       NewSessionStore(db),
		SettingsCrypto: crypto,
		Settings:       st,
		Opportunities:  opp,
	})
	h := srv.Handler()
	cookie := loginCookie(t, h, "op1", "secret")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/opportunities/x1/cancel", nil)
	req.AddCookie(cookie)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("want 409 got %d %s", rec.Code, rec.Body.String())
	}
}
