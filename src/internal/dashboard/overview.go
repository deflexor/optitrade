package dashboard

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dfr/optitrade/src/internal/deribit"
)

func dashboardCurrencies() []string {
	c := strings.TrimSpace(os.Getenv("OPTITRADE_DASHBOARD_CURRENCIES"))
	if c == "" {
		return []string{"BTC", "ETH"}
	}
	var out []string
	for _, p := range strings.Split(c, ",") {
		p = strings.TrimSpace(strings.ToUpper(p))
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return []string{"BTC", "ETH"}
	}
	return out
}

func (s *Server) handleOverview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	to := time.Now().UTC()
	from := to.AddDate(0, 0, -30)
	fromDay := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
	toDay := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, time.UTC)

	out := map[string]any{
		"account": map[string]any{
			"currency":          "BTC",
			"equity":            "0",
			"balance":           "0",
			"available_funds":   nil,
			"exchange_degraded": false,
		},
		"pnl_series": map[string]any{
			"points": []map[string]string{},
			"window": map[string]string{
				"from": fromDay.Format("2006-01-02"),
				"to":   toDay.Format("2006-01-02"),
			},
		},
		"market_mood": map[string]any{
			"label":       "",
			"score":       nil,
			"explanation": "strategy modules not wired to dashboard yet",
			"available":   false,
		},
		"strategy": map[string]any{
			"expected_pnl": map[string]any{
				"horizon_days": 30,
				"low":          "0",
				"mid":          "0",
				"high":         "0",
			},
			"win_rate":          nil,
			"win_rate_defined":  false,
			"available":         false,
			"message":           "strategy metadata not available",
		},
	}

	if s.xchg == nil {
		acc := out["account"].(map[string]any)
		acc["exchange_degraded"] = true
		writeJSON(w, http.StatusOK, out)
		return
	}

	ctx, cancel := rpcTimeout(r.Context())
	defer cancel()

	var eq, bal *float64
	var cur string
	for _, ccy := range dashboardCurrencies() {
		cc := ccy
		sums, err := s.xchg.GetAccountSummaries(ctx, &deribit.GetAccountSummariesParams{Currency: &cc})
		if err != nil || len(sums) == 0 {
			continue
		}
		su := sums[0]
		cur = su.Currency
		eq = su.Equity
		bal = su.Balance
		break
	}

	acc := out["account"].(map[string]any)
	if cur != "" {
		acc["currency"] = cur
	}
	if eq != nil {
		acc["equity"] = decStr(eq)
	}
	if bal != nil {
		acc["balance"] = decStr(bal)
	} else if eq != nil {
		acc["balance"] = decStr(eq)
	}

	startMs := from.UnixMilli()
	endMs := to.UnixMilli()
	sorting := "asc"
	dayPnL := map[string]float64{}
	degraded := false
	for _, ccy := range dashboardCurrencies() {
		trades, err := s.xchg.GetUserTrades(ctx, deribit.GetUserTradesParams{
			Currency:       ccy,
			StartTimestamp: &startMs,
			EndTimestamp:   &endMs,
			Sorting:        &sorting,
			Historical:     ptrBool(true),
		})
		if err != nil {
			degraded = true
			continue
		}
		for _, tr := range trades {
			if tr.ProfitLoss == nil {
				continue
			}
			day := time.UnixMilli(tr.Timestamp).UTC().Format("2006-01-02")
			dayPnL[day] += *tr.ProfitLoss
		}
	}
	if degraded {
		acc["exchange_degraded"] = true
	}

	points := []map[string]string{}
	cumulative := 0.0
	for t := fromDay; !t.After(toDay); t = t.AddDate(0, 0, 1) {
		day := t.Format("2006-01-02")
		cumulative += dayPnL[day]
		points = append(points, map[string]string{
			"t":         t.UTC().Format(time.RFC3339),
			"pnl_quote": strconv.FormatFloat(cumulative, 'f', -1, 64),
		})
	}
	series := out["pnl_series"].(map[string]any)
	series["points"] = points

	writeJSON(w, http.StatusOK, out)
}

func ptrBool(b bool) *bool { return &b }
