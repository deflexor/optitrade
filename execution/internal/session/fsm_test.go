package session

import (
	"errors"
	"testing"
	"time"

	"github.com/dfr/optitrade/execution/internal/config"
	"github.com/dfr/optitrade/execution/internal/market"
)

func TestFSM_MarketStaleMovesToProtectiveFlatten(t *testing.T) {
	t.Parallel()
	f := NewFSM()
	f.ObserveMarket(market.StaleBook, nil)
	if f.State() != StateProtectiveFlatten {
		t.Fatalf("state=%q want protective_flatten", f.State())
	}
	if f.LastReason() != reasonMarketQuality {
		t.Fatalf("reason=%q", f.LastReason())
	}
}

func TestFSM_WSDownGraceTriggersProtective(t *testing.T) {
	t.Parallel()
	start := time.Unix(1_700_000_000, 0)
	clock := start
	f := NewFSMForTest(func() time.Time { return clock })

	grace := 5000
	pm := &config.ProtectiveMode{WSDownGraceMs: &grace}

	f.NotifyWS(false)
	clock = start.Add(4 * time.Second)
	f.ObserveConnectivity(pm, clock)
	if f.State() != StateRunning {
		t.Fatalf("expected still running, got %s", f.State())
	}
	clock = start.Add(5 * time.Second)
	f.ObserveConnectivity(pm, clock)
	if f.State() != StateProtectiveFlatten {
		t.Fatalf("state=%q want protective_flatten", f.State())
	}
	if f.LastReason() != reasonWSDown {
		t.Fatalf("reason=%q", f.LastReason())
	}
}

func TestFSM_AllowSubmitProtectiveFlatten(t *testing.T) {
	t.Parallel()
	f := NewFSM()
	f.ObserveMarket(market.Gap, nil)

	err := f.AllowSubmit(false)
	if err == nil || !errors.Is(err, ErrSubmitBlocked) {
		t.Fatalf("non-reduce-only: err=%v", err)
	}
	if err := f.AllowSubmit(true); err != nil {
		t.Fatalf("reduce-only: %v", err)
	}
}

func TestFSM_FlattenDeadlineFrozen(t *testing.T) {
	t.Parallel()
	start := time.Unix(1_800_000_000, 0)
	clock := start
	f := NewFSMForTest(func() time.Time { return clock })

	wait := 1000
	pm := &config.ProtectiveMode{MaxFlattenWaitMs: &wait}

	f.ObserveMarket(market.WideSpread, nil)
	clock = start.Add(1500 * time.Millisecond)
	f.ObserveFlattenDeadline(pm, clock)
	if f.State() != StateFrozen {
		t.Fatalf("state=%q want frozen", f.State())
	}
	if f.LastReason() != reasonFlattenWait {
		t.Fatalf("reason=%q", f.LastReason())
	}
}
