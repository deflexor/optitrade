package deribit

// PlaceOrderParams are shared fields for private/buy and private/sell (options-aware).
type PlaceOrderParams struct {
	InstrumentName string   `json:"instrument_name"`
	Amount         *float64 `json:"amount,omitempty"`
	Contracts      *float64 `json:"contracts,omitempty"`
	Type           *string  `json:"type,omitempty"`
	Label          *string  `json:"label,omitempty"`
	Price          *float64 `json:"price,omitempty"`
	TimeInForce    *string  `json:"time_in_force,omitempty"`
	PostOnly       *bool    `json:"post_only,omitempty"`
	RejectPostOnly *bool    `json:"reject_post_only,omitempty"`
	ReduceOnly     *bool    `json:"reduce_only,omitempty"`
	// Advanced is "usd" or "implv" for options pricing semantics (see Deribit private/buy).
	Advanced *string `json:"advanced,omitempty"`
}

// PlacedOrderResponse is the result object from private/buy and private/sell.
type PlacedOrderResponse struct {
	Order  OrderDetail `json:"order"`
	Trades []UserTrade `json:"trades"`
}

// OrderDetail is the order object from placement, get_order_state, etc.
type OrderDetail struct {
	OrderID        string   `json:"order_id"`
	InstrumentName string   `json:"instrument_name"`
	OrderState     *string  `json:"order_state,omitempty"`
	OrderType      *string  `json:"order_type,omitempty"`
	Direction      *string  `json:"direction,omitempty"`
	Label          *string  `json:"label,omitempty"`
	Price          *float64 `json:"price,omitempty"`
	Amount         *float64 `json:"amount,omitempty"`
	FilledAmount   *float64 `json:"filled_amount,omitempty"`
	PostOnly       *bool    `json:"post_only,omitempty"`
	ReduceOnly     *bool    `json:"reduce_only,omitempty"`
	Advanced       *string  `json:"advanced,omitempty"`
}

// UserTrade is one element from private/get_user_trades_by_order (and placement trades[]).
type UserTrade struct {
	TradeID          string   `json:"trade_id"`
	OrderID          string   `json:"order_id"`
	InstrumentName   string   `json:"instrument_name"`
	Timestamp        int64    `json:"timestamp"`
	Direction        *string  `json:"direction,omitempty"`
	Amount           *float64 `json:"amount,omitempty"`
	Price            *float64 `json:"price,omitempty"`
	Fee              *float64 `json:"fee,omitempty"`
	FeeCurrency      *string  `json:"fee_currency,omitempty"`
	State            *string  `json:"state,omitempty"`
	ReduceOnly       *bool    `json:"reduce_only,omitempty"`
	PostOnly         *bool    `json:"post_only,omitempty"`
}

// GetUserTradesByOrderParams for private/get_user_trades_by_order.
type GetUserTradesByOrderParams struct {
	OrderID    string  `json:"order_id"`
	Sorting    *string `json:"sorting,omitempty"`
	Historical *bool   `json:"historical,omitempty"`
}

// CancelAllByInstrumentParams for private/cancel_all_by_instrument.
type CancelAllByInstrumentParams struct {
	InstrumentName string  `json:"instrument_name"`
	Type           *string `json:"type,omitempty"`
}

// CancelByLabelParams for private/cancel_by_label.
type CancelByLabelParams struct {
	Label    string  `json:"label"`
	Currency *string `json:"currency,omitempty"`
}
