package dashboard

import (
	"context"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dfr/optitrade/src/internal/config"
	"github.com/dfr/optitrade/src/internal/deribit"
	"github.com/dfr/optitrade/src/internal/opportunities"
	"github.com/dfr/optitrade/src/internal/okx"
	"github.com/dfr/optitrade/src/internal/regime"
	"github.com/dfr/optitrade/src/internal/strategy"
)

const (
	runnerTickTimeout   = 25 * time.Second
	runnerMaxSpecs      = 40
	runnerTopN          = 20
	runnerMaxExpiries   = 3
	runnerExpiryHorizon = 30 * 24 * time.Hour
)

type okxPublicBookFetcher struct {
	pub *okx.PublicClient
}

func (f *okxPublicBookFetcher) FetchBook(ctx context.Context, inst string) (deribit.OrderBook, error) {
	return f.pub.GetOrderBookDeribit(ctx, inst, 5)
}

func runnerRegimeLabel() regime.Label {
	s := strings.ToLower(strings.TrimSpace(os.Getenv("OPTITRADE_DASHBOARD_REGIME_LABEL")))
	switch s {
	case "low":
		return regime.LabelLow
	case "high":
		return regime.LabelHigh
	default:
		return regime.LabelNormal
	}
}

func parseOKXPutOptionInst(instId string) (base string, yyyymmdd string, strike int64, ok bool) {
	p := strings.Split(instId, "-")
	if len(p) != 5 || !strings.EqualFold(p[1], "USD") {
		return "", "", 0, false
	}
	if strings.ToUpper(strings.TrimSpace(p[4])) != "P" {
		return "", "", 0, false
	}
	base = strings.ToUpper(strings.TrimSpace(p[0]))
	yyMMdd := p[2]
	if len(yyMMdd) != 6 {
		return "", "", 0, false
	}
	yyyymmdd = "20" + yyMMdd
	st, err := strconv.ParseInt(p[3], 10, 64)
	if err != nil || st <= 0 {
		return "", "", 0, false
	}
	return base, yyyymmdd, st, true
}

func expiryDateUTC(yyyymmdd string) (time.Time, error) {
	return time.ParseInLocation("20060102", yyyymmdd, time.UTC)
}

func sumEquityUSD(sums []deribit.AccountSummary) float64 {
	var t float64
	for _, s := range sums {
		if s.Equity != nil && *s.Equity > 0 {
			t += *s.Equity
		}
	}
	return t
}

func policyAllowsCreditSpread(policy *config.Policy, label regime.Label) bool {
	if policy == nil {
		return false
	}
	allowed, err := strategy.AllowedStructures(policy, label)
	if err != nil {
		return false
	}
	for _, n := range allowed {
		if strings.EqualFold(strings.TrimSpace(n), "credit_spread") {
			return true
		}
	}
	return false
}

// buildPutCreditSpecs builds up to maxSpecs candidate put-credit locations from listed OKX puts.
func buildPutCreditSpecs(
	instruments []okx.InstrumentSummary,
	spot float64,
	width int,
	maxSpecs int,
	wantBase string,
) []opportunities.CandidateSpec {
	if spot <= 0 || math.IsNaN(spot) || math.IsInf(spot, 0) {
		return nil
	}
	wantBase = strings.ToUpper(strings.TrimSpace(wantBase))
	if wantBase == "" {
		wantBase = "BTC"
	}
	if width <= 0 {
		width = strategy.DefaultStrikeWidth
	}
	instSet := make(map[string]struct{}, len(instruments))
	for _, ins := range instruments {
		instSet[ins.InstId] = struct{}{}
	}
	byExpiry := make(map[string][]int64)
	for _, ins := range instruments {
		base, yyyymmdd, strike, ok := parseOKXPutOptionInst(ins.InstId)
		if !ok || base != wantBase {
			continue
		}
		exp, err := expiryDateUTC(yyyymmdd)
		if err != nil {
			continue
		}
		now := time.Now().UTC()
		startDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		if exp.Before(startDay) {
			continue
		}
		if exp.After(now.Add(runnerExpiryHorizon)) {
			continue
		}
		byExpiry[yyyymmdd] = append(byExpiry[yyyymmdd], strike)
	}
	if len(byExpiry) == 0 {
		return nil
	}
	type expKey struct {
		yyyymmdd string
		t        time.Time
	}
	var keys []expKey
	for y := range byExpiry {
		t, err := expiryDateUTC(y)
		if err != nil {
			continue
		}
		keys = append(keys, expKey{yyyymmdd: y, t: t})
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i].t.Before(keys[j].t) })
	if len(keys) > runnerMaxExpiries {
		keys = keys[:runnerMaxExpiries]
	}

	base := wantBase

	lowS := spot * 0.88
	highS := spot * 0.995
	var out []opportunities.CandidateSpec

outer:
	for _, ek := range keys {
		strikes := byExpiry[ek.yyyymmdd]
		sort.Slice(strikes, func(i, j int) bool { return strikes[i] > strikes[j] })
		for _, shortStrike := range strikes {
			sf := float64(shortStrike)
			if sf < lowS || sf > highS {
				continue
			}
			longStrike := shortStrike - int64(width)
			if longStrike <= 0 {
				continue
			}
			sid := strategy.OKXOptionInstID(base, ek.yyyymmdd, shortStrike, false)
			lid := strategy.OKXOptionInstID(base, ek.yyyymmdd, longStrike, false)
			if _, ok := instSet[sid]; !ok {
				continue
			}
			if _, ok := instSet[lid]; !ok {
				continue
			}
			out = append(out, opportunities.CandidateSpec{
				Base:        base,
				Expiry:      ek.yyyymmdd,
				ShortStrike: shortStrike,
				Width:       width,
			})
			if len(out) >= maxSpecs {
				break outer
			}
		}
	}
	return out
}

