package deribit

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestIntegrationTestnetREST(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testnet integration in -short")
	}
	id := os.Getenv("DERIBIT_CLIENT_ID")
	sec := os.Getenv("DERIBIT_CLIENT_SECRET")
	if id == "" || sec == "" {
		t.Skip("set DERIBIT_CLIENT_ID and DERIBIT_CLIENT_SECRET for live testnet test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pub, err := NewREST(TestnetRPCBaseURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	ms, err := pub.GetServerTime(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if ms <= 0 {
		t.Fatalf("unexpected server time %d", ms)
	}

	creds := &Credentials{ClientID: id, ClientSecret: sec}
	priv, err := NewREST(TestnetRPCBaseURL, creds)
	if err != nil {
		t.Fatal(err)
	}
	pos, err := priv.GetPositions(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	_ = pos

	_, err = priv.GetOpenOrders(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}

	cur := "BTC"
	summaries, err := priv.GetAccountSummaries(ctx, &GetAccountSummariesParams{Currency: &cur})
	if err != nil {
		t.Fatal(err)
	}
	if len(summaries) == 0 {
		t.Fatal("expected at least one account summary for BTC")
	}
}
