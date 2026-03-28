package cost

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dfr/optitrade/execution/internal/audit"
	"github.com/dfr/optitrade/execution/internal/regime"
)

// LogCostVeto persists a cost or IV veto via DecisionLogger (WP08 FR-010).
func LogCostVeto(
	ctx context.Context,
	dl audit.DecisionLogger,
	correlationID string,
	candidateID *string,
	regimeLabel regime.Label,
	vetoCode string,
	breakdown CostBreakdown,
) error {
	if dl == nil {
		return fmt.Errorf("cost: nil DecisionLogger")
	}
	var bdMap map[string]any
	if b, err := json.Marshal(breakdown); err == nil {
		_ = json.Unmarshal(b, &bdMap)
	}
	gates := map[string]any{
		"edge_after_costs_positive": false,
		"iv_book_consistent":        vetoCode != "iv_stale",
		"veto_code":                 vetoCode,
	}
	if bdMap != nil {
		gates["cost_breakdown"] = bdMap
	}
	rec := audit.DecisionRecord{
		CorrelationID:    correlationID,
		DecisionType:     "veto_cost",
		CandidateID:      candidateID,
		RegimeLabel:      string(regimeLabel),
		CostModelVersion: breakdown.CostModelVersion,
		GateResults:      gates,
		Reason:           auditReasonForCostVeto(vetoCode),
	}
	return dl.LogDecision(ctx, rec)
}

func auditReasonForCostVeto(code string) audit.DecisionReason {
	if code == "iv_stale" {
		return audit.ReasonIVStale
	}
	return audit.ReasonCostVeto
}
