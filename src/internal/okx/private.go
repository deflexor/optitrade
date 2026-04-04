package okx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// AccountBalanceDetail is one row from GET /api/v5/account/balance.
type AccountBalanceDetail struct {
	Ccy      string `json:"ccy"`
	Bal      string `json:"bal"`
	AvailBal string `json:"availBal"`
	EqUsd    string `json:"eqUsd"`
}

// GetBalances returns balance rows (no ccy filter = all configured assets).
func (c *Client) GetBalances(ctx context.Context) ([]AccountBalanceDetail, error) {
	data, err := c.Do(ctx, httpMethodGET, "/api/v5/account/balance", "")
	if err != nil {
		return nil, err
	}
	var rows []AccountBalanceDetail
	if err := json.Unmarshal(data, &rows); err != nil {
		return nil, err
	}
	return rows, nil
}

// PositionRow is one position from GET /api/v5/account/positions.
type PositionRow struct {
	InstID   string `json:"instId"`
	Pos      string `json:"pos"`
	PosSide  string `json:"posSide"`
	AvgPx    string `json:"avgPx"`
	Upl      string `json:"upl"`
	MarkPx   string `json:"markPx"`
	IdxPx    string `json:"idxPx"`
	Margin   string `json:"margin"`
	Notional string `json:"notionalUsd"`
}

const httpMethodGET = "GET"
const httpMethodPOST = "POST"

// GetPositionsANY loads positions with instType=ANY.
func (c *Client) GetPositionsANY(ctx context.Context) ([]PositionRow, error) {
	q := url.Values{}
	q.Set("instType", "ANY")
	data, err := c.Do(ctx, httpMethodGET, "/api/v5/account/positions?"+q.Encode(), "")
	if err != nil {
		return nil, err
	}
	var rows []PositionRow
	if err := json.Unmarshal(data, &rows); err != nil {
		return nil, err
	}
	return rows, nil
}

// FillRow from GET /api/v5/trade/fills.
type FillRow struct {
	InstID   string `json:"instId"`
	TradeID  string `json:"tradeId"`
	OrdID    string `json:"ordId"`
	Side     string `json:"side"`
	Ts       string `json:"ts"`
	Px       string `json:"px"`
	Sz       string `json:"sz"`
	Pnl      string `json:"pnl"`
	Fee      string `json:"fee"`
	FeeCcy   string `json:"feeCcy"`
	PosSide  string `json:"posSide"`
	InstType string `json:"instType"`
}

// GetFills in a millisecond window.
func (c *Client) GetFills(ctx context.Context, beginMs, endMs int64, limit string) ([]FillRow, error) {
	q := url.Values{}
	if beginMs > 0 {
		q.Set("begin", strconv.FormatInt(beginMs, 10))
	}
	if endMs > 0 {
		q.Set("end", strconv.FormatInt(endMs, 10))
	}
	if limit != "" {
		q.Set("limit", limit)
	}
	path := "/api/v5/trade/fills?" + q.Encode()
	data, err := c.Do(ctx, httpMethodGET, path, "")
	if err != nil {
		return nil, err
	}
	var rows []FillRow
	if err := json.Unmarshal(data, &rows); err != nil {
		return nil, err
	}
	return rows, nil
}

// PlaceOrderRequest is POST /api/v5/trade/order body subset.
type PlaceOrderRequest struct {
	InstID     string `json:"instId"`
	TdMode     string `json:"tdMode"`
	Side       string `json:"side"`
	OrdType    string `json:"ordType"`
	Sz         string `json:"sz"`
	ReduceOnly bool   `json:"reduceOnly,omitempty"`
}

// PlaceOrder places an order.
func (c *Client) PlaceOrder(ctx context.Context, req PlaceOrderRequest) (json.RawMessage, error) {
	if req.TdMode == "" {
		req.TdMode = "cross"
	}
	bodyObj, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	data, err := c.Do(ctx, httpMethodPOST, "/api/v5/trade/order", string(bodyObj))
	if err != nil {
		return nil, err
	}
	var arr []json.RawMessage
	if err := json.Unmarshal(data, &arr); err != nil {
		return nil, err
	}
	if len(arr) == 0 {
		return nil, fmt.Errorf("okx place order: empty data")
	}
	return arr[0], nil
}

