package dashboard

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dfr/optitrade/src/internal/opportunities"
)

// opportunitySnapshot returns live runner candidates for the user (nil = empty snapshot).
type opportunitySnapshot func(username string) opportunities.Snapshot

// buildOpportunitiesView returns the same JSON-shaped map as GET /opportunities (200 body).
func (s *Server) buildOpportunitiesView(ctx context.Context, user string) (map[string]any, error) {
	if s.settings == nil || s.settingsCrypto == nil {
		return nil, fmt.Errorf("settings unavailable")
	}
	row, err := s.settings.GetDecrypting(ctx, user, s.settingsCrypto)
	if err != nil {
		return nil, err
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
		return map[string]any{
			"paused":         false,
			"disabled":       true,
			"message":        "Trading is disabled for this account. Contact an administrator.",
			"updated_at_ms":  nowMs,
			"rows":           []any{},
		}, nil
	}
	if botMode == "paused" {
		return map[string]any{
			"paused":         true,
			"disabled":       false,
			"updated_at_ms":  nowMs,
			"rows":           []any{},
			"resume_hint":    "Switch bot mode away from paused to resume market scanning.",
		}, nil
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
			return nil, err
		}
		for i := range recs {
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

	return map[string]any{
		"paused":         false,
		"disabled":       false,
		"updated_at_ms":  updated,
		"rows":           merged,
	}, nil
}

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
	payload, err := s.buildOpportunitiesView(r.Context(), user)
	if err != nil {
		logHandlerError(s.log, "opportunities_get", err)
		writeAPIError(w, http.StatusInternalServerError, "server_error", "could not load opportunities")
		return
	}
	writeJSON(w, http.StatusOK, payload)
}

func (s *Server) handleOpportunitiesStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	user, ok := requestUser(r.Context())
	if !ok || s.settings == nil || s.settingsCrypto == nil {
		writeAPIError(w, http.StatusInternalServerError, "server_error", "settings unavailable")
		return
	}
	fl, ok := w.(http.Flusher)
	if !ok {
		writeAPIError(w, http.StatusInternalServerError, "server_error", "streaming not supported")
		return
	}

	payload, err := s.buildOpportunitiesView(r.Context(), user)
	if err != nil {
		logHandlerError(s.log, "opportunities_stream", err)
		writeAPIError(w, http.StatusInternalServerError, "server_error", "could not load opportunities")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	if err := writeSSEEvent(w, payload); err != nil {
		return
	}
	fl.Flush()

	tick := time.NewTicker(1 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-tick.C:
			payload, err := s.buildOpportunitiesView(r.Context(), user)
			if err != nil {
				logHandlerError(s.log, "opportunities_stream_tick", err)
				return
			}
			if err := writeSSEEvent(w, payload); err != nil {
				return
			}
			fl.Flush()
		}
	}
}

func writeSSEEvent(w http.ResponseWriter, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "data: %s\n\n", b)
	return err
}
