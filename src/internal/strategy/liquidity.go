package strategy

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/dfr/optitrade/src/internal/config"
	"github.com/dfr/optitrade/src/internal/deribit"
)

// TouchLiquidity is best bid/ask price and displayed size at the touch.
type TouchLiquidity struct {
	BidPrice  float64
	AskPrice  float64
	BidSize   float64
	AskSize   float64
	Mid       float64
	SpreadBps float64
}

// TouchFromBook extracts top-of-book prices and sizes. ok is false if either
// side is missing.
func TouchFromBook(book deribit.OrderBook) (t TouchLiquidity, ok bool) {
	if len(book.Bids) == 0 || len(book.Asks) == 0 {
		return TouchLiquidity{}, false
	}
	b := book.Bids[0]
	a := book.Asks[0]
	mid := (b.Price + a.Price) / 2
	if mid <= 0 || math.IsNaN(mid) || math.IsInf(mid, 0) {
		return TouchLiquidity{}, false
	}
	spread := a.Price - b.Price
	spreadBps := (spread / mid) * 10_000
	return TouchLiquidity{
		BidPrice:  b.Price,
		AskPrice:  a.Price,
		BidSize:   b.Amount,
		AskSize:   a.Amount,
		Mid:       mid,
		SpreadBps: spreadBps,
	}, true
}

// LiquidityOk returns true when touch spread and sizes satisfy policy
// liquidity.max_spread_bps and liquidity.min_top_size. When false, reason
// briefly explains the first failing check (for logs).
func LiquidityOk(book deribit.OrderBook, liq config.Liquidity) (ok bool, reason string) {
	touch, okBook := TouchFromBook(book)
	if !okBook {
		return false, "book missing bid or ask"
	}
	if touch.SpreadBps > float64(liq.MaxSpreadBps) {
		return false, fmt.Sprintf("spread %.2f bps exceeds max %d bps", touch.SpreadBps, liq.MaxSpreadBps)
	}
	minTop, err := parsePositiveSize(liq.MinTopSize)
	if err != nil {
		return false, fmt.Sprintf("policy min_top_size: %v", err)
	}
	minSide := touch.BidSize
	if touch.AskSize < minSide {
		minSide = touch.AskSize
	}
	if minSide < minTop {
		return false, fmt.Sprintf("top size %.6g below min_top_size %.6g", minSide, minTop)
	}
	return true, ""
}

func parsePositiveSize(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty size")
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	if math.IsNaN(f) || math.IsInf(f, 0) || f <= 0 {
		return 0, fmt.Errorf("invalid size %q", s)
	}
	return f, nil
}
