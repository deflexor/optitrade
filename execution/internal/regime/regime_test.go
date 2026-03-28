package regime

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/dfr/optitrade/execution/internal/config"
	"github.com/dfr/optitrade/execution/internal/market"
	"github.com/dfr/optitrade/execution/internal/state"
	"github.com/dfr/optitrade/execution/internal/state/sqlite"
)

func policyWithRegime(low, high string) *config.Policy {
	return &config.Policy{
		Version: "1.0.0",
		Limits: config.Limits{
			MaxLossPerTrade:            "1",
			MaxDailyLoss:               "1",
			MaxOpenPremiumAtRisk:       "1",
			MaxPortfolioDelta:          "1",
			MaxPortfolioVega:           "1",
			MaxOpenOrdersPerInstrument: 1,
			MaxTimeInTradeSeconds:      1,
		},
		Liquidity: config.Liquidity{MinTopSize: "1", MaxSpreadBps: 1},
		Regime: &config.Regime{
			Classifier:            "rules_v1",
			LowVolThresholdIndex:  low,
			HighVolThresholdIndex: high,
		},
		Playbooks: config.Playbooks{
			Low:    config.Playbook{AllowedStructures: []string{"iron_condor"}},
			Normal: config.Playbook{AllowedStructures: []string{"iron_condor"}},
			High:   config.Playbook{AllowedStructures: []string{"iron_condor"}},
		},
	}
}

func TestRulesV1Boundaries(t *testing.T) {
	t.Parallel()
	p := policyWithRegime("0.12", "0.35")
	now := time.Unix(1_700_000_000, 0)

	cases := []struct {
		vol  float64
		volT int64
		want Label
	}{
		{0.05, 1, LabelLow},
		{0.12, 1, LabelLow},
		{0.20, 1, LabelNormal},
		{0.35, 1, LabelHigh},
		{0.50, 1, LabelHigh},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("vol_%g", tc.vol), func(t *testing.T) {
			t.Parallel()
			h := new(Hysteresis)
			ev, err := EvaluateSnapshot(p, market.MarketSnapshot{
				Instrument: "BTC-PERPETUAL",
				VolIndex:   tc.vol,
				VolIndexTS: tc.volT,
			}, h, now)
			if err != nil {
				t.Fatal(err)
			}
			if ev.Outcome.Label != tc.want {
				t.Fatalf("got %q want %q", ev.Outcome.Label, tc.want)
			}
			if ev.Outcome.ClassifierVersion != ClassifierRulesV1 {
				t.Fatalf("classifier version: %q", ev.Outcome.ClassifierVersion)
			}
		})
	}
}

func TestHoldLastMissingVol(t *testing.T) {
	t.Parallel()
	p := policyWithRegime("0.12", "0.35")
	p.Regime.OnMissingVol = "hold_last"
	h := new(Hysteresis)
	t0 := time.Unix(1_700_000_000, 0)
	ev1, err := EvaluateSnapshot(p, market.MarketSnapshot{VolIndex: 0.99, VolIndexTS: 1}, h, t0)
	if err != nil {
		t.Fatal(err)
	}
	if ev1.Outcome.Label != LabelHigh {
		t.Fatalf("setup: %q", ev1.Outcome.Label)
	}
	ev2, err := EvaluateSnapshot(p, market.MarketSnapshot{VolIndex: 0, VolIndexTS: 0}, h, t0)
	if err != nil {
		t.Fatal(err)
	}
	if ev2.Outcome.Label != LabelHigh {
		t.Fatalf("hold_last: got %q", ev2.Outcome.Label)
	}
}

func TestMissingVolDefaultNormal(t *testing.T) {
	t.Parallel()
	p := policyWithRegime("0.12", "0.35")
	now := time.Unix(1_700_000_000, 0)
	h := new(Hysteresis)
	ev, err := EvaluateSnapshot(p, market.MarketSnapshot{
		Instrument: "BTC-PERPETUAL",
		VolIndex:   0.50,
		VolIndexTS: 0,
	}, h, now)
	if err != nil {
		t.Fatal(err)
	}
	if ev.Outcome.Label != LabelNormal {
		t.Fatalf("got %q", ev.Outcome.Label)
	}
}

func TestHysteresisDelaysFlip(t *testing.T) {
	t.Parallel()
	p := policyWithRegime("0.12", "0.35")
	n := 2
	p.Regime.HysteresisMinutes = &n
	h := new(Hysteresis)
	t0 := time.Unix(1_700_000_000, 0)
	// Start in normal (vol in band)
	ev1, err := EvaluateSnapshot(p, market.MarketSnapshot{
		VolIndex: 0.20, VolIndexTS: 1,
	}, h, t0)
	if err != nil {
		t.Fatal(err)
	}
	if ev1.Outcome.Label != LabelNormal {
		t.Fatalf("start: %q", ev1.Outcome.Label)
	}
	// Spike high — still normal until hysteresis elapses
	ev2, err := EvaluateSnapshot(p, market.MarketSnapshot{
		VolIndex: 0.99, VolIndexTS: 1,
	}, h, t0)
	if err != nil {
		t.Fatal(err)
	}
	if ev2.RawLabel != LabelHigh {
		t.Fatalf("raw: %q", ev2.RawLabel)
	}
	if ev2.Outcome.Label != LabelNormal {
		t.Fatalf("effective before wait: %q", ev2.Outcome.Label)
	}
	tAfter := t0.Add(2*time.Minute + time.Second)
	ev3, err := EvaluateSnapshot(p, market.MarketSnapshot{
		VolIndex: 0.99, VolIndexTS: 1,
	}, h, tAfter)
	if err != nil {
		t.Fatal(err)
	}
	if ev3.Outcome.Label != LabelHigh {
		t.Fatalf("effective after wait: %q", ev3.Outcome.Label)
	}
}

func TestPersistIfChangedInsertsOnceThenSkips(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	db, err := sqlite.Open(filepath.Join(dir, "r.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := sqlite.NewStore(db)
	ctx := context.Background()
	p := policyWithRegime("0.12", "0.35")
	h := new(Hysteresis)
	now := time.Unix(1_700_000_000, 0)
	ev, err := EvaluateSnapshot(p, market.MarketSnapshot{VolIndex: 0.2, VolIndexTS: 1}, h, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := PersistIfChanged(ctx, store, ev, now.UnixMilli()); err != nil {
		t.Fatal(err)
	}
	if err := PersistIfChanged(ctx, store, ev, now.UnixMilli()+1); err != nil {
		t.Fatal(err)
	}
	latest, err := store.LatestRegimeState(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if latest.Label != string(LabelNormal) {
		t.Fatalf("label %q", latest.Label)
	}
	var n int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(1) FROM regime_state`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("rows: %d want 1", n)
	}

	ev2, err := EvaluateSnapshot(p, market.MarketSnapshot{VolIndex: 0.05, VolIndexTS: 1}, h, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := PersistIfChanged(ctx, store, ev2, now.UnixMilli()+100); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(1) FROM regime_state`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("rows after change: %d want 2", n)
	}
}

func TestErrNoRegimeState(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	db, err := sqlite.Open(filepath.Join(dir, "empty.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := sqlite.NewStore(db)
	_, err = store.LatestRegimeState(context.Background())
	if !errors.Is(err, state.ErrNoRegimeState) {
		t.Fatalf("got %v", err)
	}
}
