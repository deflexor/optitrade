package dashboard

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dfr/optitrade/src/internal/opportunities"
	"github.com/dfr/optitrade/src/internal/okx"
)

func (s *Server) okxExchangeForOpportunities(ctx context.Context, user string) (*okxExchange, error) {
	if s.testXchg != nil {
		return nil, fmt.Errorf("batch trading not available with test exchange")
	}
	if s.settings == nil || s.settingsCrypto == nil {
		return nil, fmt.Errorf("settings unavailable")
	}
	row, err := s.settings.GetDecrypting(ctx, user, s.settingsCrypto)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, fmt.Errorf("configure exchange settings first")
	}
	if row.Provider != ProviderOKX {
		return nil, fmt.Errorf("opportunity execution requires OKX provider")
	}
	x, err := newOKXExchange(row)
	if err != nil {
		return nil, err
	}
	if s != nil {
		if strings.TrimSpace(s.okxExecBaseURL) != "" {
			x.c.BaseURL = strings.TrimSpace(s.okxExecBaseURL)
		}
		if s.okxExecHTTP != nil {
			x.c.HTTP = s.okxExecHTTP
		}
	}
	return x, nil
}

func findSnapshotCandidateRow(s *Server, user, id string) (*opportunities.Row, bool) {
	if s.opportunitySnapshot == nil {
		return nil, false
	}
	snap := s.opportunitySnapshot(user)
	for i := range snap.Rows {
		if snap.Rows[i].ID != id {
			continue
		}
		if snap.Rows[i].Status != opportunities.StatusCandidate {
			continue
		}
		r := snap.Rows[i]
		return &r, true
	}
	return nil, false
}

func legSidesForCreditSpread(row *opportunities.Row) ([]legSideMeta, error) {
	if row == nil {
		return nil, fmt.Errorf("nil row")
	}
	if !strings.EqualFold(strings.TrimSpace(row.StrategyName), "credit_spread") {
		return nil, fmt.Errorf("unsupported strategy %q (M1: credit_spread only)", row.StrategyName)
	}
	if len(row.Legs) != 2 {
		return nil, fmt.Errorf("credit spread needs 2 legs, got %d", len(row.Legs))
	}
	return []legSideMeta{
		{Instrument: strings.TrimSpace(row.Legs[0].Instrument), Side: "sell"},
		{Instrument: strings.TrimSpace(row.Legs[1].Instrument), Side: "buy"},
	}, nil
}

func oppositeSide(side string) string {
	switch strings.ToLower(strings.TrimSpace(side)) {
	case "buy":
		return "sell"
	default:
		return "buy"
	}
}

func tradingPausedOrDisabled(row *OperatorSettingsRow) (blocked bool, code, msg string) {
	if row == nil {
		return false, "", ""
	}
	switch strings.ToLower(strings.TrimSpace(row.AccountStatus)) {
	case "disabled":
		return true, "account_disabled", "trading is disabled for this account"
	}
	if strings.EqualFold(strings.TrimSpace(row.BotMode), "paused") {
		return true, "bot_paused", "switch bot mode away from paused to trade"
	}
	return false, "", ""
}

// maybeCancelVenueOrdersThenDelete cancels OKX orders when meta has order_ids aligned with leg_sides, then removes the DB row.
func (s *Server) maybeCancelVenueOrdersThenDelete(ctx context.Context, user, id string, rec *OpportunityRecord) error {
	var meta opportunityMetaJSON
	if err := json.Unmarshal([]byte(rec.MetaJSON), &meta); err != nil {
		return fmt.Errorf("corrupt meta: %w", err)
	}
	if len(meta.OrderIDs) > 0 && len(meta.LegSides) == len(meta.OrderIDs) {
		xchg, err := s.okxExchangeForOpportunities(ctx, user)
		if err != nil {
			return fmt.Errorf("exchange: %w", err)
		}
		items := make([]okx.BatchCancelItem, 0, len(meta.OrderIDs))
		for i := range meta.OrderIDs {
			items = append(items, okx.BatchCancelItem{
				InstID: meta.LegSides[i].Instrument,
				OrdID:  meta.OrderIDs[i],
			})
		}
		rpcCtx, cancel := rpcTimeout(ctx)
		defer cancel()
		if err := xchg.BatchCancelOrders(rpcCtx, items); err != nil {
			return err
		}
	}
	if err := s.opportunities.Delete(ctx, id, user); err != nil {
		return fmt.Errorf("delete opportunity: %w", err)
	}
	return nil
}

