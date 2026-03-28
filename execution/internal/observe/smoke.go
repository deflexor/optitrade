package observe

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dfr/optitrade/execution/internal/config"
	"github.com/dfr/optitrade/execution/internal/deribit"
)

func allowTestnetOrders() bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(EnvAllowTestnetOrders)))
	return v == "1" || v == "true" || v == "yes"
}

// RunSmokeOrder places a deep out-of-market post-only limit buy, then cancels all orders
// on the instrument. It is blocked unless EnvAllowTestnetOrders is set and policy.environment is testnet.
func RunSmokeOrder(ctx context.Context, baseURL, clientID, secret, policyPath, instrument string, amount float64) error {
	if !allowTestnetOrders() {
		return fmt.Errorf("%s must be 1 to place testnet orders (safety gate)", EnvAllowTestnetOrders)
	}
	pol, err := config.LoadFile(policyPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(strings.ToLower(pol.Environment)) != "testnet" {
		return fmt.Errorf("smoke-order: policy.environment must be testnet, got %q", pol.Environment)
	}
	creds := &deribit.Credentials{ClientID: strings.TrimSpace(clientID), ClientSecret: strings.TrimSpace(secret)}
	r, err := deribit.NewREST(baseURL, creds)
	if err != nil {
		return err
	}
	inst := strings.TrimSpace(instrument)
	if inst == "" {
		return fmt.Errorf("smoke-order: instrument required")
	}
	if amount <= 0 {
		return fmt.Errorf("smoke-order: amount must be positive")
	}
	defer func() {
		cctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		_, _ = r.CancelAllByInstrument(cctx, deribit.CancelAllByInstrumentParams{InstrumentName: inst})
	}()

	book, err := r.GetOrderBook(ctx, deribit.GetOrderBookParams{InstrumentName: inst, Depth: intPtr(1)})
	if err != nil {
		return err
	}
	var bid float64
	if len(book.Bids) > 0 {
		bid = book.Bids[0].Price
	}
	if bid <= 0 {
		return fmt.Errorf("smoke-order: no bid for %s", inst)
	}
	// Far below market: unlikely to fill; still cancels via defer.
	price := bid * 0.5
	lim := "limit"
	po := true
	tif := "good_til_cancelled"
	params := deribit.PlaceOrderParams{
		InstrumentName: inst,
		Amount:         &amount,
		Type:           &lim,
		Price:          &price,
		PostOnly:       &po,
		TimeInForce:    &tif,
		Label:          strPtr("optitrade-smoke"),
	}
	resp, err := r.Buy(ctx, params)
	if err != nil {
		return err
	}
	if resp == nil || resp.Order.OrderID == "" {
		return fmt.Errorf("smoke-order: empty order response")
	}
	return nil
}

func strPtr(s string) *string { return &s }
