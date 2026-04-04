package opportunities

import (
	"context"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/dfr/optitrade/src/internal/config"
	"github.com/dfr/optitrade/src/internal/deribit"
	"github.com/dfr/optitrade/src/internal/regime"
	"github.com/dfr/optitrade/src/internal/strategy"
)

func policyExamplePath(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
	return filepath.Join(root, "config", "examples", "policy.example.json")
}

func loadExamplePolicy(t *testing.T) *config.Policy {
	t.Helper()
	p, err := config.LoadFile(policyExamplePath(t))
	if err != nil {
		t.Fatal(err)
	}
	return p
}

type mapBooks map[string]deribit.OrderBook

func (m mapBooks) FetchBook(_ context.Context, inst string) (deribit.OrderBook, error) {
	b, ok := m[inst]
	if !ok {
		return deribit.OrderBook{}, nil
	}
	return b, nil
}

func TestSelector_costVeto_wideSpread(t *testing.T) {
	t.Parallel()
	raw := []byte(`{
  "version": "1.0.0",
  "limits": {
    "max_loss_per_trade": "500","max_daily_loss":"1500","max_open_premium_at_risk":"8000",
    "max_portfolio_delta":"15","max_portfolio_vega":"2",
    "max_open_orders_per_instrument": 10,"max_time_in_trade_seconds": 28800
  },
  "liquidity": { "min_top_size": "0.001", "max_spread_bps": 500 },
  "cost_model": { "taker_fee_bps": 3, "maker_fee_bps": 0, "slippage_bps": 2,
    "adverse_selection_bps_low": 1, "adverse_selection_bps_high": 4 },
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
	legs, err := strategy.VerticalPutCreditOKX("BTC", "20260327", 95000, 500)
	if err != nil {
		t.Fatal(err)
	}
	wide := deribit.OrderBook{
		Bids: []deribit.PriceLevel{{Price: 1, Amount: 2}},
		Asks: []deribit.PriceLevel{{Price: 10_000, Amount: 2}},
	}
	mb := mapBooks{
		legs[0].Instrument: wide,
		legs[1].Instrument: wide,
	}
	sel := &Selector{Policy: pol, Label: regime.LabelNormal, Books: mb}
	rows, err := sel.Evaluate(context.Background(), []CandidateSpec{
		{Base: "BTC", Expiry: "20260327", ShortStrike: 95000, Width: 500},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("rows %d", len(rows))
	}
	if rows[0].Recommend != "pass" || rows[0].Rationale != "non_positive_credit_at_conservative_quotes" {
		t.Fatalf("want non-positive credit pass, got %+v", rows[0])
	}
}

func TestSelector_costVeto_nonPositiveEdgeAfterCosts(t *testing.T) {
	t.Parallel()
	raw := []byte(`{
  "version": "1.0.0",
  "limits": {
    "max_loss_per_trade": "500","max_daily_loss":"1500","max_open_premium_at_risk":"8000",
    "max_portfolio_delta":"15","max_portfolio_vega":"2",
    "max_open_orders_per_instrument": 10,"max_time_in_trade_seconds": 28800
  },
  "liquidity": { "min_top_size": "0.001", "max_spread_bps": 500 },
  "cost_model": { "taker_fee_bps": 5000, "maker_fee_bps": 0, "slippage_bps": 5000,
    "adverse_selection_bps_low": 5000, "adverse_selection_bps_high": 5000 },
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
	legs, err := strategy.VerticalPutCreditOKX("BTC", "20260327", 95000, 500)
	if err != nil {
		t.Fatal(err)
	}
	mb := mapBooks{
		legs[0].Instrument: tightSpreadBook(0.05, 0.001),
		legs[1].Instrument: tightSpreadBook(0.04, 0.001),
	}
	sel := &Selector{Policy: pol, Label: regime.LabelNormal, Books: mb}
	rows, err := sel.Evaluate(context.Background(), []CandidateSpec{
		{Base: "BTC", Expiry: "20260327", ShortStrike: 95000, Width: 500},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("rows %d", len(rows))
	}
	if rows[0].Recommend != "pass" || !strings.HasPrefix(rows[0].Rationale, "cost veto:") {
		t.Fatalf("want cost veto, got %+v", rows[0])
	}
}

func TestSelector_recommendOpen_tightBooks(t *testing.T) {
	t.Parallel()
	pol := loadExamplePolicy(t)
	legs, err := strategy.VerticalPutCreditOKX("BTC", "20260327", 95000, 500)
	if err != nil {
		t.Fatal(err)
	}
	shortMid, longMid := 0.08, 0.03
	mb := mapBooks{
		legs[0].Instrument: tightSpreadBook(shortMid, 0.002),
		legs[1].Instrument: tightSpreadBook(longMid, 0.002),
	}
	sel := &Selector{Policy: pol, Label: regime.LabelNormal, Books: mb}
	rows, err := sel.Evaluate(context.Background(), []CandidateSpec{
		{Base: "BTC", Expiry: "20260327", ShortStrike: 95000, Width: 500},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("rows %d", len(rows))
	}
	if rows[0].Recommend != "open" {
		t.Fatalf("want open, got %+v", rows[0])
	}
}

func TestSelector_riskVeto_maxLossOverPolicy(t *testing.T) {
	t.Parallel()
	raw := []byte(`{
  "version": "1.0.0",
  "limits": {
    "max_loss_per_trade": "0.05","max_daily_loss":"1500","max_open_premium_at_risk":"8000",
    "max_portfolio_delta":"15","max_portfolio_vega":"2",
    "max_open_orders_per_instrument": 10,"max_time_in_trade_seconds": 28800
  },
  "liquidity": { "min_top_size": "0.001", "max_spread_bps": 500 },
  "cost_model": { "taker_fee_bps": 1, "maker_fee_bps": 0, "slippage_bps": 1,
    "adverse_selection_bps_low": 1, "adverse_selection_bps_high": 2 },
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
	legs, err := strategy.VerticalPutCreditOKX("BTC", "20260327", 95000, 500)
	if err != nil {
		t.Fatal(err)
	}
	shortMid, longMid := 0.08, 0.03
	mb := mapBooks{
		legs[0].Instrument: tightSpreadBook(shortMid, 0.002),
		legs[1].Instrument: tightSpreadBook(longMid, 0.002),
	}
	sel := &Selector{Policy: pol, Label: regime.LabelNormal, Books: mb}
	rows, err := sel.Evaluate(context.Background(), []CandidateSpec{
		{Base: "BTC", Expiry: "20260327", ShortStrike: 95000, Width: 500},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Recommend != "pass" {
		t.Fatalf("want risk pass veto, got %+v", rows)
	}
	if rows[0].Rationale != "risk veto (see audit)" {
		t.Fatalf("rationale: %q", rows[0].Rationale)
	}
}

func TestSelector_ranksByEdgeAfterCosts(t *testing.T) {
	t.Parallel()
	pol := loadExamplePolicy(t)
	legsA, _ := strategy.VerticalPutCreditOKX("BTC", "20260327", 95000, 500)
	legsB, _ := strategy.VerticalPutCreditOKX("BTC", "20260327", 94000, 500)
	mb := mapBooks{
		legsA[0].Instrument: tightSpreadBook(0.10, 0.002),
		legsA[1].Instrument: tightSpreadBook(0.02, 0.002),
		legsB[0].Instrument: tightSpreadBook(0.05, 0.002),
		legsB[1].Instrument: tightSpreadBook(0.02, 0.002),
	}
	sel := &Selector{Policy: pol, Label: regime.LabelNormal, Books: mb}
	rows, err := sel.Evaluate(context.Background(), []CandidateSpec{
		{Base: "BTC", Expiry: "20260327", ShortStrike: 95000, Width: 500},
		{Base: "BTC", Expiry: "20260327", ShortStrike: 94000, Width: 500},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("rows %d", len(rows))
	}
	if rows[0].EdgeAfter < rows[1].EdgeAfter {
		t.Fatalf("sort order: first=%v second=%v", rows[0].EdgeAfter, rows[1].EdgeAfter)
	}
}

func tightSpreadBook(mid float64, spread float64) deribit.OrderBook {
	h := spread / 2
	return deribit.OrderBook{
		Bids: []deribit.PriceLevel{{Price: mid - h, Amount: 2}},
		Asks: []deribit.PriceLevel{{Price: mid + h, Amount: 2}},
	}
}
