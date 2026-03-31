package dashboard

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dfr/optitrade/src/internal/deribit"
)

const previewTTL = 5 * time.Minute

func (s *Server) handleClosePreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id := r.PathValue("id")
	if id == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "missing position")
		return
	}
	xchg, err := s.exchangeForRequest(r.Context())
	if err != nil {
		logHandlerError(s.log, "close_preview", err)
		writeAPIError(w, http.StatusInternalServerError, "server_error", "could not resolve exchange")
		return
	}
	if xchg == nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "exchange not configured")
		return
	}
	inst, dir, ok := parsePositionRowID(id)
	if !ok {
		writeAPIError(w, http.StatusNotFound, "not_found", "unknown position")
		return
	}
	ctx, cancel := rpcTimeout(r.Context())
	defer cancel()
	rows, err := xchg.GetPositions(ctx, &deribit.GetPositionsParams{})
	if err != nil {
		writeAPIError(w, http.StatusBadGateway, "exchange_error", "could not load positions")
		return
	}
	var found *deribit.Position
	for i := range rows {
		p := &rows[i]
		if p.InstrumentName != inst || p.Size == 0 {
			continue
		}
		if dir != "" && derefStr(p.Direction) != dir {
			continue
		}
		found = p
		break
	}
	if found == nil {
		writeAPIError(w, http.StatusNotFound, "not_found", "unknown position")
		return
	}

	est := decStr(found.FloatingProfitLoss)
	tok := s.previews.issue("close", map[string]any{
		"position_id": id,
		"instrument":  inst,
		"direction":   derefStr(found.Direction),
		"size":        found.Size,
	}, previewTTL)

	writeJSON(w, http.StatusOK, map[string]any{
		"estimated_exit_pnl":       est,
		"wait_vs_close_guidance":   "Review mark vs. index and spread before confirming. Execution uses reduce-only placement; fills are not guaranteed at estimate.",
		"assumptions":              []string{"Estimate uses floating P/L from the exchange position row.", "Fees are not modeled in this preview."},
		"preview_token":            tok,
		"position_id_echo":        id,
		"labels":                   []string{"Estimates only — not a guarantee of fill price or slippage."},
	})
}

func (s *Server) handleCloseConfirm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id := r.PathValue("id")
	raw, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	var body struct {
		PreviewToken string `json:"preview_token"`
	}
	_ = json.Unmarshal(raw, &body)
	tok := strings.TrimSpace(body.PreviewToken)
	data, ok := s.previews.take("close", tok)
	if !ok {
		writeAPIError(w, http.StatusBadRequest, "stale_preview", "confirm requires a fresh preview_token")
		return
	}
	echo, _ := data["position_id"].(string)
	if echo != id {
		writeAPIError(w, http.StatusBadRequest, "position_mismatch", "preview does not match this position")
		return
	}

	xchg, err := s.exchangeForRequest(r.Context())
	if err != nil {
		logHandlerError(s.log, "close_confirm", err)
		writeAPIError(w, http.StatusInternalServerError, "server_error", "could not resolve exchange")
		return
	}
	if xchg == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "exchange_unavailable", "exchange not configured")
		return
	}
	inst, dir, ok := parsePositionRowID(id)
	if !ok {
		writeAPIError(w, http.StatusNotFound, "not_found", "unknown position")
		return
	}
	ctx, cancel := rpcTimeout(r.Context())
	defer cancel()
	rows, err := xchg.GetPositions(ctx, &deribit.GetPositionsParams{})
	if err != nil {
		writeAPIError(w, http.StatusBadGateway, "exchange_error", "could not load positions")
		return
	}
	var found *deribit.Position
	for i := range rows {
		p := &rows[i]
		if p.InstrumentName != inst || p.Size == 0 {
			continue
		}
		if dir != "" && derefStr(p.Direction) != dir {
			continue
		}
		found = p
		break
	}
	if found == nil {
		writeAPIError(w, http.StatusNotFound, "not_found", "position closed or unknown")
		return
	}

	// Reduce-only market-style flatten: Deribit requires explicit buy/sell with amount.
	amt := found.Size
	if amt < 0 {
		amt = -amt
	}
	typ := "market"
	ro := true
	params := deribit.PlaceOrderParams{
		InstrumentName: inst,
		Amount:         &amt,
		Type:           &typ,
		ReduceOnly:     &ro,
	}
	ow, hasW := xchg.(exchangeWriter)
	if !hasW {
		writeAPIError(w, http.StatusInternalServerError, "server_error", "close execution unavailable")
		return
	}
	var resp *deribit.PlacedOrderResponse
	var cerr error
	if strings.EqualFold(derefStr(found.Direction), "buy") {
		resp, cerr = ow.Sell(ctx, params)
	} else {
		resp, cerr = ow.Buy(ctx, params)
	}
	if cerr != nil {
		logHandlerError(s.log, "close_confirm", cerr)
		writeAPIError(w, http.StatusBadGateway, "exchange_error", "order rejected or failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "submitted",
		"result":  resp,
		"warning": "Verify fills and residual size in the exchange UI; reduce-only market orders may still partially fill.",
	})
}
