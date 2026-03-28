package regime

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/dfr/optitrade/src/internal/config"
)

// ClassifierRulesV1 is embedded in persisted regime_state for audit (research.md, classifier MVP).
const ClassifierRulesV1 = "rules_v1/1"

// Label is a playbook regime bucket (policy playbooks keys).
type Label string

const (
	LabelLow    Label = "low"
	LabelNormal Label = "normal"
	LabelHigh   Label = "high"
)

// Outcome is the effective regime after rules and hysteresis.
type Outcome struct {
	Label             Label
	ClassifierVersion string
}

type thresholdKind int

const (
	thrLow thresholdKind = iota
	thrHigh
)

// classifyRulesV1 maps vol index to a raw label before hysteresis.
func classifyRulesV1(r *config.Regime, vol float64, volOK bool) (Label, error) {
	if r == nil || strings.TrimSpace(r.Classifier) == "" || r.Classifier != "rules_v1" {
		return LabelNormal, nil
	}
	if !volOK {
		if strings.TrimSpace(r.OnMissingVol) == "hold_last" {
			return "", fmt.Errorf("vol missing: hold_last requires hysteresis state (use EvaluateSnapshot)")
		}
		return LabelNormal, nil
	}
	low, err := parseThreshold(r.LowVolThresholdIndex, thrLow)
	if err != nil {
		return LabelNormal, nil
	}
	high, err := parseThreshold(r.HighVolThresholdIndex, thrHigh)
	if err != nil {
		return LabelNormal, nil
	}
	if low > high {
		return "", fmt.Errorf("regime thresholds invalid: low %g > high %g", low, high)
	}
	if vol <= low {
		return LabelLow, nil
	}
	if vol >= high {
		return LabelHigh, nil
	}
	return LabelNormal, nil
}

func parseThreshold(s string, k thresholdKind) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("regime: empty %v threshold", k)
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("regime threshold %q: %w", s, err)
	}
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return 0, fmt.Errorf("regime threshold %q: not finite", s)
	}
	return f, nil
}

func volFromSnapshot(vol float64, volTS int64) (float64, bool) {
	if volTS == 0 || math.IsNaN(vol) || math.IsInf(vol, 0) {
		return vol, false
	}
	return vol, true
}
