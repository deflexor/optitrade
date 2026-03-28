package risk

import (
	"context"
	"log/slog"
	"math/big"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/dfr/optitrade/execution/internal/audit"
	"github.com/dfr/optitrade/execution/internal/config"
	"github.com/dfr/optitrade/execution/internal/deribit"
	"github.com/dfr/optitrade/execution/internal/state/sqlite"
)

func policyFile(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// .../execution/internal/risk/*_test.go -> worktree root is ../..
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
	return filepath.Join(root, "config", "examples", "policy.example.json")
}

func loadPolicy(t *testing.T) *config.Policy {
	t.Helper()
	p, err := config.LoadFile(policyFile(t))
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func gateFalse(t *testing.T, gates map[string]any, key string) bool {
	t.Helper()
	v, ok := gates[key].(bool)
	return ok && !v
}

func TestBuildPortfolioSnapshotSumsGreeks(t *testing.T) {
	t.Parallel()
	d1, d2 := 0.05, 0.12
	v1 := 0.001
	m1, m2 := 40000.0, 50000.0
	im := 100.0
	positions := []deribit.Position{
		{InstrumentName: "A", Size: 1, Delta: &d1, Vega: &v1, MarkPrice: &m1},
		{InstrumentName: "B", Size: -2, Delta: &d2, MarkPrice: &m2, InitialMargin: &im},
	}
	snap := BuildPortfolioSnapshot(positions, nil, nil, nil)
	rf := func(f float64) *big.Rat {
		r := new(big.Rat)
		r.SetFloat64(f)
		return r
	}
	want := new(big.Rat).Add(rf(0.05), rf(0.12))
	if snap.DeltaTotal.Cmp(want) != 0 {
		t.Fatalf("delta: %v want %v", snap.DeltaTotal.FloatString(12), want.FloatString(12))
	}
	if snap.OpenOrdersByInstrument != nil && len(snap.OpenOrdersByInstrument) != 0 {
		t.Fatalf("orders map: %v", snap.OpenOrdersByInstrument)
	}
}

func TestEngineDeltaVetoAudits(t *testing.T) {
	t.Parallel()
	policy := loadPolicy(t)
	dir := t.TempDir()
	db, err := sqlite.Open(filepath.Join(dir, "r.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	dl := audit.NewDecisionLogger(sqlite.NewStore(db), slog.Default(), audit.LoggerOptions{})
	eng := NewEngine(policy, dl)

	d := 1.0
	pos := []deribit.Position{{InstrumentName: "X", Size: 1, Delta: &d, MarkPrice: &d}}
	in := PreTradeInput{
		CorrelationID:    "corr-delta-veto-1",
		RegimeLabel:      "normal",
		CostModelVersion: "cm-v1",
		Positions:        pos,
		Candidate: CandidateRisk{
			ID:           "c1",
			StrategyID:   "s-new",
			MaxLossQuote: "100",
			Instruments:  []string{"X"},
		},
		Now:           time.Unix(1_800_000_000, 0),
		CumulativePnL: big.NewRat(0, 1),
		DailyTracker:  &DailyLossTracker{},
	}

	allowed, err := eng.Check(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	if allowed {
		t.Fatal("expected veto")
	}
	ok, gates := eng.EvaluateDryRun(in)
	if ok || !gateFalse(t, gates, "limit_delta") {
		t.Fatalf("dry run: ok=%v gates=%v", ok, gates)
	}
}

// Baseline daily-loss veto; see TestSC002_* for explicit SC-002 / session-boundary acceptance.
func TestEngineDailyLossVeto(t *testing.T) {
	t.Parallel()
	policy := loadPolicy(t)
	dir := t.TempDir()
	db, err := sqlite.Open(filepath.Join(dir, "d.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	dl := audit.NewDecisionLogger(sqlite.NewStore(db), slog.Default(), audit.LoggerOptions{})
	eng := NewEngine(policy, dl)

	now := time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC)
	tr := &DailyLossTracker{}
	start := big.NewRat(0, 1)
	_ = tr.SessionLoss(now, start)
	pnl := big.NewRat(-2000, 1)

	in := PreTradeInput{
		CorrelationID:    "corr-daily-veto-xxxxx",
		RegimeLabel:      "normal",
		CostModelVersion: "cm-v1",
		Positions:        nil,
		Candidate: CandidateRisk{
			MaxLossQuote: "1",
			Instruments:  nil,
		},
		Now:           now,
		CumulativePnL: pnl,
		DailyTracker:  tr,
	}
	allowed, err := eng.Check(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	if allowed {
		t.Fatal("expected daily loss veto")
	}
	ok, gates := NewEngine(policy, nil).EvaluateDryRun(in)
	if ok || !gateFalse(t, gates, "daily_loss") {
		t.Fatalf("gates=%v", gates)
	}
}

func TestEngineTimeInTradeVeto(t *testing.T) {
	t.Parallel()
	policy := loadPolicy(t)
	eng := NewEngine(policy, nil)
	opened := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	now := opened.Add(time.Duration(policy.Limits.MaxTimeInTradeSeconds) * time.Second).Add(time.Second)

	in := PreTradeInput{
		CorrelationID: "corr-time-veto-xxxxxx",
		RegimeLabel:   "normal",
		Candidate: CandidateRisk{
			StrategyID:   "iron-1",
			MaxLossQuote: "1",
		},
		StrategyOpenedAt: map[string]time.Time{"iron-1": opened},
		Now:              now,
		CumulativePnL:    big.NewRat(0, 1),
		DailyTracker:     &DailyLossTracker{},
	}
	ok, gates := eng.EvaluateDryRun(in)
	if ok || !gateFalse(t, gates, "time_in_trade") {
		t.Fatalf("expected time veto: ok=%v gates=%v", ok, gates)
	}
}

func TestEngineOpenOrdersVeto(t *testing.T) {
	t.Parallel()
	policy := loadPolicy(t)
	eng := NewEngine(policy, nil)
	inst := "BTC-PERP"
	orders := []deribit.OpenOrder{
		{InstrumentName: inst},
		{InstrumentName: inst},
	}
	in := PreTradeInput{
		CorrelationID: "corr-orders-veto-xxxxxx",
		RegimeLabel:   "normal",
		OpenOrders:    orders,
		Candidate: CandidateRisk{
			MaxLossQuote: "1",
			Instruments:  []string{inst},
		},
		Now:           time.Now(),
		CumulativePnL: big.NewRat(0, 1),
		DailyTracker:  &DailyLossTracker{},
	}
	ok, gates := eng.EvaluateDryRun(in)
	if ok || !gateFalse(t, gates, "limit_open_orders") {
		t.Fatalf("expected open order cap veto: ok=%v gates=%v", ok, gates)
	}
}

func TestEnginePerTradeMaxLossVeto(t *testing.T) {
	t.Parallel()
	policy := loadPolicy(t)
	eng := NewEngine(policy, nil)
	in := PreTradeInput{
		CorrelationID: "corr-trade-veto-xxxxxx",
		RegimeLabel:   "normal",
		Candidate: CandidateRisk{
			MaxLossQuote: "1000",
			FeesQuote:    "1",
		},
		Now:           time.Now(),
		CumulativePnL: big.NewRat(0, 1),
		DailyTracker:  &DailyLossTracker{},
	}
	ok, gates := eng.EvaluateDryRun(in)
	if ok || !gateFalse(t, gates, "per_trade_max_loss") {
		t.Fatalf("expected per-trade veto: ok=%v gates=%v", ok, gates)
	}
}

func TestEngineAllPass(t *testing.T) {
	t.Parallel()
	policy := loadPolicy(t)
	eng := NewEngine(policy, nil)
	d, v, m := 0.01, 0.001, 1000.0
	im := 50.0
	pos := []deribit.Position{
		{InstrumentName: "A", Size: 1, Delta: &d, Vega: &v, MarkPrice: &m, InitialMargin: &im},
	}
	in := PreTradeInput{
		CorrelationID: "corr-all-pass-xxxxxx",
		RegimeLabel:   "normal",
		Positions:     pos,
		OpenOrders:    nil,
		Candidate: CandidateRisk{
			MaxLossQuote: "10",
			FeesQuote:    "1",
			Instruments:  []string{"A"},
			StrategyID:   "new-strat",
		},
		Now:           time.Now(),
		CumulativePnL: big.NewRat(0, 1),
		DailyTracker:  &DailyLossTracker{},
	}
	ok, gates := eng.EvaluateDryRun(in)
	if !ok {
		t.Fatalf("expected pass gates=%v", gates)
	}
}

// SC-002 (spec): daily loss cap blocks further risk-increasing activity until session reset.
// See [DailyLossTracker] for UTC calendar-day boundary alignment with WP09.
func TestSC002_DailyLossCapBlocksSubsequentCheck(t *testing.T) {
	t.Parallel()
	policy := loadPolicy(t)
	eng := NewEngine(policy, nil)
	now := time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC)
	tr := &DailyLossTracker{}
	tr.SessionLoss(now, big.NewRat(0, 1))
	pnl := big.NewRat(-2000, 1)
	in := PreTradeInput{
		CorrelationID:    "sc002-corr-xxxxxx",
		RegimeLabel:      "normal",
		CostModelVersion: "cm-v1",
		Positions:        nil,
		Candidate: CandidateRisk{
			MaxLossQuote: "1",
			Instruments:  nil,
		},
		Now:           now,
		CumulativePnL: pnl,
		DailyTracker:  tr,
	}
	if ok, _ := eng.EvaluateDryRun(in); ok {
		t.Fatal("expected first check to fail daily_loss gate")
	}
	if ok, _ := eng.EvaluateDryRun(in); ok {
		t.Fatal("expected subsequent evaluations to remain blocked same session")
	}
}

func TestSC002_SessionBoundaryResetsDailyLossGate(t *testing.T) {
	t.Parallel()
	policy := loadPolicy(t)
	eng := NewEngine(policy, nil)
	tr := &DailyLossTracker{}
	day1 := time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC)
	_ = tr.SessionLoss(day1, big.NewRat(0, 1))
	pnl := big.NewRat(-2000, 1)
	in1 := PreTradeInput{
		CorrelationID:    "sc002b-corr-xxxxx",
		RegimeLabel:      "normal",
		CostModelVersion: "cm-v1",
		Candidate:        CandidateRisk{MaxLossQuote: "1"},
		Now:              day1,
		CumulativePnL:    pnl,
		DailyTracker:     tr,
	}
	ok1, g1 := eng.EvaluateDryRun(in1)
	if ok1 || g1["daily_loss"].(bool) {
		t.Fatalf("expected loss cap on day1: ok=%v daily_loss=%v", ok1, g1["daily_loss"])
	}
	day2 := time.Date(2026, 3, 29, 12, 0, 0, 0, time.UTC)
	in2 := PreTradeInput{
		CorrelationID:    "sc002b2-corr-xxxxx",
		RegimeLabel:      "normal",
		CostModelVersion: "cm-v1",
		Candidate:        CandidateRisk{MaxLossQuote: "1"},
		Now:              day2,
		CumulativePnL:    pnl,
		DailyTracker:     tr,
	}
	ok2, g2 := eng.EvaluateDryRun(in2)
	if !ok2 || !g2["daily_loss"].(bool) {
		t.Fatalf("expected new UTC day to reset session PnL baseline: ok=%v gates=%v", ok2, g2)
	}
}

