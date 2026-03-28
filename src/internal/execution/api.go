package execution

import (
	"context"

	"github.com/dfr/optitrade/src/internal/deribit"
)

// PrivateREST is the subset of Deribit trading endpoints used by the execution layer.
type PrivateREST interface {
	Buy(ctx context.Context, params deribit.PlaceOrderParams) (*deribit.PlacedOrderResponse, error)
	Sell(ctx context.Context, params deribit.PlaceOrderParams) (*deribit.PlacedOrderResponse, error)
	CancelAllByInstrument(ctx context.Context, params deribit.CancelAllByInstrumentParams) (int, error)
	CancelByLabel(ctx context.Context, params deribit.CancelByLabelParams) (int, error)
	GetOpenOrders(ctx context.Context, params *deribit.GetOpenOrdersParams) ([]deribit.OpenOrder, error)
	GetOrderState(ctx context.Context, orderID string) (*deribit.OrderDetail, error)
	GetUserTradesByOrder(ctx context.Context, params deribit.GetUserTradesByOrderParams) ([]deribit.UserTrade, error)
}

// Verify *deribit.REST satisfies PrivateREST at compile time.
var _ PrivateREST = (*deribit.REST)(nil)
