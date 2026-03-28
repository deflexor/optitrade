package execution

import (
	"strings"
	"sync"
)

// ExposureBook tracks signed quantity by instrument after fills (MVP in-memory view for T049).
type ExposureBook struct {
	mu  sync.Mutex
	qty map[string]float64 // instrument_name -> net (buy positive, sell negative in coin units as reported by trades)
}

// NewExposureBook returns an empty book.
func NewExposureBook() *ExposureBook {
	return &ExposureBook{qty: make(map[string]float64)}
}

// ApplyUserTrade updates nets using Deribit user_trade direction and amount.
func (e *ExposureBook) ApplyUserTrade(instrument, direction string, amount float64) {
	if e == nil || instrument == "" || amount == 0 {
		return
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.qty == nil {
		e.qty = make(map[string]float64)
	}
	sign := 1.0
	if strings.EqualFold(strings.TrimSpace(direction), "sell") {
		sign = -1
	}
	e.qty[instrument] += sign * amount
}

// Net returns the signed amount for an instrument (0 if unknown).
func (e *ExposureBook) Net(instrument string) float64 {
	if e == nil {
		return 0
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.qty[instrument]
}
