package audit

import (
	"strings"
	"testing"
)

func sc005RowComplete(r DecisionRecord) bool {
	if strings.TrimSpace(r.RegimeLabel) == "" || strings.TrimSpace(r.CostModelVersion) == "" {
		return false
	}
	if r.GateResults == nil || len(r.GateResults) == 0 {
		return false
	}
	return true
}

func sc005FractionComplete(rows []DecisionRecord) (pct float64) {
	if len(rows) == 0 {
		return 0
	}
	var n int
	for _, r := range rows {
		if sc005RowComplete(r) {
			n++
		}
	}
	return float64(n) / float64(len(rows)) * 100
}

// SC-005 (spec): acceptance sampling — at least 90% of a representative corpus carries
// regime label, cost model version, and risk gate map.
func TestSC005_SamplingMeetsNinetyPercent(t *testing.T) {
	t.Parallel()
	corpus := make([]DecisionRecord, 10)
	for i := range 9 {
		corpus[i] = DecisionRecord{
			RegimeLabel:      "normal",
			CostModelVersion: "cost_rules_v1",
			GateResults:      map[string]any{"daily_loss": true, "limit_delta": true},
		}
	}
	corpus[9] = DecisionRecord{
		RegimeLabel:      "",
		CostModelVersion: "cost_rules_v1",
		GateResults:      map[string]any{"daily_loss": true},
	}
	if p := sc005FractionComplete(corpus); p < 90.0 {
		t.Fatalf("got %.1f%% complete, want >= 90%%", p)
	}
}

// Breached / warning-style vetoes must be fully attributed (100% on injected set).
func TestSC005_WarningBreachRowsFullyAttributed(t *testing.T) {
	t.Parallel()
	breach := []DecisionRecord{
		{
			RegimeLabel:      "high",
			CostModelVersion: "cost_rules_v1",
			GateResults:      map[string]any{"limit_delta": false},
			Reason:           ReasonRiskVeto,
		},
		{
			RegimeLabel:      "low",
			CostModelVersion: "cm-v2",
			GateResults:      map[string]any{"limit_vega": false},
			Reason:           ReasonRiskVeto,
		},
	}
	for i, r := range breach {
		if !sc005RowComplete(r) {
			t.Fatalf("row %d incomplete: %+v", i, r)
		}
	}
}
