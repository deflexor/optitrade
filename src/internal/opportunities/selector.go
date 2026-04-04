package opportunities

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dfr/optitrade/src/internal/audit"
	"github.com/dfr/optitrade/src/internal/config"
	"github.com/dfr/optitrade/src/internal/cost"
	"github.com/dfr/optitrade/src/internal/deribit"
	"github.com/dfr/optitrade/src/internal/regime"
	"github.com/dfr/optitrade/src/internal/risk"
	"github.com/dfr/optitrade/src/internal/strategy"
)

// BookFetcher returns top-of-book for an instrument id (OKX instId or Deribit name).
type BookFetcher interface {
	FetchBook(ctx context.Context, inst string) (deribit.OrderBook, error)
}

// CandidateSpec is one put-credit spread location in the grid (M1).
type CandidateSpec struct {
	Base        string
	Expiry      string // YYMMDD (8 digits from OKX chain)
	ShortStrike int64
	Width       int
}

// Selector evaluates spread candidates with cost and risk gates.
type Selector struct {
	Policy *config.Policy
	Label  regime.Label
	Books  BookFetcher
}

// Evaluate ranks specs and returns rows sorted by edge after costs (desc).
func (s *Selector) Evaluate(ctx context.Context, specs []CandidateSpec) ([]Row, error) {
	if s == nil || s.Books == nil {
		return nil, fmt.Errorf("opportunities: nil selector or book fetcher")
	}
	if s.Policy == nil {
		return nil, fmt.Errorf("opportunities: nil policy")
	}
	eng := risk.NewEngine(s.Policy, audit.NopDecisionLogger{})
	now := time.Now().UTC()

	var out []Row
	for i, sp := range specs {
		row, err := s.evaluateOne(ctx, eng, now, i, sp)
		if err != nil {
			return nil, err
		}
		out = append(out, row)
	}

	sort.SliceStable(out, func(i, j int) bool { return out[i].EdgeAfter > out[j].EdgeAfter })
	return out, nil
}

func (s *Selector) evaluateOne(ctx context.Context, eng *risk.Engine, now time.Time, seq int, sp CandidateSpec) (Row, error) {
	legs, err := strategy.VerticalPutCreditOKX(sp.Base, sp.Expiry, sp.ShortStrike, sp.Width)
	if err != nil {
		return Row{}, err
	}
	width := float64(sp.Width)
	if sp.Width <= 0 {
		width = float64(strategy.DefaultStrikeWidth)
	}

	books := make([]deribit.OrderBook, 0, len(legs))
	quotes := make([]LegQuote, 0, len(legs))
	for _, lg := range legs {
		b, err := s.Books.FetchBook(ctx, lg.Instrument)
		if err != nil {
			return Row{}, fmt.Errorf("book %q: %w", lg.Instrument, err)
		}
		b.InstrumentName = lg.Instrument
		books = append(books, b)
		touch, ok := strategy.TouchFromBook(b)
		if !ok {
			return Row{
				ID:           fmt.Sprintf("cand-%s-%s-%d", sp.Expiry, sp.Base, seq),
				StrategyName: "credit_spread",
				Status:       StatusCandidate,
				Legs:         quotes,
				MaxProfit:    "0",
				MaxLoss:      strconv.FormatFloat(width, 'f', -1, 64),
				Recommend:    "pass",
				Rationale:    "missing bid or ask on leg",
				ExpectedEdge: "0",
				EdgeAfter:    math.Inf(-1),
			}, nil
		}
		quotes = append(quotes, LegQuote{
			Instrument: lg.Instrument,
			Bid:        touch.BidPrice,
			Ask:        touch.AskPrice,
		})
	}

	// legs[0] = short (sell), legs[1] = long (buy). Conservative credit: bid short - ask long.
	credit := quotes[0].Bid - quotes[1].Ask
	if credit <= 0 || math.IsNaN(credit) || math.IsInf(credit, 0) {
		return Row{
			ID:           fmt.Sprintf("cand-%s-%s-%d-%d", sp.Base, sp.Expiry, sp.ShortStrike, seq),
			StrategyName: "credit_spread",
			Status:       StatusCandidate,
			Legs:         quotes,
			MaxProfit:    "0",
			MaxLoss:      strconv.FormatFloat(width, 'f', -1, 64),
			Recommend:    "pass",
			Rationale:    "non_positive_credit_at_conservative_quotes",
			ExpectedEdge: strconv.FormatFloat(credit, 'f', -1, 64),
			EdgeAfter:    math.Inf(-1),
		}, nil
	}
	maxLoss := width - credit
	if maxLoss < 0 {
		maxLoss = width
	}
	maxProfit := credit

	edgeStr := strconv.FormatFloat(credit, 'f', -1, 64)
	candIn := cost.CandidateInput{
		ExpectedEdge:    edgeStr,
		QuoteCurrency:   strings.ToUpper(strings.TrimSpace(sp.Base)),
		PrimaryMidValid: len(quotes) > 0,
	}
	if len(quotes) > 0 {
		candIn.PrimaryMid = (quotes[0].Bid + quotes[0].Ask) / 2
	}

	okCost, veto, bd := cost.ScoreCandidate(s.Policy, s.Label, candIn, books, nil, cost.IVSanityOptions{})

	row := Row{
		ID:           fmt.Sprintf("cand-%s-%s-%d-%d", sp.Base, sp.Expiry, sp.ShortStrike, seq),
		StrategyName: "credit_spread",
		Status:       StatusCandidate,
		Legs:         quotes,
		GreeksNote:   "",
		MaxProfit:    strconv.FormatFloat(maxProfit, 'f', -1, 64),
		MaxLoss:      strconv.FormatFloat(maxLoss, 'f', -1, 64),
		ExpectedEdge: edgeStr,
		EdgeAfter:    bd.EdgeAfterCosts,
	}

	if !okCost {
		row.Recommend = "pass"
		row.Rationale = "cost veto: " + veto
		row.EdgeAfter = bd.EdgeAfterCosts
		return row, nil
	}

	corrID := fmt.Sprintf("sel-%s-%d", sp.Expiry, seq)
	maxLossStr := strconv.FormatFloat(maxLoss, 'f', -1, 64)
	insts := make([]string, len(legs))
	for i := range legs {
		insts[i] = legs[i].Instrument
	}
	riskIn := risk.PreTradeInput{
		CorrelationID:    corrID,
		RegimeLabel:      string(s.Label),
		CostModelVersion: cost.CostModelVersion,
		Positions:        nil,
		OpenOrders:       nil,
		Candidate: risk.CandidateRisk{
			ID:           row.ID,
			StrategyID:   "opportunities-m1",
			MaxLossQuote: maxLossStr,
			FeesQuote:    "0",
			Instruments:  insts,
		},
		Now:              now,
		CumulativePnL:    big.NewRat(0, 1),
		DailyTracker:     &risk.DailyLossTracker{},
		StrategyOpenedAt: map[string]time.Time{},
	}

	allowed, err := eng.Check(ctx, riskIn)
	if err != nil {
		return Row{}, fmt.Errorf("risk check: %w", err)
	}
	if !allowed {
		row.Recommend = "pass"
		row.Rationale = "risk veto (see audit)"
		row.EdgeAfter = bd.EdgeAfterCosts
		return row, nil
	}

	row.Recommend = "open"
	row.Rationale = fmt.Sprintf("edge_after_costs=%.6g max_loss=%s", bd.EdgeAfterCosts, maxLossStr)
	return row, nil
}
