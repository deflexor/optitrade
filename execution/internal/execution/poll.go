package execution

import (
	"context"
	"fmt"

	"github.com/dfr/optitrade/execution/internal/deribit"
	"github.com/dfr/optitrade/execution/internal/state"
)

// PollUserTradesForOrders calls private/get_user_trades_by_order for every non-terminal order that already has an exchange id (T049 polling path).
func PollUserTradesForOrders(ctx context.Context, api PrivateREST, orders state.OrderRepository, ingest *FillIngestor, states []string) error {
	if api == nil || orders == nil || ingest == nil {
		return fmt.Errorf("execution: poll requires api, orders, and ingest")
	}
	rows, err := orders.ListOrdersByStates(ctx, states)
	if err != nil {
		return err
	}
	for _, o := range rows {
		if o.ExchangeOrderID == nil || *o.ExchangeOrderID == "" {
			continue
		}
		trades, err := api.GetUserTradesByOrder(ctx, deribit.GetUserTradesByOrderParams{OrderID: *o.ExchangeOrderID})
		if err != nil {
			return err
		}
		for _, t := range trades {
			if err := ingest.IngestUserTrade(ctx, o.InternalOrderID, t); err != nil {
				return err
			}
		}
	}
	return nil
}