func applyMaxLossEquityGate(rows []opportunities.Row, equityUSD float64, maxLossPct int) {
	if equityUSD <= 0 || maxLossPct < 1 {
		return
	}
	limit := equityUSD * float64(maxLossPct) / 100.0
	for i := range rows {
		if !strings.EqualFold(rows[i].Recommend, "open") {
			continue
		}
		ml, err := strconv.ParseFloat(strings.TrimSpace(rows[i].MaxLoss), 64)
		if err != nil || math.IsNaN(ml) || math.IsInf(ml, 0) {
			continue
		}
		if ml > limit {
			rows[i].Recommend = "pass"
			rows[i].Rationale = fmt.Sprintf("max_loss_exceeds_equity_pct (limit %.4g USD on EqUsd rollup)", limit)
		}
	}
}

func (rm *RunnerManager) runTick(parent context.Context, user string) {
	if rm == nil {
		return
	}
	ctx, cancel := context.WithTimeout(parent, runnerTickTimeout)
	defer cancel()

	row, err := rm.settings.GetDecrypting(ctx, user, rm.crypto)
	if err != nil {
		rm.log.Warn("runner tick settings", "user", user, "err", err)
		return
	}
	if row == nil || !RunnerEligible(row) {
		rm.setSnapshot(user, opportunities.Snapshot{})
		return
	}
	if rm.policy == nil {
		rm.log.Debug("runner tick skip: no policy", "user", user)
		rm.setSnapshot(user, opportunities.Snapshot{})
		return
	}
	label := runnerRegimeLabel()
	if !policyAllowsCreditSpread(rm.policy, label) {
		rm.setSnapshot(user, opportunities.Snapshot{})
		return
	}

	xchg, err := newOKXExchange(row)
	if err != nil {
		rm.log.Warn("runner tick exchange", "user", user, "err", err)
		rm.setSnapshot(user, opportunities.Snapshot{})
		return
	}
	sums, err := xchg.GetAccountSummaries(ctx, nil)
	if err != nil {
		rm.log.Warn("runner tick balances", "user", user, "err", err)
	}
	equityUSD := sumEquityUSD(sums)

	pub := &okx.PublicClient{}
	idxInst := "BTC-USD"
	if row.Currencies != "" {
		parts := strings.Split(row.Currencies, ",")
		if len(parts) > 0 {
			ccy := strings.ToUpper(strings.TrimSpace(parts[0]))
			if ccy != "" && ccy != "BTC" {
				idxInst = ccy + "-USD"
			}
		}
	}
	spot, err := pub.GetIndexPrice(ctx, idxInst)
	if err != nil {
		rm.log.Warn("runner tick index", "user", user, "inst", idxInst, "err", err)
		rm.setSnapshot(user, opportunities.Snapshot{})
		return
	}
	wantBase := strings.ToUpper(strings.TrimSuffix(idxInst, "-USD"))
	instruments, err := pub.GetInstruments(ctx, "OPTION", idxInst)
	if err != nil {
		rm.log.Warn("runner tick instruments", "user", user, "uly", idxInst, "err", err)
		rm.setSnapshot(user, opportunities.Snapshot{})
		return
	}

	width := strategy.DefaultStrikeWidth
	specs := buildPutCreditSpecs(instruments, spot, width, runnerMaxSpecs, wantBase)
	if len(specs) == 0 {
		rm.setSnapshot(user, opportunities.Snapshot{UpdatedAtMs: time.Now().UnixMilli(), Rows: nil})
		return
	}

	sel := &opportunities.Selector{
		Policy: rm.policy,
		Label:  label,
		Books:  &okxPublicBookFetcher{pub: pub},
	}
	rows, err := sel.Evaluate(ctx, specs)
	if err != nil {
		rm.log.Warn("runner tick selector", "user", user, "err", err)
		rm.setSnapshot(user, opportunities.Snapshot{})
		return
	}
	applyMaxLossEquityGate(rows, equityUSD, row.MaxLossEquityPct)
	if len(rows) > runnerTopN {
		rows = rows[:runnerTopN]
	}
	rm.setSnapshot(user, opportunities.Snapshot{
		UpdatedAtMs: time.Now().UnixMilli(),
		Rows:        rows,
	})
}

func (rm *RunnerManager) setSnapshot(user string, snap opportunities.Snapshot) {
	if rm == nil {
		return
	}
	u := strings.TrimSpace(user)
	rm.snapMu.Lock()
	defer rm.snapMu.Unlock()
	if rm.snapshots == nil {
		rm.snapshots = map[string]opportunities.Snapshot{}
	}
	rm.snapshots[u] = snap
}

// Snapshot returns the last successful runner snapshot for the user (empty if none).
func (rm *RunnerManager) Snapshot(username string) opportunities.Snapshot {
	if rm == nil {
		return opportunities.Snapshot{}
	}
	u := strings.TrimSpace(username)
	rm.snapMu.RLock()
	defer rm.snapMu.RUnlock()
	if rm.snapshots == nil {
		return opportunities.Snapshot{}
	}
	s, ok := rm.snapshots[u]
	if !ok {
		return opportunities.Snapshot{}
	}
	return s
}

func (rm *RunnerManager) clearSnapshot(user string) {
	if rm == nil {
		return
	}
	u := strings.TrimSpace(user)
	rm.snapMu.Lock()
	defer rm.snapMu.Unlock()
	if rm.snapshots != nil {
		delete(rm.snapshots, u)
	}
}
