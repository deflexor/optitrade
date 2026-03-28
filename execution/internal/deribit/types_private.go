package deribit

// GetPositionsParams are optional filters for private/get_positions.
type GetPositionsParams struct {
	Currency *string `json:"currency,omitempty"`
	Kind     *string `json:"kind,omitempty"`
}

// Position mirrors common fields from private/get_positions entries (nullable per API).
type Position struct {
	InstrumentName       string   `json:"instrument_name"`
	Size                 float64  `json:"size"`
	Direction            *string  `json:"direction,omitempty"`
	AveragePrice         *float64 `json:"average_price,omitempty"`
	FloatingProfitLoss   *float64 `json:"floating_profit_loss,omitempty"`
	RealizedProfitLoss   *float64 `json:"realized_profit_loss,omitempty"`
	InitialMargin        *float64 `json:"initial_margin,omitempty"`
	MaintenanceMargin    *float64 `json:"maintenance_margin,omitempty"`
	Delta                *float64 `json:"delta,omitempty"`
	Gamma                *float64 `json:"gamma,omitempty"`
	Vega                 *float64 `json:"vega,omitempty"`
	Theta                *float64 `json:"theta,omitempty"`
	MarkPrice            *float64 `json:"mark_price,omitempty"`
	IndexPrice           *float64 `json:"index_price,omitempty"`
	AveragePriceUSD      *float64 `json:"average_price_usd,omitempty"`
}

// GetOpenOrdersParams for private/get_open_orders.
type GetOpenOrdersParams struct {
	Currency *string `json:"currency,omitempty"`
	Kind     *string `json:"kind,omitempty"`
	Type     *string `json:"type,omitempty"`
}

// OpenOrder is a subset of private/get_open_orders rows.
type OpenOrder struct {
	OrderID            string   `json:"order_id"`
	InstrumentName     string   `json:"instrument_name"`
	Direction          *string  `json:"direction,omitempty"`
	Price              *float64 `json:"price,omitempty"`
	Amount             *float64 `json:"amount,omitempty"`
	FilledAmount       *float64 `json:"filled_amount,omitempty"`
	AveragePrice       *float64 `json:"average_price,omitempty"`
	OrderState         *string  `json:"order_state,omitempty"`
	OrderType          *string  `json:"order_type,omitempty"`
	TimeInForce        *string  `json:"time_in_force,omitempty"`
	CreationTimestamp  *int64   `json:"creation_timestamp,omitempty"`
	LastUpdateTimestamp *int64  `json:"last_update_timestamp,omitempty"`
}

// GetAccountSummariesParams for private/get_account_summaries.
type GetAccountSummariesParams struct {
	Currency *string `json:"currency,omitempty"`
	Extended *bool   `json:"extended,omitempty"`
}

// AccountSummary is one element of private/get_account_summaries.
type AccountSummary struct {
	Currency              string   `json:"currency"`
	Balance               *float64 `json:"balance,omitempty"`
	Equity                *float64 `json:"equity,omitempty"`
	PortfolioMargining  *bool    `json:"portfolio_margining,omitempty"`
	DeltaTotal            *float64 `json:"delta_total,omitempty"`
	SessionRPL            *float64 `json:"session_rpl,omitempty"`
	SessionUPL            *float64 `json:"session_upl,omitempty"`
}
