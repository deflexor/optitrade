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
