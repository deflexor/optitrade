package risk

import (
	"math/big"
	"time"
)

// DailyLossTracker resets at each policy session boundary in UTC, optionally
// shifted by Offset (e.g. Offset=2h for UTC+2 midnight as boundary anchor).
type DailyLossTracker struct {
	Offset time.Duration

	sessionKey string
	startPnL   *big.Rat
}

func (t *DailyLossTracker) sessionKeyFor(now time.Time) string {
	boundary := now.UTC().Add(-t.Offset)
	return boundary.Format("2006-01-02")
}

func (t *DailyLossTracker) ensureSession(now time.Time, cumPnL *big.Rat) {
	if cumPnL == nil {
		return
	}
	key := t.sessionKeyFor(now)
	if key != t.sessionKey {
		t.sessionKey = key
		t.startPnL = new(big.Rat).Set(cumPnL)
	}
	if t.startPnL == nil {
		t.startPnL = new(big.Rat).Set(cumPnL)
	}
}

// SessionLoss returns non-negative loss since the session start for signed
// cumulative PnL (higher is better). If cumPnL is nil, returns nil.
func (t *DailyLossTracker) SessionLoss(now time.Time, cumPnL *big.Rat) *big.Rat {
	if cumPnL == nil {
		return nil
	}
	t.ensureSession(now, cumPnL)
	diff := new(big.Rat).Sub(new(big.Rat).Set(t.startPnL), cumPnL)
	if diff.Sign() <= 0 {
		return new(big.Rat)
	}
	return diff
}
