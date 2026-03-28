package cost

import (
	"github.com/dfr/optitrade/src/internal/config"
	"github.com/dfr/optitrade/src/internal/regime"
)

// CostModelVersion labels persisted audit rows for this scorer (T036–T037).
const CostModelVersion = "cost_rules_v1"

// default cost knobs when policy omits cost_model (conservative).
const (
	defaultTakerFeeBps   = 5
	defaultMakerFeeBps   = 0
	defaultSlippageBps   = 3
	defaultAdverseLow    = 2
	defaultAdverseHigh   = 6
	defaultAdverseNormal = 3
)

func feeBpsFromPolicy(p *config.Policy) int {
	if p == nil || p.CostModel == nil || p.CostModel.TakerFeeBps == nil {
		return defaultTakerFeeBps
	}
	return *p.CostModel.TakerFeeBps
}

func slippageBpsFromPolicy(p *config.Policy) int {
	if p == nil || p.CostModel == nil || p.CostModel.SlippageBps == nil {
		return defaultSlippageBps
	}
	return *p.CostModel.SlippageBps
}

// adverseBpsForRegime returns adverse-selection haircut in bps (T037).
//
// Formula (piecewise): adverse_bps = policy.adverse_selection_bps_low in low vol,
// _normal uses the average of low/high when only endpoints are set, else midpoint of configured low/high;
// _high uses adverse_selection_bps_high. This is a linear stress overlay on the bid–ask microstructure,
// not an options model.
func adverseBpsForRegime(p *config.Policy, label regime.Label) int {
	low, high := defaultAdverseLow, defaultAdverseHigh
	if p != nil && p.CostModel != nil {
		if p.CostModel.AdverseSelectionBpsLow != nil {
			low = *p.CostModel.AdverseSelectionBpsLow
		}
		if p.CostModel.AdverseSelectionBpsHigh != nil {
			high = *p.CostModel.AdverseSelectionBpsHigh
		}
	}
	switch label {
	case regime.LabelLow:
		return low
	case regime.LabelHigh:
		return high
	default:
		mid := (low + high) / 2
		if mid < 1 {
			mid = defaultAdverseNormal
		}
		return mid
	}
}
