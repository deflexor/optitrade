package deribit

import (
	"encoding/json"
)

// GetInstrumentsParams for public/get_instruments.
type GetInstrumentsParams struct {
	Currency *string `json:"currency,omitempty"`
	Kind     *string `json:"kind,omitempty"`
	Expired  *bool   `json:"expired,omitempty"`
}

// Instrument is a row from public/get_instruments.
type Instrument struct {
	InstrumentName      string   `json:"instrument_name"`
	InstrumentType      *string  `json:"instrument_type,omitempty"`
	Kind                *string  `json:"kind,omitempty"`
	OptionType          *string  `json:"option_type,omitempty"`
	TickSize            *float64 `json:"tick_size,omitempty"`
	ContractSize        *float64 `json:"contract_size,omitempty"`
	SettlementPeriod    *string  `json:"settlement_period,omitempty"`
	BaseCurrency        *string  `json:"base_currency,omitempty"`
	QuoteCurrency       *string  `json:"quote_currency,omitempty"`
	CreationTimestamp   *int64   `json:"creation_timestamp,omitempty"`
	ExpirationTimestamp *int64   `json:"expiration_timestamp,omitempty"`
	IsActive            *bool    `json:"is_active,omitempty"`
}

// GetOrderBookParams for public/get_order_book.
type GetOrderBookParams struct {
	InstrumentName string `json:"instrument_name"`
	Depth          *int   `json:"depth,omitempty"`
}

// PriceLevel is one bid/ask level (Deribit uses [price, amount] arrays).
type PriceLevel struct {
	Price  float64
	Amount float64
}

func (p *PriceLevel) UnmarshalJSON(data []byte) error {
	var v []float64
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	if len(v) < 2 {
		return errShortTuple
	}
	p.Price = v[0]
	p.Amount = v[1]
	return nil
}

// OrderBook is public/get_order_book result (partial fields).
type OrderBook struct {
	InstrumentName string       `json:"instrument_name"`
	Bids           []PriceLevel `json:"bids"`
	Asks           []PriceLevel `json:"asks"`
	Timestamp      *int64       `json:"timestamp,omitempty"`
	LastPrice      *float64     `json:"last_price,omitempty"`
	MarkPrice      *float64     `json:"mark_price,omitempty"`
	IndexPrice     *float64     `json:"index_price,omitempty"`
}

// TickerParams for public/ticker (instrument_name).
type TickerParams struct {
	InstrumentName string `json:"instrument_name"`
}

// Ticker is public/ticker result (subset used by options playbook).
type Ticker struct {
	InstrumentName  string   `json:"instrument_name"`
	LastPrice       *float64 `json:"last_price,omitempty"`
	BestBidPrice    *float64 `json:"best_bid_price,omitempty"`
	BestAskPrice    *float64 `json:"best_ask_price,omitempty"`
	MarkPrice       *float64 `json:"mark_price,omitempty"`
	IndexPrice      *float64 `json:"index_price,omitempty"`
	UnderlyingIndex *string  `json:"underlying_index,omitempty"`
	OpenInterest    *float64 `json:"open_interest,omitempty"`
	State           *string  `json:"state,omitempty"`
	Timestamp       *int64   `json:"timestamp,omitempty"`
}
