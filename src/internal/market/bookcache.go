package market

import (
	"sync"
	"time"

	"github.com/dfr/optitrade/src/internal/deribit"
)

// OrderBookCache stores the latest order book up to depth levels (thread-safe).
type OrderBookCache struct {
	mu     sync.RWMutex
	depth  int
	book   deribit.OrderBook
	tsWall time.Time
}

// NewOrderBookCache returns a cache that retains at most depth bid/ask levels (per side). depth <= 0 means no trimming.
func NewOrderBookCache(depth int) *OrderBookCache {
	return &OrderBookCache{depth: depth}
}

// Update replaces the cached book and records local wall time (Deribit book timestamp is also stored on the struct).
func (c *OrderBookCache) Update(ob *deribit.OrderBook) {
	if c == nil || ob == nil {
		return
	}
	b := *ob
	if c.depth > 0 {
		b.Bids = trimLevels(b.Bids, c.depth)
		b.Asks = trimLevels(b.Asks, c.depth)
	}
	c.mu.Lock()
	c.book = b
	c.tsWall = time.Now()
	c.mu.Unlock()
}

func trimLevels(levels []deribit.PriceLevel, n int) []deribit.PriceLevel {
	if n <= 0 || len(levels) <= n {
		return levels
	}
	cp := make([]deribit.PriceLevel, n)
	copy(cp, levels[:n])
	return cp
}

// Book returns a shallow copy of the cached book and when it was last updated locally.
func (c *OrderBookCache) Book() (deribit.OrderBook, time.Time) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.book, c.tsWall
}

// BestBidAsk returns top-of-book prices if both sides exist.
func (c *OrderBookCache) BestBidAsk() (bid, ask float64, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.book.Bids) == 0 || len(c.book.Asks) == 0 {
		return 0, 0, false
	}
	return c.book.Bids[0].Price, c.book.Asks[0].Price, true
}

// Depth returns up to n levels per side from the cached book (0 = all stored levels).
func (c *OrderBookCache) Depth(n int) (bids, asks []deribit.PriceLevel) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	bids = c.book.Bids
	asks = c.book.Asks
	if n > 0 {
		bids = trimLevels(bids, n)
		asks = trimLevels(asks, n)
	} else {
		bids = append([]deribit.PriceLevel(nil), bids...)
		asks = append([]deribit.PriceLevel(nil), asks...)
	}
	return bids, asks
}
