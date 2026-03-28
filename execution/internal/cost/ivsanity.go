package cost

import (
	"fmt"
	"math"
	"time"

	"github.com/dfr/optitrade/execution/internal/market"
)

// IVSanityOptions configures IV vs book cross-checks (WP08 T039, research.md IV sanity hook).
// When UseIVQuotes is false, checks are skipped (pure book strategies).
type IVSanityOptions struct {
	UseIVQuotes bool

	// MaxVolBookLag is the max allowed |book_ts - vol_ts| before veto iv_stale.
	MaxVolBookLag time.Duration

	// PrevUnderlyingMid optional prior mid (same instrument). When set with UseIVQuotes,
	// a rough consistency band compares relative mid jump to σ·√Δt using vol index as annualized DVOL-style σ.
	PrevUnderlyingMid *float64
	// JumpSigmaMultiplier widens the band (default 3).
	JumpSigmaMultiplier float64
}

// DefaultIVSanityOptions returns conservative defaults when o is zero-valued.
func DefaultIVSanityOptions(o IVSanityOptions) IVSanityOptions {
	if o.MaxVolBookLag == 0 {
		o.MaxVolBookLag = 2 * time.Minute
	}
	if o.JumpSigmaMultiplier == 0 {
		o.JumpSigmaMultiplier = 3
	}
	return o
}

// ivBookConflict implements T039: if IV quotes are in use, require timestamps aligned with the book
// and (when PrevUnderlyingMid is provided) reject explosive mid moves inconsistent with σ√Δt.
//
// σ is treated as a DVOL-style annualized index (0–5 typical). Δt is inferred from elapsed time
// between bookLocalTS and snapshot timestamp (falls back to 60s).
func ivBookConflict(snap market.MarketSnapshot, touchMid float64, opts IVSanityOptions) (conflict bool, detail string) {
	opts = DefaultIVSanityOptions(opts)
	if !opts.UseIVQuotes || snap.VolIndex <= 0 || snap.VolIndexTS <= 0 {
		return false, ""
	}
	if touchMid <= 0 || math.IsNaN(touchMid) {
		return true, "book_mid_invalid"
	}

	bookMs := snap.BookLocalTS.UnixMilli()
	if bookMs == 0 {
		bookMs = snap.Timestamp.UnixMilli()
	}
	lag := time.Duration(abs64(bookMs-snap.VolIndexTS)) * time.Millisecond
	if lag > opts.MaxVolBookLag {
		return true, fmt.Sprintf("iv_book_ts_lag_%s", lag.String())
	}

	if opts.PrevUnderlyingMid != nil && *opts.PrevUnderlyingMid > 0 {
		prev := *opts.PrevUnderlyingMid
		dtSec := snap.Timestamp.Sub(snap.BookLocalTS).Seconds()
		if dtSec <= 0 || math.IsNaN(dtSec) {
			dtSec = 60
		}
		// Δt in years for rough variance-time scaling (calendar minute heuristic).
		dtYear := dtSec / (365.0 * 24 * 3600)
		if dtYear <= 0 {
			dtYear = 60 / (365.0 * 24 * 3600)
		}
		sigmaMove := snap.VolIndex * math.Sqrt(dtYear)
		jump := math.Abs(touchMid-prev) / prev
		if jump > opts.JumpSigmaMultiplier*sigmaMove {
			return true, fmt.Sprintf("iv_mid_jump_jump=%.6f_band=%.6f", jump, opts.JumpSigmaMultiplier*sigmaMove)
		}
	}
	return false, ""
}

func abs64(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
