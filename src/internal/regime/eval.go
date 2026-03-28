package regime

import (
	"strings"
	"time"

	"github.com/dfr/optitrade/src/internal/config"
	"github.com/dfr/optitrade/src/internal/market"
)

// Evaluation is one snapshot pass: raw rules output plus hysteresis-resolved regime.
type Evaluation struct {
	Outcome     Outcome
	RawLabel    Label
	VolOK       bool
	InputsDigest string
}

// EvaluateSnapshot applies rules_v1 (+ optional hysteresis) to a market snapshot.
func EvaluateSnapshot(policy *config.Policy, snap market.MarketSnapshot, h *Hysteresis, now time.Time) (Evaluation, error) {
	reg := policyRegime(policy)
	vol, volOK := volFromSnapshot(snap.VolIndex, snap.VolIndexTS)

	var raw Label
	var err error
	if reg != nil && volOK && reg.Classifier == "rules_v1" {
		raw, err = classifyRulesV1(reg, vol, true)
		if err != nil {
			return Evaluation{}, err
		}
	} else if reg != nil && !volOK && reg.Classifier == "rules_v1" {
		switch strings.TrimSpace(reg.OnMissingVol) {
		case "hold_last":
			raw = h.Effective()
		default:
			raw = LabelNormal
		}
	} else {
		raw, err = classifyRulesV1(reg, vol, volOK)
		if err != nil {
			return Evaluation{}, err
		}
	}

	minD := hysteresisDuration(reg)
	final := h.Step(raw, minD, now)

	cv := ClassifierRulesV1
	if reg == nil || strings.TrimSpace(reg.Classifier) == "" || reg.Classifier != "rules_v1" {
		cv = "passthrough/0"
	}

	digest, err := InputsDigestSHA256(snap, reg, raw, volOK)
	if err != nil {
		return Evaluation{}, err
	}

	return Evaluation{
		Outcome: Outcome{
			Label:             final,
			ClassifierVersion: cv,
		},
		RawLabel:     raw,
		VolOK:        volOK,
		InputsDigest: digest,
	}, nil
}

func policyRegime(p *config.Policy) *config.Regime {
	if p == nil {
		return nil
	}
	return p.Regime
}

func hysteresisDuration(r *config.Regime) time.Duration {
	if r == nil || r.HysteresisMinutes == nil || *r.HysteresisMinutes <= 0 {
		return 0
	}
	return time.Duration(*r.HysteresisMinutes) * time.Minute
}
