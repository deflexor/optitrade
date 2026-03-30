package dashboard

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/dfr/optitrade/src/internal/deribit"
)

// exchangeReader is the Deribit surface the dashboard BFF uses (test doubles in unit tests).
type exchangeReader interface {
	GetAccountSummaries(ctx context.Context, p *deribit.GetAccountSummariesParams) ([]deribit.AccountSummary, error)
	GetPositions(ctx context.Context, p *deribit.GetPositionsParams) ([]deribit.Position, error)
	GetUserTrades(ctx context.Context, p deribit.GetUserTradesParams) ([]deribit.UserTrade, error)
	GetServerTime(ctx context.Context) (int64, error)
}

// exchangeWriter is optional reduce-only placement for close/rebalance confirms.
type exchangeWriter interface {
	Buy(ctx context.Context, params deribit.PlaceOrderParams) (*deribit.PlacedOrderResponse, error)
	Sell(ctx context.Context, params deribit.PlaceOrderParams) (*deribit.PlacedOrderResponse, error)
}

func rpcTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, 20*time.Second)
}

// Exchange is the Deribit facade consumed by dashboard handlers (nil when keys are absent).
type Exchange = exchangeReader

// ExchangeFromEnv builds a REST client when DERIBIT_CLIENT_ID and DERIBIT_CLIENT_SECRET are set.
func ExchangeFromEnv() Exchange {
	base := strings.TrimSpace(os.Getenv("DERIBIT_BASE_URL"))
	if base == "" {
		base = deribit.TestnetRPCBaseURL
	}
	id, sec := strings.TrimSpace(os.Getenv("DERIBIT_CLIENT_ID")), strings.TrimSpace(os.Getenv("DERIBIT_CLIENT_SECRET"))
	if id == "" || sec == "" {
		return nil
	}
	r, err := deribit.NewREST(base, &deribit.Credentials{ClientID: id, ClientSecret: sec})
	if err != nil {
		return nil
	}
	return r
}
