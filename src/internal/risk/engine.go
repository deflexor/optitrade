package risk

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/dfr/optitrade/src/internal/audit"
	"github.com/dfr/optitrade/src/internal/config"
)

// RiskModelVersion labels audit rows for this risk engine (WP09).
const RiskModelVersion = "risk_rules_v1"

// Engine runs portfolio limit, daily loss, per-trade, and time-in-trade gates.
type Engine struct {
	policy *config.Policy
	log    audit.DecisionLogger
}

// NewEngine returns a risk engine. policy must be non-nil for checks to succeed.
func NewEngine(policy *config.Policy, log audit.DecisionLogger) *Engine {
	return &Engine{policy: policy, log: log}
}

// Check runs all gates. On veto it persists/logs an audit_decision (FR-010) and returns allowed=false.
func (e *Engine) Check(ctx context.Context, in PreTradeInput) (allowed bool, err error) {
	if e == nil {
		return false, fmt.Errorf("risk: nil engine")
	}
	if e.policy == nil {
		return false, fmt.Errorf("risk: nil policy")
	}
	if strings.TrimSpace(in.CorrelationID) == "" {
		return false, fmt.Errorf("risk: correlation_id required")
	}

	snap := BuildPortfolioSnapshot(in.Positions, in.OpenOrders, in.DeltaAdjustment, in.VegaAdjustment)
	gates := e.runGates(in, snap)
	ok := true
	for _, k := range gateOrder {
		if v, isBool := gates[k].(bool); !isBool || !v {
			ok = false
			break
		}
	}
	if ok {
		return true, nil
	}
	if e.log == nil {
		return false, fmt.Errorf("risk: nil decision logger")
	}
	cmv := strings.TrimSpace(in.CostModelVersion)
	if cmv == "" {
		cmv = RiskModelVersion
	}
	rec := audit.DecisionRecord{
		CorrelationID:    in.CorrelationID,
		DecisionType:     "risk_veto",
		CandidateID:      in.CandidateID,
		RegimeLabel:      strings.TrimSpace(in.RegimeLabel),
		CostModelVersion: cmv,
		GateResults:      gates,
		Reason:           audit.ReasonRiskVeto,
		TsMs:             in.Now.UnixMilli(),
	}
	if err := e.log.LogDecision(ctx, rec); err != nil {
		return false, err
	}
	return false, nil
}

func (e *Engine) runGates(in PreTradeInput, snap PortfolioSnapshot) map[string]any {
	p := e.policy
	out := map[string]any{}

	maxDelta, err := ParseDecimalRat(p.Limits.MaxPortfolioDelta)
	out["limit_delta"] = err == nil && ratLessOrEqual(ratAbsCopy(snap.DeltaTotal), maxDelta)
	if err != nil {
		out["limit_delta"] = false
	}

	maxVega, err := ParseDecimalRat(p.Limits.MaxPortfolioVega)
	out["limit_vega"] = err == nil && ratLessOrEqual(ratAbsCopy(snap.VegaTotal), maxVega)
	if err != nil {
		out["limit_vega"] = false
	}

	maxPAR, err := ParseDecimalRat(p.Limits.MaxOpenPremiumAtRisk)
	out["limit_premium_at_risk"] = err == nil && ratLessOrEqual(snap.PremiumAtRisk, maxPAR)
	if err != nil {
		out["limit_premium_at_risk"] = false
	}

	limit := p.Limits.MaxOpenOrdersPerInstrument
	okOrders := true
	for _, inst := range in.Candidate.Instruments {
		inst = strings.TrimSpace(inst)
		if inst == "" {
			continue
		}
		have := snap.OpenOrdersByInstrument[inst]
		if have+1 > limit {
			okOrders = false
			break
		}
	}
	out["limit_open_orders"] = okOrders

	out["daily_loss"] = e.gateDailyLoss(in)
	out["per_trade_max_loss"] = e.gatePerTradeMaxLoss(in)
	out["time_in_trade"] = e.gateTimeInTrade(in)

	return out
}

var gateOrder = []string{
	"limit_delta",
	"limit_vega",
	"limit_premium_at_risk",
	"limit_open_orders",
	"daily_loss",
	"per_trade_max_loss",
	"time_in_trade",
}

func (e *Engine) gateDailyLoss(in PreTradeInput) bool {
	if in.CumulativePnL == nil {
		return true
	}
	maxLoss, err := ParseDecimalRat(e.policy.Limits.MaxDailyLoss)
	if err != nil {
		return false
	}
	tr := in.DailyTracker
	if tr == nil {
		tr = &DailyLossTracker{}
	}
	loss := tr.SessionLoss(in.Now, in.CumulativePnL)
	if loss == nil {
		return true
	}
	return ratLessOrEqual(loss, maxLoss)
}

func (e *Engine) gatePerTradeMaxLoss(in PreTradeInput) bool {
	max, err := ParseDecimalRat(e.policy.Limits.MaxLossPerTrade)
	if err != nil {
		return false
	}
	tradeMax, err := ParseDecimalRat(strings.TrimSpace(in.Candidate.MaxLossQuote))
	if err != nil {
		return false
	}
	fees := new(big.Rat)
	if strings.TrimSpace(in.Candidate.FeesQuote) != "" {
		f, err := ParseDecimalRat(in.Candidate.FeesQuote)
		if err != nil {
			return false
		}
		fees = f
	}
	total := new(big.Rat).Add(tradeMax, fees)
	return ratLessOrEqual(total, max)
}

func (e *Engine) gateTimeInTrade(in PreTradeInput) bool {
	sid := strings.TrimSpace(in.Candidate.StrategyID)
	if sid == "" {
		return true
	}
	opened, ok := in.StrategyOpenedAt[sid]
	if !ok {
		return true
	}
	maxDur := time.Duration(e.policy.Limits.MaxTimeInTradeSeconds) * time.Second
	return in.Now.Sub(opened) < maxDur
}

// EvaluateDryRun runs gates without auditing (tests and diagnostics).
func (e *Engine) EvaluateDryRun(in PreTradeInput) (ok bool, gates map[string]any) {
	if e == nil || e.policy == nil {
		return false, nil
	}
	snap := BuildPortfolioSnapshot(in.Positions, in.OpenOrders, in.DeltaAdjustment, in.VegaAdjustment)
	gates = e.runGates(in, snap)
	for _, k := range gateOrder {
		if v, isBool := gates[k].(bool); !isBool || !v {
			return false, gates
		}
	}
	return true, gates
}
