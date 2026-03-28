package regime

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/dfr/optitrade/src/internal/config"
	"github.com/dfr/optitrade/src/internal/market"
)

type digestInputs struct {
	Vol                float64 `json:"vol"`
	VolTS              int64   `json:"vol_ts"`
	VolOK              bool    `json:"vol_ok"`
	Instrument         string  `json:"instrument"`
	Classifier         string  `json:"classifier"`
	LowThr             string  `json:"low_vol_threshold_index"`
	HighThr            string  `json:"high_vol_threshold_index"`
	RawLabel           string  `json:"raw_label"`
	OnMissingVol       string  `json:"on_missing_vol"`
	HysteresisMinutes  *int    `json:"hysteresis_minutes,omitempty"`
	BookFlags          uint8   `json:"book_flags"`
}

// InputsDigestSHA256 returns a stable hex digest of classifier inputs for regime_state.
func InputsDigestSHA256(snap market.MarketSnapshot, reg *config.Regime, raw Label, volOK bool) (string, error) {
	d := digestInputs{
		Vol:        snap.VolIndex,
		VolTS:      snap.VolIndexTS,
		VolOK:      volOK,
		Instrument: snap.Instrument,
		RawLabel:   string(raw),
		BookFlags:  uint8(snap.Flags),
	}
	if reg != nil {
		d.Classifier = reg.Classifier
		d.LowThr = reg.LowVolThresholdIndex
		d.HighThr = reg.HighVolThresholdIndex
		d.OnMissingVol = reg.OnMissingVol
		d.HysteresisMinutes = reg.HysteresisMinutes
	}
	b, err := json.Marshal(d)
	if err != nil {
		return "", fmt.Errorf("regime digest: %w", err)
	}
	h := sha256.Sum256(b)
	return fmt.Sprintf("%x", h[:]), nil
}
