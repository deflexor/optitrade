package dashboard

import (
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dfr/optitrade/src/internal/deribit"
)

func positionRowID(p deribit.Position) string {
	dir := ""
	if p.Direction != nil {
		dir = *p.Direction
	}
	return url.QueryEscape(p.InstrumentName) + "|" + url.QueryEscape(dir)
}

func parsePositionRowID(raw string) (instrument, direction string, ok bool) {
	parts := strings.SplitN(raw, "|", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	inst, err := url.QueryUnescape(parts[0])
	if err != nil {
		return "", "", false
	}
	dir, err := url.QueryUnescape(parts[1])
	if err != nil {
		return "", "", false
	}
	return inst, dir, true
}

func (s *Server) handleOpenPositions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if s.xchg == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "exchange_unavailable", "exchange not configured")
		return
	}
	limit := 25
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 25 {
			limit = n
		}
	}
	cursorOffset := 0
	if c := r.URL.Query().Get("cursor"); c != "" {
		if n, err := strconv.Atoi(c); err == nil && n >= 0 {
			cursorOffset = n
		}
	}

	ctx, cancel := rpcTimeout(r.Context())
	defer cancel()

	rows, err := s.xchg.GetPositions(ctx, &deribit.GetPositionsParams{})
	if err != nil {
		logHandlerError(s.log, "get_positions", err)
		writeAPIError(w, http.StatusBadGateway, "exchange_error", "could not load positions")
		return
	}

	var open []deribit.Position
	for _, p := range rows {
		if p.Size != 0 {
			open = append(open, p)
		}
	}
	total := len(open)
	if cursorOffset > total {
		cursorOffset = total
	}
	end := cursorOffset + limit
	if end > total {
		end = total
	}
	page := open[cursorOffset:end]

	outRows := make([]map[string]any, 0, len(page))
	for _, p := range page {
		outRows = append(outRows, map[string]any{
			"id":                  positionRowID(p),
			"instrument_summary": p.InstrumentName,
			"open":                true,
			"direction":           derefStr(p.Direction),
			"size":                p.Size,
			"quote_pnl":           decStr(p.FloatingProfitLoss),
			"usd_pnl":             decStr(p.FloatingProfitLoss),
			"legs": []map[string]any{
				{
					"instrument_name": p.InstrumentName,
					"size":            p.Size,
					"direction":       derefStr(p.Direction),
				},
			},
			"metrics": map[string]any{
				"mark_price": decStr(p.MarkPrice),
				"index_price": decStr(p.IndexPrice),
			},
			"greeks": greeksMap(p),
		})
	}

	nextCursor := ""
	if end < total {
		nextCursor = strconv.Itoa(end)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items":        outRows,
		"next_cursor":  nextCursor,
		"total_count":  total,
	})
}

func (s *Server) handleClosedPositions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if s.xchg == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "exchange_unavailable", "exchange not configured")
		return
	}

	to := time.Now().UTC()
	from := to.AddDate(0, 0, -30)
	startMs := from.UnixMilli()
	endMs := to.UnixMilli()
	sorting := "desc"

	ctx, cancel := rpcTimeout(r.Context())
	defer cancel()

	type row struct {
		tr deribit.UserTrade
	}
	var all []row
	for _, ccy := range dashboardCurrencies() {
		trades, err := s.xchg.GetUserTrades(ctx, deribit.GetUserTradesParams{
			Currency:       ccy,
			StartTimestamp: &startMs,
			EndTimestamp:   &endMs,
			Sorting:        &sorting,
			Historical:     ptrBool(true),
		})
		if err != nil {
			logHandlerError(s.log, "get_user_trades", err)
			writeAPIError(w, http.StatusBadGateway, "exchange_error", "could not load closed trades")
			return
		}
		for _, tr := range trades {
			if tr.ProfitLoss != nil && *tr.ProfitLoss != 0 {
				all = append(all, row{tr: tr})
			}
		}
	}

	sort.Slice(all, func(i, j int) bool { return all[i].tr.Timestamp > all[j].tr.Timestamp })

	capN := 200
	if len(all) > capN {
		all = all[:capN]
	}

	items := make([]map[string]any, 0, len(all))
	for _, rw := range all {
		tr := rw.tr
		pl := decStr(tr.ProfitLoss)
		items = append(items, map[string]any{
			"id":                   tr.TradeID,
			"instrument_summary":   tr.InstrumentName,
			"closed_at":            time.UnixMilli(tr.Timestamp).UTC().Format(time.RFC3339),
			"realized_pnl_usd":     pl,
			"realized_pnl_pct":     nil,
			"percent_basis":        "not_applicable",
			"percent_basis_label":  "Realized P/L (quote); % basis unavailable from trade feed",
		})
	}

	truncated := len(items) >= capN
	writeJSON(w, http.StatusOK, map[string]any{
		"items":     items,
		"truncated": truncated,
	})
}

func (s *Server) handlePositionDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id := r.PathValue("id")
	if id == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_id", "missing id")
		return
	}
	if s.xchg == nil {
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
	rows, err := s.xchg.GetPositions(ctx, &deribit.GetPositionsParams{})
	if err != nil {
		writeAPIError(w, http.StatusBadGateway, "exchange_error", "could not load positions")
		return
	}
	var found *deribit.Position
	for i := range rows {
		p := &rows[i]
		if p.InstrumentName != inst {
			continue
		}
		pdir := derefStr(p.Direction)
		if dir != "" && pdir != dir {
			continue
		}
		if p.Size == 0 {
			continue
		}
		found = p
		break
	}
	if found == nil {
		writeAPIError(w, http.StatusNotFound, "not_found", "unknown position")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"id":                  positionRowID(*found),
		"instrument_summary": found.InstrumentName,
		"open":                true,
		"direction":           derefStr(found.Direction),
		"size":                found.Size,
		"quote_pnl":           decStr(found.FloatingProfitLoss),
		"usd_pnl":             decStr(found.FloatingProfitLoss),
		"legs": []map[string]any{
			{
				"instrument_name": found.InstrumentName,
				"size":            found.Size,
				"direction":       derefStr(found.Direction),
			},
		},
		"metrics": map[string]any{
			"mark_price":    decStr(found.MarkPrice),
			"index_price":   decStr(found.IndexPrice),
			"initial_margin": decStr(found.InitialMargin),
			"maintenance_margin": decStr(found.MaintenanceMargin),
			"liquidity_note": "order book depth not loaded in dashboard v0",
		},
		"greeks": greeksMap(*found),
	})
}

func greeksMap(p deribit.Position) map[string]any {
	if p.Delta == nil && p.Gamma == nil && p.Vega == nil && p.Theta == nil {
		return map[string]any{"available": false, "note": "Greeks not provided for this instrument"}
	}
	return map[string]any{
		"available": true,
		"delta":     decStr(p.Delta),
		"gamma":     decStr(p.Gamma),
		"vega":      decStr(p.Vega),
		"theta":     decStr(p.Theta),
	}
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