// BatchPlaceOrderItem is one leg in POST /api/v5/trade/batch-orders (max 20 per request).
type BatchPlaceOrderItem struct {
	InstID     string `json:"instId"`
	TdMode     string `json:"tdMode"`
	Side       string `json:"side"`
	OrdType    string `json:"ordType"`
	Sz         string `json:"sz"`
	ReduceOnly bool   `json:"reduceOnly,omitempty"`
}

type batchPlaceResult struct {
	OrdID string `json:"ordId"`
	SCode string `json:"sCode"`
	SMsg  string `json:"sMsg"`
}

// BatchPlaceOrders places multiple orders; returns OKX ordIds when sCode is "0" for each row.
func (c *Client) BatchPlaceOrders(ctx context.Context, orders []BatchPlaceOrderItem) ([]string, error) {
	if len(orders) == 0 {
		return nil, fmt.Errorf("okx: batch orders empty")
	}
	if len(orders) > 20 {
		return nil, fmt.Errorf("okx: batch orders max 20")
	}
	for i := range orders {
		if orders[i].TdMode == "" {
			orders[i].TdMode = "cross"
		}
		if orders[i].OrdType == "" {
			orders[i].OrdType = "market"
		}
		if orders[i].Sz == "" {
			orders[i].Sz = "1"
		}
	}
	body, err := json.Marshal(orders)
	if err != nil {
		return nil, err
	}
	data, err := c.Do(ctx, httpMethodPOST, "/api/v5/trade/batch-orders", string(body))
	if err != nil {
		return nil, err
	}
	var rows []batchPlaceResult
	if err := json.Unmarshal(data, &rows); err != nil {
		return nil, fmt.Errorf("okx batch-orders decode: %w", err)
	}
	if len(rows) != len(orders) {
		return nil, fmt.Errorf("okx batch-orders: got %d results for %d orders", len(rows), len(orders))
	}
	out := make([]string, 0, len(rows))
	for i := range rows {
		if strings.TrimSpace(rows[i].SCode) != "" && rows[i].SCode != "0" {
			return nil, fmt.Errorf("okx batch order %d: code %s msg %s", i, rows[i].SCode, rows[i].SMsg)
		}
		if strings.TrimSpace(rows[i].OrdID) == "" {
			return nil, fmt.Errorf("okx batch order %d: missing ordId", i)
		}
		out = append(out, rows[i].OrdID)
	}
	return out, nil
}

// BatchCancelItem is one order to cancel in POST /api/v5/trade/cancel-batch-orders.
type BatchCancelItem struct {
	InstID string `json:"instId"`
	OrdID  string `json:"ordId"`
}

type batchCancelResult struct {
	OrdID string `json:"ordId"`
	SCode string `json:"sCode"`
	SMsg  string `json:"sMsg"`
}

// BatchCancelOrders cancels up to 20 orders.
func (c *Client) BatchCancelOrders(ctx context.Context, items []BatchCancelItem) error {
	if len(items) == 0 {
		return nil
	}
	if len(items) > 20 {
		return fmt.Errorf("okx: cancel batch max 20")
	}
	body, err := json.Marshal(items)
	if err != nil {
		return err
	}
	data, err := c.Do(ctx, httpMethodPOST, "/api/v5/trade/cancel-batch-orders", string(body))
	if err != nil {
		return err
	}
	var rows []batchCancelResult
	if err := json.Unmarshal(data, &rows); err != nil {
		return fmt.Errorf("okx cancel-batch decode: %w", err)
	}
	for i := range rows {
		if strings.TrimSpace(rows[i].SCode) != "" && rows[i].SCode != "0" {
			return fmt.Errorf("okx cancel %d ord %s: code %s msg %s", i, rows[i].OrdID, rows[i].SCode, rows[i].SMsg)
		}
	}
	return nil
}

// ParseFloat is a shared helper for adapters.
func ParseFloat(s string) (*float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("empty")
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil, err
	}
	return &v, nil
}
