package observe

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/dfr/optitrade/execution/internal/deribit"
)

// EnvAllowTestnetOrders must be "1" before RunSmokeOrder sends any private order (T061).
const EnvAllowTestnetOrders = "OPTITRADE_ALLOW_TESTNET_ORDERS"

// Config holds read-only observation settings (T060).
type Config struct {
	BaseURL    string
	ClientID   string
	Secret     string
	Instrument string
	Interval   time.Duration
}

// RunReadOnly prints positions and top-of-book snapshots until ctx is cancelled.
func RunReadOnly(ctx context.Context, cfg Config, log *slog.Logger) error {
	if log == nil {
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	creds := &deribit.Credentials{ClientID: strings.TrimSpace(cfg.ClientID), ClientSecret: strings.TrimSpace(cfg.Secret)}
	r, err := deribit.NewREST(cfg.BaseURL, creds)
	if err != nil {
		return err
	}
	inst := strings.TrimSpace(cfg.Instrument)
	if inst == "" {
		return fmt.Errorf("observe: instrument is required")
	}
	tick := cfg.Interval
	if tick <= 0 {
		tick = 3 * time.Second
	}
	poll := func() {
		if _, err := r.GetServerTime(ctx); err != nil {
			log.Error("observe: server time", "err", err)
		}
		pos, err := r.GetPositions(ctx, nil)
		if err != nil {
			log.Error("observe: get_positions", "err", err)
		} else {
			log.Info("observe: positions", "count", len(pos))
			for _, p := range pos {
				log.Info("position", "instrument", p.InstrumentName, "size", p.Size)
			}
		}
		book, err := r.GetOrderBook(ctx, deribit.GetOrderBookParams{InstrumentName: inst, Depth: intPtr(5)})
		if err != nil {
			log.Error("observe: order_book", "instrument", inst, "err", err)
		} else {
			bid, ask := bookTop(book)
			log.Info("observe: book", "instrument", inst, "best_bid", bid, "best_ask", ask)
		}
	}
	poll()
	t := time.NewTicker(tick)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Info("observe: done", "err", ctx.Err())
			return nil
		case <-t.C:
			poll()
		}
	}
}

func bookTop(book *deribit.OrderBook) (bid, ask float64) {
	if book == nil {
		return 0, 0
	}
	if len(book.Bids) > 0 {
		bid = book.Bids[0].Price
	}
	if len(book.Asks) > 0 {
		ask = book.Asks[0].Price
	}
	return bid, ask
}

func intPtr(n int) *int { return &n }
