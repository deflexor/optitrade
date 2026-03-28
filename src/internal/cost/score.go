package cost

import (
	"math"
	"strconv"
	"strings"

	"github.com/dfr/optitrade/src/internal/config"
	"github.com/dfr/optitrade/src/internal/deribit"
	"github.com/dfr/optitrade/src/internal/market"
	"github.com/dfr/optitrade/src/internal/regime"
	"github.com/dfr/optitrade/src/internal/strategy"
)

// CandidateInput is the minimum trade candidate metadata needed for scoring (expected edge + books).
type CandidateInput struct {
	ExpectedEdge    string // decimal string, BTC/ETH quote currency per instrument (FR-006)
	QuoteCurrency   string // optional hint, e.g. "BTC"; used for breakdown only
	PrimaryMid      float64
	PrimaryMidValid bool
}

// ScoreCandidate estimates costs from policy + touch books and compares to expected edge (WP08 T038).
// vetoReason is stable machine text: iv_stale, non_positive_edge, invalid_expected_edge, nil_policy, no_book.
//
// All haircuts are applied as: haircut = edge * (totalBps / 10_000), i.e. bps of the premium leg.
// This intentionally conflates underlying touch bps with premium bps as a conservative linear proxy when
// quote_currency is the underlying coin (Deribit linear options); see WP08 risk note.
func ScoreCandidate(
	policy *config.Policy,
	label regime.Label,
	cand CandidateInput,
	books []deribit.OrderBook,
	snap *market.MarketSnapshot,
	iv IVSanityOptions,
) (ok bool, vetoReason string, breakdown CostBreakdown) {
	breakdown.CostModelVersion = CostModelVersion
	breakdown.QuoteCurrency = strings.TrimSpace(cand.QuoteCurrency)
	breakdown.ExpectedEdgeRaw = cand.ExpectedEdge

	if policy == nil {
		return false, "nil_policy", breakdown
	}
	if len(books) == 0 {
		return false, "no_book", breakdown
	}

	touch, maxHalf, okBook := maxHalfSpreadBps(books)
	if !okBook {
		return false, "no_book", breakdown
	}

	edge, err := strconv.ParseFloat(strings.TrimSpace(cand.ExpectedEdge), 64)
	if err != nil || math.IsNaN(edge) || math.IsInf(edge, 0) {
		return false, "invalid_expected_edge", breakdown
	}
	breakdown.ExpectedEdge = edge

	midForIV := touch.Mid
	if cand.PrimaryMidValid && cand.PrimaryMid > 0 {
		midForIV = cand.PrimaryMid
	}

	if snap != nil {
		if conflict, _ := ivBookConflict(*snap, midForIV, iv); conflict {
			breakdown = buildBreakdown(policy, label, cand, edge, maxHalf, touch)
			return false, "iv_stale", breakdown
		}
	}

	breakdown = buildBreakdown(policy, label, cand, edge, maxHalf, touch)
	if breakdown.EdgeAfterCosts <= 0 {
		return false, "non_positive_edge", breakdown
	}
	return true, "", breakdown
}

func buildBreakdown(
	policy *config.Policy,
	label regime.Label,
	cand CandidateInput,
	edge float64,
	maxHalfSpread float64,
	touch strategy.TouchLiquidity,
) CostBreakdown {
	fee := feeBpsFromPolicy(policy)
	slip := slippageBpsFromPolicy(policy)
	adv := adverseBpsForRegime(policy, label)
	totalBps := float64(fee+slip+adv) + maxHalfSpread

	haircut := edge * (totalBps / 10_000)
	after := edge - haircut

	return CostBreakdown{
		CostModelVersion: CostModelVersion,
		QuoteCurrency:    strings.TrimSpace(cand.QuoteCurrency),
		ExpectedEdgeRaw:  cand.ExpectedEdge,
		ExpectedEdge:     edge,
		FeeBps:           fee,
		HalfSpreadBps:    maxHalfSpread,
		SlippageBps:      slip,
		AdverseBps:       adv,
		TotalHaircutBps:  totalBps,
		HaircutAmount:    haircut,
		EdgeAfterCosts:   after,
	}
}

func maxHalfSpreadBps(books []deribit.OrderBook) (touch strategy.TouchLiquidity, maxHalf float64, ok bool) {
	var first strategy.TouchLiquidity
	var seen bool
	for _, b := range books {
		t, okBook := strategy.TouchFromBook(b)
		if !okBook {
			continue
		}
		if !seen {
			first = t
			seen = true
		}
		h := t.SpreadBps / 2
		if h > maxHalf {
			maxHalf = h
		}
	}
	return first, maxHalf, seen
}
