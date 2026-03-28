package regime

import "time"

// Hysteresis carries bounded state for rules_v1 flip confirmation.
type Hysteresis struct {
	effective Label
	pending   Label
	since     time.Time
}

// Effective returns the current post-hysteresis label (default normal before first Step).
func (h *Hysteresis) Effective() Label {
	if h == nil || h.effective == "" {
		return LabelNormal
	}
	return h.effective
}

// Step applies flip confirmation: when minDuration is zero, raw is adopted immediately.
func (h *Hysteresis) Step(raw Label, minDuration time.Duration, now time.Time) Label {
	if h == nil {
		return raw
	}
	if h.effective == "" {
		h.effective = LabelNormal
	}
	if minDuration <= 0 || now.IsZero() {
		h.effective = raw
		h.pending = ""
		h.since = time.Time{}
		return h.effective
	}
	if raw == h.effective {
		h.pending = ""
		h.since = time.Time{}
		return h.effective
	}
	if h.pending != raw {
		h.pending = raw
		h.since = now
		return h.effective
	}
	if h.since.IsZero() {
		h.since = now
	}
	if now.Sub(h.since) >= minDuration {
		h.effective = raw
		h.pending = ""
		h.since = time.Time{}
	}
	return h.effective
}
