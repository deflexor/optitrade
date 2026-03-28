package deribit

import (
	"context"
	"encoding/json"

	"github.com/dfr/optitrade/src/internal/deribit/rpc"
)

// REST is a typed facade over Deribit JSON-RPC. Private methods require non-nil Credentials in NewREST.
type REST struct {
	public  *rpc.Client
	private *rpc.Client
	tokens  *tokenManager
}

// NewREST builds a REST client. creds may be nil for public endpoints only.
func NewREST(baseURL string, creds *Credentials) (*REST, error) {
	pub := rpc.NewClient(baseURL, nil)
	if creds == nil {
		return &REST{public: pub, private: nil, tokens: nil}, nil
	}
	tm := newTokenManager(baseURL, *creds)
	priv := rpc.NewClient(baseURL, tm.authorize)
	return &REST{public: pub, private: priv, tokens: tm}, nil
}

func (r *REST) priv() (*rpc.Client, error) {
	if r.private == nil {
		return nil, ErrNoCredentials
	}
	return r.private, nil
}

// GetPositions calls private/get_positions.
func (r *REST) GetPositions(ctx context.Context, params *GetPositionsParams) ([]Position, error) {
	c, err := r.priv()
	if err != nil {
		return nil, err
	}
	var p any = map[string]any{}
	if params != nil {
		p = params
	}
	raw, err := c.Call(ctx, "private/get_positions", p)
	if err != nil {
		return nil, err
	}
	var out []Position
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetOpenOrders calls private/get_open_orders.
func (r *REST) GetOpenOrders(ctx context.Context, params *GetOpenOrdersParams) ([]OpenOrder, error) {
	c, err := r.priv()
	if err != nil {
		return nil, err
	}
	var p any = map[string]any{}
	if params != nil {
		p = params
	}
	raw, err := c.Call(ctx, "private/get_open_orders", p)
	if err != nil {
		return nil, err
	}
	var out []OpenOrder
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetAccountSummaries calls private/get_account_summaries.
func (r *REST) GetAccountSummaries(ctx context.Context, params *GetAccountSummariesParams) ([]AccountSummary, error) {
	c, err := r.priv()
	if err != nil {
		return nil, err
	}
	var p any = map[string]any{}
	if params != nil {
		p = params
	}
	raw, err := c.Call(ctx, "private/get_account_summaries", p)
	if err != nil {
		return nil, err
	}
	var out []AccountSummary
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetInstruments calls public/get_instruments.
func (r *REST) GetInstruments(ctx context.Context, params *GetInstrumentsParams) ([]Instrument, error) {
	var p any = map[string]any{}
	if params != nil {
		p = params
	}
	raw, err := r.public.Call(ctx, "public/get_instruments", p)
	if err != nil {
		return nil, err
	}
	var out []Instrument
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetVolatilityIndexData calls public/get_volatility_index_data.
func (r *REST) GetVolatilityIndexData(ctx context.Context, params GetVolatilityIndexDataParams) (*VolatilityIndexData, error) {
	raw, err := r.public.Call(ctx, "public/get_volatility_index_data", params)
	if err != nil {
		return nil, err
	}
	var out VolatilityIndexData
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetOrderBook calls public/get_order_book.
func (r *REST) GetOrderBook(ctx context.Context, params GetOrderBookParams) (*OrderBook, error) {
	raw, err := r.public.Call(ctx, "public/get_order_book", params)
	if err != nil {
		return nil, err
	}
	var ob OrderBook
	if err := json.Unmarshal(raw, &ob); err != nil {
		return nil, err
	}
	return &ob, nil
}

// GetTicker calls public/ticker.
func (r *REST) GetTicker(ctx context.Context, params TickerParams) (*Ticker, error) {
	raw, err := r.public.Call(ctx, "public/ticker", params)
	if err != nil {
		return nil, err
	}
	var t Ticker
	if err := json.Unmarshal(raw, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

// GetServerTime calls public/get_time (useful connectivity smoke test).
func (r *REST) GetServerTime(ctx context.Context) (int64, error) {
	raw, err := r.public.Call(ctx, "public/get_time", map[string]any{})
	if err != nil {
		return 0, err
	}
	// Deribit returns unix ms number
	var ms int64
	if err := json.Unmarshal(raw, &ms); err != nil {
		return 0, err
	}
	return ms, nil
}

// Buy calls private/buy.
func (r *REST) Buy(ctx context.Context, params PlaceOrderParams) (*PlacedOrderResponse, error) {
	return r.placeOrder(ctx, "private/buy", params)
}

// Sell calls private/sell.
func (r *REST) Sell(ctx context.Context, params PlaceOrderParams) (*PlacedOrderResponse, error) {
	return r.placeOrder(ctx, "private/sell", params)
}

func (r *REST) placeOrder(ctx context.Context, method string, params PlaceOrderParams) (*PlacedOrderResponse, error) {
	c, err := r.priv()
	if err != nil {
		return nil, err
	}
	if params.InstrumentName == "" {
		return nil, errMissingInstrument
	}
	raw, err := c.Call(ctx, method, params)
	if err != nil {
		return nil, err
	}
	var out PlacedOrderResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CancelAllByInstrument calls private/cancel_all_by_instrument (count of cancelled orders).
func (r *REST) CancelAllByInstrument(ctx context.Context, params CancelAllByInstrumentParams) (int, error) {
	c, err := r.priv()
	if err != nil {
		return 0, err
	}
	if params.InstrumentName == "" {
		return 0, errMissingInstrument
	}
	raw, err := c.Call(ctx, "private/cancel_all_by_instrument", params)
	if err != nil {
		return 0, err
	}
	return unmarshalIntish(raw)
}

// CancelByLabel calls private/cancel_by_label.
func (r *REST) CancelByLabel(ctx context.Context, params CancelByLabelParams) (int, error) {
	c, err := r.priv()
	if err != nil {
		return 0, err
	}
	if params.Label == "" {
		return 0, errMissingLabel
	}
	raw, err := c.Call(ctx, "private/cancel_by_label", params)
	if err != nil {
		return 0, err
	}
	return unmarshalIntish(raw)
}

// GetOrderState calls private/get_order_state (trade:read).
func (r *REST) GetOrderState(ctx context.Context, orderID string) (*OrderDetail, error) {
	c, err := r.priv()
	if err != nil {
		return nil, err
	}
	if orderID == "" {
		return nil, errMissingOrderID
	}
	raw, err := c.Call(ctx, "private/get_order_state", map[string]string{"order_id": orderID})
	if err != nil {
		return nil, err
	}
	var out OrderDetail
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetUserTradesByOrder calls private/get_user_trades_by_order.
func (r *REST) GetUserTradesByOrder(ctx context.Context, params GetUserTradesByOrderParams) ([]UserTrade, error) {
	c, err := r.priv()
	if err != nil {
		return nil, err
	}
	if params.OrderID == "" {
		return nil, errMissingOrderID
	}
	raw, err := c.Call(ctx, "private/get_user_trades_by_order", params)
	if err != nil {
		return nil, err
	}
	var out []UserTrade
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func unmarshalIntish(raw json.RawMessage) (int, error) {
	var n int
	if err := json.Unmarshal(raw, &n); err == nil {
		return n, nil
	}
	var f float64
	if err := json.Unmarshal(raw, &f); err != nil {
		return 0, err
	}
	return int(f), nil
}
