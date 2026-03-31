package dashboard

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/dfr/optitrade/src/internal/deribit"
	"github.com/dfr/optitrade/src/internal/okx"
)

type okxExchange struct {
	c *okx.Client
}

func newOKXExchange(row *OperatorSettingsRow) (*okxExchange, error) {
	if row == nil {
		return nil, fmt.Errorf("nil settings")
	}
	cl := &okx.Client{
		Key:        row.Secrets.OKXAPIKey,
		Secret:     row.Secrets.OKXSecretKey,
		Passphrase: row.Secrets.OKXPassphrase,
		Simulated:  row.OKXDemo,
	}
	return &okxExchange{c: cl}, nil
}

func (o *okxExchange) GetServerTime(ctx context.Context) (int64, error) {
	return o.c.GetServerTime(ctx)
}

func (o *okxExchange) GetAccountSummaries(ctx context.Context, p *deribit.GetAccountSummariesParams) ([]deribit.AccountSummary, error) {
	rows, err := o.c.GetBalances(ctx)
	if err != nil {
		return nil, err
	}
	var prefer []string
	if p != nil && p.Currency != nil && strings.TrimSpace(*p.Currency) != "" {
		prefer = []string{strings.ToUpper(strings.TrimSpace(*p.Currency))}
	}
	return okxBalancesToSummaries(rows, prefer), nil
}

func okxBalancesToSummaries(rows []okx.AccountBalanceDetail, preferCcy []string) []deribit.AccountSummary {
	if len(rows) == 0 {
		return nil
	}
	pick := map[string]struct{}{}
	for _, c := range preferCcy {
		pick[strings.ToUpper(strings.TrimSpace(c))] = struct{}{}
	}
	var out []deribit.AccountSummary
	for _, r := range rows {
		ccy := strings.ToUpper(strings.TrimSpace(r.Ccy))
		if len(pick) > 0 {
			if _, ok := pick[ccy]; !ok {
				continue
			}
		}
		bal, _ := okx.ParseFloat(r.Bal)
		eq, _ := okx.ParseFloat(r.EqUsd)
		eqVal := 0.0
		if eq != nil {
			eqVal = *eq
		}
		if eqVal == 0 && bal != nil {
			eqVal = *bal
		}
		su := deribit.AccountSummary{Currency: ccy}
		if bal != nil && *bal != 0 {
			su.Balance = bal
		}
		if eqVal != 0 {
			su.Equity = &eqVal
		}
		out = append(out, su)
	}
	return out
}

func (o *okxExchange) GetPositions(ctx context.Context, _ *deribit.GetPositionsParams) ([]deribit.Position, error) {
	rows, err := o.c.GetPositionsANY(ctx)
	if err != nil {
		return nil, err
	}
	return okxPositionsToDeribit(rows), nil
}

func okxPositionsToDeribit(rows []okx.PositionRow) []deribit.Position {
	out := make([]deribit.Position, 0, len(rows))
	for _, r := range rows {
		pos, _ := strconv.ParseFloat(strings.TrimSpace(r.Pos), 64)
		if pos == 0 {
			continue
		}
		dir := okxPosSideToDirection(r.PosSide, pos)
		p := deribit.Position{
			InstrumentName: r.InstID,
			Size:           pos,
			Direction:      strPtrOrNil(dir),
		}
		if v, err := okx.ParseFloat(r.AvgPx); err == nil {
			p.AveragePrice = v
		}
		if v, err := okx.ParseFloat(r.Upl); err == nil {
			p.FloatingProfitLoss = v
		}
		if v, err := okx.ParseFloat(r.MarkPx); err == nil {
			p.MarkPrice = v
		}
		if v, err := okx.ParseFloat(r.IdxPx); err == nil {
			p.IndexPrice = v
		}
		out = append(out, p)
	}
	return out
}

func okxPosSideToDirection(posSide string, pos float64) string {
	ps := strings.ToLower(strings.TrimSpace(posSide))
	switch ps {
	case "long":
		return "buy"
	case "short":
		return "sell"
	case "net":
		if pos > 0 {
			return "buy"
		}
		if pos < 0 {
			return "sell"
		}
	}
	if pos > 0 {
		return "buy"
	}
	return "sell"
}

func strPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (o *okxExchange) GetUserTrades(ctx context.Context, p deribit.GetUserTradesParams) ([]deribit.UserTrade, error) {
	var begin, end int64
	if p.StartTimestamp != nil {
		begin = *p.StartTimestamp
	}
	if p.EndTimestamp != nil {
		end = *p.EndTimestamp
	}
	fills, err := o.c.GetFills(ctx, begin, end, "100")
	if err != nil {
		return nil, err
	}
	return okxFillsToUserTrades(fills, p.Currency), nil
}

