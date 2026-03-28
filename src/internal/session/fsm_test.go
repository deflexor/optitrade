package session

import (
	"errors"
	"testing"
	"time"

	"github.com/dfr/optitrade/src/internal/config"
	"github.com/dfr/optitrade/src/internal/market"
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

// SC-003 (spec): auth failure moves to protective_flatten immediately (detection latency ~0 here);
// non-reduce-only submits are blocked afterward (FR-009).
func TestSC003_RPCAuthFailureImmediateProtective(t *testing.T) {
	t.Parallel()
	f := NewFSM()
	f.NotifyRPCAuthFailure()
	if f.State() != StateProtectiveFlatten {
		t.Fatalf("state=%s", f.State())
	}
	if f.LastReason() != reasonRPCAuth {
		t.Fatalf("reason=%s", f.LastReason())
	}
	if err := f.AllowSubmit(false); err == nil {
		t.Fatal("expected blocked non-reduce-only submit")
	}
}

// SC-003: feed-quality trip is immediate on ObserveMarket (stale book); see WP12 T058–T059 for WS grace timing.
func TestSC003_MarketStaleImmediateProtective(t *testing.T) {
	t.Parallel()
	f := NewFSM()
	f.ObserveMarket(market.StaleBook, nil)
	if f.State() != StateProtectiveFlatten {
		t.Fatalf("state=%s", f.State())
	}
	if err := f.AllowSubmit(false); err == nil {
		t.Fatal("expected block")
	}
}

// SC-003 budget: configured ws_down_grace_ms is under the 60s worst-case target for this build (policy default 10s; here 59s).
func TestSC003_WSDownGraceTransitionsBefore60s(t *testing.T) {
	t.Parallel()
	start := time.Unix(1_710_000_000, 0)
	clock := start
	f := NewFSMForTest(func() time.Time { return clock })
	ms := 59_000
	pm := &config.ProtectiveMode{WSDownGraceMs: &ms}
	f.NotifyWS(false)
	clock = start.Add(58 * time.Second)
	f.ObserveConnectivity(pm, clock)
	if f.State() != StateRunning {
		t.Fatalf("pre-grace: %s", f.State())
	}
	clock = start.Add(59*time.Second + time.Millisecond)
	f.ObserveConnectivity(pm, clock)
	if f.State() != StateProtectiveFlatten {
		t.Fatalf("want protective after 59s grace, got %s", f.State())
	}
	if f.LastReason() != reasonWSDown {
		t.Fatalf("reason=%s", f.LastReason())
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
