//go:build integration

package observe

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/dfr/optitrade/execution/internal/deribit"
)

// T060: live testnet read-only observation (skipped without credentials).
func TestT060_ReadOnlyObserveTestnet(t *testing.T) {
	if os.Getenv("DERIBIT_CLIENT_ID") == "" || os.Getenv("DERIBIT_CLIENT_SECRET") == "" {
		t.Skip("set DERIBIT_CLIENT_ID and DERIBIT_CLIENT_SECRET for integration")
	}
	base := os.Getenv("DERIBIT_BASE_URL")
	if base == "" {
		base = deribit.TestnetRPCBaseURL
	}
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	err := RunReadOnly(ctx, Config{
		BaseURL:    base,
		ClientID:   os.Getenv("DERIBIT_CLIENT_ID"),
		Secret:     os.Getenv("DERIBIT_CLIENT_SECRET"),
		Instrument: "BTC-PERPETUAL",
		Interval:   4 * time.Second,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
}