func (s *Server) handleOpportunityOpen(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	user, ok := requestUser(r.Context())
	if !ok || s.opportunities == nil || s.settings == nil || s.settingsCrypto == nil {
		writeAPIError(w, http.StatusInternalServerError, "server_error", "opportunities unavailable")
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "missing opportunity id")
		return
	}
	ctx := r.Context()
	setRow, err := s.settings.GetDecrypting(ctx, user, s.settingsCrypto)
	if err != nil {
		logHandlerError(s.log, "opportunity_open_settings", err)
		writeAPIError(w, http.StatusInternalServerError, "server_error", "could not load settings")
		return
	}
	if blocked, code, msg := tradingPausedOrDisabled(setRow); blocked {
		writeAPIError(w, http.StatusConflict, code, msg)
		return
	}

	existing, err := s.opportunities.Get(ctx, id, user)
	if err != nil {
		logHandlerError(s.log, "opportunity_open_get", err)
		writeAPIError(w, http.StatusInternalServerError, "server_error", "could not load opportunity")
		return
	}
	if existing != nil {
		st := strings.ToLower(strings.TrimSpace(existing.Status))
		if st == string(opportunities.StatusOpening) || st == string(opportunities.StatusActive) || st == string(opportunities.StatusPartial) {
			writeJSON(w, http.StatusOK, map[string]any{"id": id, "status": existing.Status, "idempotent": true})
			return
		}
	}

	cand, ok := findSnapshotCandidateRow(s, user, id)
	if !ok {
		writeAPIError(w, http.StatusNotFound, "not_found", "unknown opportunity or not a candidate")
		return
	}

	legs, err := legSidesForCreditSpread(cand)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	xchg, err := s.okxExchangeForOpportunities(ctx, user)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	orders := make([]okx.BatchPlaceOrderItem, 0, len(legs))
	for _, lg := range legs {
		orders = append(orders, okx.BatchPlaceOrderItem{
			InstID:  lg.Instrument,
			Side:    strings.ToLower(lg.Side),
			OrdType: "market",
			Sz:      "1",
		})
	}
	rpcCtx, cancel := rpcTimeout(ctx)
	defer cancel()
	ordIDs, err := xchg.BatchPlaceOrders(rpcCtx, orders)
	if err != nil {
		logHandlerError(s.log, "opportunity_open_batch", err)
		writeAPIError(w, http.StatusBadGateway, "exchange_error", err.Error())
		return
	}

	persist := *cand
	persist.Status = opportunities.StatusOpening
	exec := &opportunityExecPersist{OrderIDs: ordIDs, LegSides: legs}
	legsJ, metaJ, err := encodeOpportunityRowPersist(persist, exec)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "server_error", "encode opportunity")
		return
	}
	rec := &OpportunityRecord{
		ID:           id,
		Username:     user,
		Status:       string(persist.Status),
		StrategyName: persist.StrategyName,
		LegsJSON:     legsJ,
		MetaJSON:     metaJ,
		CreatedAtMs:  time.Now().UnixMilli(),
	}
	if err := s.opportunities.Upsert(ctx, rec); err != nil {
		logHandlerError(s.log, "opportunity_open_upsert", err)
		writeAPIError(w, http.StatusInternalServerError, "server_error", "could not save opportunity")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": id, "status": string(persist.Status), "order_ids": ordIDs})
}

func classifyOpportunityVenueErr(err error) (status int, code, msg string) {
	if err == nil {
		return 0, "", ""
	}
	s := err.Error()
	if strings.HasPrefix(s, "corrupt meta:") {
		return http.StatusInternalServerError, "server_error", "corrupt opportunity meta"
	}
	if strings.HasPrefix(s, "exchange:") {
		return http.StatusBadRequest, "invalid_request", strings.TrimSpace(strings.TrimPrefix(s, "exchange:"))
	}
	if strings.HasPrefix(s, "delete opportunity:") {
		return http.StatusInternalServerError, "server_error", "could not remove opportunity"
	}
	return http.StatusBadGateway, "exchange_error", s
}

