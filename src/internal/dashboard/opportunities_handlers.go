package dashboard

import (
	"net/http"
	"strings"
	"time"

	"github.com/dfr/optitrade/src/internal/opportunities"
)

// opportunitySnapshot returns live runner candidates for the user (nil = empty snapshot).
type opportunitySnapshot func(username string) opportunities.Snapshot

func (s *Server) handleOpportunitiesGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	user, ok := requestUser(r.Context())
	if !ok || s.settings == nil || s.settingsCrypto == nil {
		writeAPIError(w, http.StatusInternalServerError, "server_error", "settings unavailable")
		return
	}
	ctx := r.Context()
	row, err := s.settings.GetDecrypting(ctx, user, s.settingsCrypto)
	if err != nil {
		logHandlerError(s.log, "opportunities_get", err)
		writeAPIError(w, http.StatusInternalServerError, "server_error", "could not load settings")
		return
	}

	nowMs := time.Now().UnixMilli()
	accountStatus := "active"
	botMode := "manual"
	if row != nil {
		if strings.TrimSpace(row.AccountStatus) != "" {
			accountStatus = row.AccountStatus
		}
		if strings.TrimSpace(row.BotMode) != "" {
			botMode = row.BotMode
		}
	}

	if accountStatus == "disabled" {
		writeJSON(w, http.StatusOK, map[string]any{
			"paused":         false,
			"disabled":       true,
			"message":        "Trading is disabled for this account. Contact an administrator.",
			"updated_at_ms":  nowMs,
			"rows":           []any{},
		})
		return
	}
	if botMode == "paused" {
		writeJSON(w, http.StatusOK, map[string]any{
			"paused":         true,
			"disabled":       false,
			"updated_at_ms":  nowMs,
			"rows":           []any{},
			"resume_hint":    "Switch bot mode away from paused to resume market scanning.",
		})
		return
	}

	var snap opportunities.Snapshot
	if s.opportunitySnapshot != nil {
		snap = s.opportunitySnapshot(user)
	}

	seen := make(map[string]struct{})
	var merged []opportunities.Row

	if s.opportunities != nil {
		recs, err := s.opportunities.ListByUser(ctx, user)
		if err != nil {
			logHandlerError(s.log, "opportunities_list", err)
			writeAPIError(w, http.StatusInternalServerError, "server_error", "could not load opportunities")
			return
		}
		for i := range recs {
			// Only merge persisted lifecycle rows; candidates come from the runner snapshot.
			st := strings.ToLower(strings.TrimSpace(recs[i].Status))
			if st == string(opportunities.StatusCandidate) {
				continue
			}
			orow, err := decodeOpportunityRow(&recs[i])
			if err != nil {
				logHandlerError(s.log, "opportunities_decode", err)
				continue
			}
			merged = append(merged, orow)
			seen[orow.ID] = struct{}{}
		}
	}

	for _, c := range snap.Rows {
		if _, dup := seen[c.ID]; dup {
			continue
		}
		merged = append(merged, c)
		seen[c.ID] = struct{}{}
	}

	updated := snap.UpdatedAtMs
	if updated == 0 {
		updated = nowMs
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"paused":         false,
		"disabled":       false,
		"updated_at_ms":  updated,
		"rows":           merged,
	})
}
