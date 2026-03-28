package deribit

import (
	"context"
	"encoding/json"

	"github.com/dfr/optitrade/execution/internal/deribit/rpc"
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