func (s *Server) handleOpportunityCancel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	user, ok := requestUser(r.Context())
	if !ok || s.opportunities == nil {
		writeAPIError(w, http.StatusInternalServerError, "server_error", "opportunities unavailable")
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "missing opportunity id")
		return
	}
	ctx := r.Context()
	rec, err := s.opportunities.Get(ctx, id, user)
	if err != nil {
		logHandlerError(s.log, "opportunity_cancel_get", err)
		writeAPIError(w, http.StatusInternalServerError, "server_error", "could not load opportunity")
		return
	}
	if rec == nil {
		writeAPIError(w, http.StatusNotFound, "not_found", "unknown opportunity")
		return
	}
	if strings.ToLower(strings.TrimSpace(rec.Status)) != string(opportunities.StatusOpening) {
		writeAPIError(w, http.StatusConflict, "invalid_state", "can only cancel while opening")
		return
	}

	if err := s.maybeCancelVenueOrdersThenDelete(ctx, user, id, rec); err != nil {
		logHandlerError(s.log, "opportunity_cancel", err)
		cst, cc, cm := classifyOpportunityVenueErr(err)
		writeAPIError(w, cst, cc, cm)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": id, "status": "cancelled"})
}

func (s *Server) handleOpportunityClose(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	user, ok := requestUser(r.Context())
	if !ok || s.opportunities == nil {
		writeAPIError(w, http.StatusInternalServerError, "server_error", "opportunities unavailable")
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "missing opportunity id")
		return
	}
	ctx := r.Context()
	rec, err := s.opportunities.Get(ctx, id, user)
	if err != nil {
		logHandlerError(s.log, "opportunity_close_get", err)
		writeAPIError(w, http.StatusInternalServerError, "server_error", "could not load opportunity")
		return
	}
	if rec == nil {
		writeAPIError(w, http.StatusNotFound, "not_found", "unknown opportunity")
		return
	}
	st := strings.ToLower(strings.TrimSpace(rec.Status))
	if st == string(opportunities.StatusOpening) {
		if err := s.maybeCancelVenueOrdersThenDelete(ctx, user, id, rec); err != nil {
			logHandlerError(s.log, "opportunity_close_opening", err)
			cst, cc, cm := classifyOpportunityVenueErr(err)
			writeAPIError(w, cst, cc, cm)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"id": id, "status": "closed"})
		return
	}
	if st != string(opportunities.StatusActive) && st != string(opportunities.StatusPartial) {
		writeAPIError(w, http.StatusConflict, "invalid_state", "can only close active or partial opportunities")
		return
	}

	row, err := decodeOpportunityRow(rec)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "server_error", "corrupt opportunity")
		return
	}
	var meta opportunityMetaJSON
	_ = json.Unmarshal([]byte(rec.MetaJSON), &meta)

	var legs []legSideMeta
	if len(meta.LegSides) >= 2 {
		legs = meta.LegSides
	} else {
		legs, err = legSidesForCreditSpread(&row)
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", err.Error())
			return
		}
	}

	xchg, err := s.okxExchangeForOpportunities(ctx, user)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	orders := make([]okx.BatchPlaceOrderItem, 0, len(legs))
	for _, lg := range legs {
		orders = append(orders, okx.BatchPlaceOrderItem{
			InstID:  lg.Instrument,
			Side:    oppositeSide(lg.Side),
			OrdType: "market",
			Sz:      "1",
		})
	}
	rpcCtx, cancel := rpcTimeout(ctx)
	defer cancel()
	if _, err := xchg.BatchPlaceOrders(rpcCtx, orders); err != nil {
		logHandlerError(s.log, "opportunity_close_batch", err)
		writeAPIError(w, http.StatusBadGateway, "exchange_error", err.Error())
		return
	}
	if err := s.opportunities.Delete(ctx, id, user); err != nil {
		logHandlerError(s.log, "opportunity_close_delete", err)
		writeAPIError(w, http.StatusInternalServerError, "server_error", "could not remove opportunity")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": id, "status": "closed"})
}