func okxFillsToUserTrades(fills []okx.FillRow, currency string) []deribit.UserTrade {
	cur := strings.ToUpper(strings.TrimSpace(currency))
	out := make([]deribit.UserTrade, 0, len(fills))
	for _, f := range fills {
		if cur != "" && !strings.HasPrefix(strings.ToUpper(strings.Split(f.InstID, "-")[0]), cur) {
			continue
		}
		ts, _ := strconv.ParseInt(strings.TrimSpace(f.Ts), 10, 64)
		px, _ := okx.ParseFloat(f.Px)
		sz, _ := okx.ParseFloat(f.Sz)
		pnl, _ := okx.ParseFloat(f.Pnl)
		fee, _ := okx.ParseFloat(f.Fee)
		dir := normalizeOKXSide(f.Side)
		tr := deribit.UserTrade{
			TradeID:        f.TradeID,
			OrderID:        f.OrdID,
			InstrumentName: f.InstID,
			Timestamp:      ts,
			Direction:      strPtrOrNil(dir),
			Amount:         sz,
			Price:          px,
			Fee:            fee,
			ProfitLoss:     pnl,
		}
		if f.FeeCcy != "" {
			tr.FeeCurrency = &f.FeeCcy
		}
		out = append(out, tr)
	}
	return out
}

func normalizeOKXSide(side string) string {
	switch strings.ToLower(strings.TrimSpace(side)) {
	case "buy":
		return "buy"
	case "sell":
		return "sell"
	default:
		return side
	}
}

func (o *okxExchange) Buy(ctx context.Context, params deribit.PlaceOrderParams) (*deribit.PlacedOrderResponse, error) {
	return o.place(ctx, "buy", params)
}

func (o *okxExchange) Sell(ctx context.Context, params deribit.PlaceOrderParams) (*deribit.PlacedOrderResponse, error) {
	return o.place(ctx, "sell", params)
}

func (o *okxExchange) place(ctx context.Context, side string, params deribit.PlaceOrderParams) (*deribit.PlacedOrderResponse, error) {
	if params.InstrumentName == "" {
		return nil, fmt.Errorf("okx order: missing instrument")
	}
	sz := ""
	if params.Amount != nil && math.Abs(*params.Amount) > 0 {
		sz = formatOKXSz(*params.Amount)
	} else if params.Contracts != nil {
		sz = formatOKXSz(*params.Contracts)
	}
	if sz == "" {
		return nil, fmt.Errorf("okx order: missing size")
	}
	ordType := "market"
	if params.Type != nil && strings.TrimSpace(*params.Type) != "" {
		ordType = strings.ToLower(strings.TrimSpace(*params.Type))
	}
	ro := false
	if params.ReduceOnly != nil {
		ro = *params.ReduceOnly
	}
	raw, err := o.c.PlaceOrder(ctx, okx.PlaceOrderRequest{
		InstID:     params.InstrumentName,
		Side:       side,
		OrdType:    ordType,
		Sz:         sz,
		ReduceOnly: ro,
	})
	if err != nil {
		return nil, err
	}
	return decodeOKXOrderResult(raw)
}

func formatOKXSz(v float64) string {
	s := strconv.FormatFloat(math.Abs(v), 'f', -1, 64)
	return strings.TrimRight(strings.TrimRight(s, "0"), ".")
}

func decodeOKXOrderResult(raw json.RawMessage) (*deribit.PlacedOrderResponse, error) {
	var od struct {
		OrdID   string `json:"ordId"`
		InstID  string `json:"instId"`
		State   string `json:"state"`
		Side    string `json:"side"`
		AvgPx   string `json:"avgPx"`
		AccFillSz string `json:"accFillSz"`
	}
	if err := json.Unmarshal(raw, &od); err != nil {
		return nil, err
	}
	amt, _ := okx.ParseFloat(od.AccFillSz)
	if amt == nil {
		z := 0.0
		amt = &z
	}
	px, _ := okx.ParseFloat(od.AvgPx)
	st := od.State
	return &deribit.PlacedOrderResponse{
		Order: deribit.OrderDetail{
			OrderID:        od.OrdID,
			InstrumentName: od.InstID,
			OrderState:     &st,
			Direction:      strPtrOrNil(normalizeOKXSide(od.Side)),
			Amount:         amt,
			Price:          px,
		},
		Trades: nil,
	}, nil
}
