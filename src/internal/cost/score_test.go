package cost

import (
	"context"
	"log/slog"
	"math"
	"path/filepath"
	"testing"
	"time"

	"github.com/dfr/optitrade/src/internal/audit"
	"github.com/dfr/optitrade/src/internal/config"
	"github.com/dfr/optitrade/src/internal/deribit"
	"github.com/dfr/optitrade/src/internal/market"
	"github.com/dfr/optitrade/src/internal/regime"
	"github.com/dfr/optitrade/src/internal/state/sqlite"
)

func TestScoreCandidate_tightBookPositiveEdge_ok(t *testing.T) {
	t.Parallel()
	pol := mustPolicy(t)
	book := deribit.OrderBook{
		Bids: []deribit.PriceLevel{{Price: 100, Amount: 2}},
		Asks: []deribit.PriceLevel{{Price: 100.02, Amount: 2}},
	}
	cand := CandidateInput{ExpectedEdge: "0.01", QuoteCurrency: "BTC", PrimaryMidValid: true, PrimaryMid: 100}
	ok, veto, bd := ScoreCandidate(pol, regime.LabelNormal, cand, []deribit.OrderBook{book}, nil, IVSanityOptions{})
	if !ok || veto != "" {
		t.Fatalf("want ok, got ok=%v veto=%q bd=%+v", ok, veto, bd)
	}
	if bd.EdgeAfterCosts <= 0 {
		t.Fatalf("edge after costs: %+v", bd)
	}
}

func TestScoreCandidate_highCosts_veto(t *testing.T) {
	t.Parallel()
	raw := []byte(`{
  "version": "1.0.0",
  "limits": {
    "max_loss_per_trade": "1","max_daily_loss":"1","max_open_premium_at_risk":"1",
    "max_portfolio_delta":"0","max_portfolio_vega":"0",
    "max_open_orders_per_instrument": 1,"max_time_in_trade_seconds": 1
  },
  "liquidity": { "min_top_size": "1", "max_spread_bps": 500 },
  "cost_model": { "taker_fee_bps": 50, "maker_fee_bps": 0, "slippage_bps": 50,
    "adverse_selection_bps_low": 20, "adverse_selection_bps_high": 80 },
  "playbooks": {
    "low": { "allowed_structures": ["credit_spread"] },
    "normal": { "allowed_structures": ["credit_spread"] },
    "high": { "allowed_structures": ["credit_spread"] }
  }
}`)
	pol, err := config.LoadBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	// Extreme touch yields half-spread bps ~10k; with fees, total haircut exceeds 100% of edge.
	book := deribit.OrderBook{
		Bids: []deribit.PriceLevel{{Price: 1, Amount: 2}},
		Asks: []deribit.PriceLevel{{Price: 10_000, Amount: 2}},
	}
	cand := CandidateInput{ExpectedEdge: "1", QuoteCurrency: "BTC"}
	ok, veto, bd := ScoreCandidate(pol, regime.LabelHigh, cand, []deribit.OrderBook{book}, nil, IVSanityOptions{})
	if ok || veto != "non_positive_edge" {
		t.Fatalf("want veto non_positive_edge, ok=%v veto=%q bd=%+v", ok, veto, bd)
	}
	_ = bd
}

func TestScoreCandidate_ivLag_veto(t *testing.T) {
	t.Parallel()
	pol := mustPolicy(t)
	book := deribit.OrderBook{
		Bids: []deribit.PriceLevel{{Price: 50_000, Amount: 1}},
		Asks: []deribit.PriceLevel{{Price: 50_010, Amount: 1}},
	}
	now := time.Unix(1_700_000_000, 0)
	snap := market.MarketSnapshot{
		Timestamp:   now,
		Instrument:  "BTC-PERPETUAL",
		Book:        book,
		BookLocalTS: now,
		VolIndex:    0.5,
		VolIndexTS:  now.Add(-10 * time.Minute).UnixMilli(),
	}
	cand := CandidateInput{ExpectedEdge: "0.02", QuoteCurrency: "BTC", PrimaryMidValid: true, PrimaryMid: 50_005}
	iv := IVSanityOptions{UseIVQuotes: true, MaxVolBookLag: time.Minute}
	ok, veto, _ := ScoreCandidate(pol, regime.LabelNormal, cand, []deribit.OrderBook{book}, &snap, iv)
	if ok || veto != "iv_stale" {
		t.Fatalf("want iv_stale, ok=%v veto=%q", ok, veto)
	}
}

func TestScoreCandidate_tableBpsGolden(t *testing.T) {
	t.Parallel()
	pol := mustPolicy(t)
	// spread 1 on mid 100 -> 100 bps full -> 50 half
	book := deribit.OrderBook{
		Bids: []deribit.PriceLevel{{Price: 100, Amount: 2}},
		Asks: []deribit.PriceLevel{{Price: 101, Amount: 2}},
	}
	cand := CandidateInput{ExpectedEdge: "1", QuoteCurrency: "BTC"}
	ok, veto, bd := ScoreCandidate(pol, regime.LabelLow, cand, []deribit.OrderBook{book}, nil, IVSanityOptions{})
	if !ok {
		t.Fatalf("unexpected veto %q %+v", veto, bd)
	}
	// policy example: taker 3, slip 2, adverse low 1; half-spread from 100/101 mid 100.5
	touch, half, okb := maxHalfSpreadBps([]deribit.OrderBook{book})
	if !okb {
		t.Fatal("book")
	}
	_ = touch
	wantBps := float64(3+2+1) + half
	if math.Abs(bd.TotalHaircutBps-wantBps) > 1e-9 {
		t.Fatalf("total bps: got %v want %v", bd.TotalHaircutBps, wantBps)
	}
	wantAfter := 1.0 - 1.0*(wantBps/10_000)
	if math.Abs(bd.EdgeAfterCosts-wantAfter) > 1e-9 {
		t.Fatalf("edge after: got %v want %v", bd.EdgeAfterCosts, wantAfter)
	}
}

func TestLogCostVeto_persists(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	db, err := sqlite.Open(filepath.Join(dir, "a.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	dl := audit.NewDecisionLogger(sqlite.NewStore(db), slog.Default(), audit.LoggerOptions{})
	ctx := context.Background()
	bd := CostBreakdown{CostModelVersion: CostModelVersion, ExpectedEdge: 0.01}
	if err := LogCostVeto(ctx, dl, "corr-12345678", nil, regime.LabelNormal, "non_positive_edge", bd); err != nil {
		t.Fatal(err)
	}
}

func mustPolicy(t *testing.T) *config.Policy {
	t.Helper()
	p, err := config.LoadFile("../../../config/examples/policy.example.json")
	if err != nil {
		t.Fatal(err)
	}
	return p
}
