package dashboard

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

func (s *Server) handleRebalancePreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(r.Body, 1<<20))

	tok := s.previews.issue("rebalance", map[string]any{
		"proposal_version": "v0",
		"note":             "Automated rebalance suggestions are not available yet; preview returns an empty proposal for operator review only.",
	}, previewTTL)

	writeJSON(w, http.StatusOK, map[string]any{
		"suggested_adjustments": []map[string]any{},
		"projected_outcome": map[string]any{
			"summary":     "No automated deltas calculated — confirm will not send orders in this build.",
			"pnl_impact":  "0",
			"risk_impact": "unchanged",
		},
		"preview_token": tok,
		"disclaimer":    "Estimates only; rebalance does not change exposure until explicitly confirmed with a future strategy module.",
	})
}

func (s *Server) handleRebalanceConfirm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	raw, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	var body struct {
		PreviewToken string `json:"preview_token"`
	}
	_ = json.Unmarshal(raw, &body)
	tok := strings.TrimSpace(body.PreviewToken)
	data, ok := s.previews.take("rebalance", tok)
	if !ok {
		writeAPIError(w, http.StatusBadRequest, "stale_preview", "confirm requires a fresh preview_token")
		return
	}
	_ = data

	// No automated order placement for rebalance v0 — operator workflow stays read-only here.
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "no_op",
		"detail":  "Rebalance confirm acknowledged; this build does not submit rebalance orders. Wire strategy + risk gates before enabling.",
		"warning": "If future versions place orders, they will require the same preview/confirm discipline as close flows.",
	})
}
