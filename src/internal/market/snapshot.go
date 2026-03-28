package market

import (
	"time"

	"github.com/dfr/optitrade/src/internal/deribit"
)

// QualityFlag describes book or feed quality issues.
type QualityFlag uint8

const (
	FlagNone QualityFlag = 0
	// StaleBook indicates the order book was not refreshed within the expected interval.
	StaleBook QualityFlag = 1 << iota
	// WideSpread indicates bid-ask spread exceeds a configured threshold (absolute or relative).
	WideSpread
	// Gap indicates a discontinuity in sequence or time (e.g. missed WS update).
	Gap
)

// MarketSnapshot is a unified view for strategy/risk modules.
type MarketSnapshot struct {
	Timestamp   time.Time
	Instrument  string
	Book        deribit.OrderBook
	BookLocalTS time.Time
	VolIndex    float64
	VolIndexTS  int64
	Flags       QualityFlag
}

// HasFlag reports whether f is set in the bitmask.
func (s MarketSnapshot) HasFlag(f QualityFlag) bool {
	return s.Flags&f != 0
}

// BuildSnapshotFromBook applies thresholds to mark StaleBook and WideSpread.
func BuildSnapshotFromBook(
	instrument string,
	book deribit.OrderBook,
	bookLocalTS time.Time,
	now time.Time,
	maxBookAge time.Duration,
	maxSpreadAbs float64,
	maxSpreadRel float64,
	vol float64,
	volTS int64,
) MarketSnapshot {
	s := MarketSnapshot{
		Timestamp:   now,
		Instrument:  instrument,
		Book:        book,
		BookLocalTS: bookLocalTS,
		VolIndex:    vol,
		VolIndexTS:  volTS,
	}
	if maxBookAge > 0 && !bookLocalTS.IsZero() && now.Sub(bookLocalTS) > maxBookAge {
		s.Flags |= StaleBook
	}
	if bid, ask, ok := bestBidAskFromBook(book); ok {
		spread := ask - bid
		if maxSpreadAbs > 0 && spread > maxSpreadAbs {
			s.Flags |= WideSpread
		}
		if maxSpreadRel > 0 && bid > 0 {
			if spread/bid > maxSpreadRel {
				s.Flags |= WideSpread
			}
		}
	}
	return s
}

func bestBidAskFromBook(book deribit.OrderBook) (bid, ask float64, ok bool) {
	if len(book.Bids) == 0 || len(book.Asks) == 0 {
		return 0, 0, false
	}
	return book.Bids[0].Price, book.Asks[0].Price, true
}
