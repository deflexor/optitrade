package strategy

import (
	"testing"

	"github.com/dfr/optitrade/execution/internal/config"
	"github.com/dfr/optitrade/execution/internal/deribit"
	"github.com/dfr/optitrade/execution/internal/regime"
)

func TestLiquidityOk_wideSpread_rejected(t *testing.T) {
	book := deribit.OrderBook{
		Bids: []deribit.PriceLevel{{Price: 100, Amount: 5}},
		Asks: []deribit.PriceLevel{{Price: 102.5, Amount: 5}},
	}
	liq := config.Liquidity{MinTopSize: "0.1", MaxSpreadBps: 100}
	ok, reason := LiquidityOk(book, liq)
	if ok {
		t.Fatalf("expected illiquid book to fail; mid spread is wide in bps")
	}
	if reason == "" {
		t.Fatalf("expected reason")
	}
	t.Log(reason)
}

func TestLiquidityOk_tightBook_accepted(t *testing.T) {
	book := deribit.OrderBook{
		Bids: []deribit.PriceLevel{{Price: 100, Amount: 2}},
		Asks: []deribit.PriceLevel{{Price: 100.02, Amount: 2}},
	}
	liq := config.Liquidity{MinTopSize: "1.0", MaxSpreadBps: 50}
	ok, reason := LiquidityOk(book, liq)
	if !ok {
		t.Fatalf("expected liquid: %s", reason)
	}
}

func TestLiquidityOk_smallTopSize_rejected(t *testing.T) {
	book := deribit.OrderBook{
		Bids: []deribit.PriceLevel{{Price: 100, Amount: 0.05}},
		Asks: []deribit.PriceLevel{{Price: 100.01, Amount: 5}},
	}
	liq := config.Liquidity{MinTopSize: "1.0", MaxSpreadBps: 500}
	ok, _ := LiquidityOk(book, liq)
	if ok {
		t.Fatal("expected size gate to fail")
	}
}

func TestAllowedStructures(t *testing.T) {
	policyJSON := []byte(`{
  "version": "1.0.0",
  "limits": {
    "max_loss_per_trade": "1","max_daily_loss":"1","max_open_premium_at_risk":"1",
    "max_portfolio_delta":"0","max_portfolio_vega":"0",
    "max_open_orders_per_instrument": 1,"max_time_in_trade_seconds": 1
  },
  "liquidity": { "min_top_size": "1", "max_spread_bps": 25 },
  "playbooks": {
    "low": { "allowed_structures": ["iron_condor"] },
    "normal": { "allowed_structures": ["credit_spread","debit_spread"] },
    "high": { "allowed_structures": ["credit_spread"] }
  }
}`)
	pol, err := config.LoadBytes(policyJSON)
	if err != nil {
		t.Fatal(err)
	}
	got, err := AllowedStructures(pol, regime.LabelHigh)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "credit_spread" {
		t.Fatalf("got %+v", got)
	}
}

func TestTemplates_namesAndValidate(t *testing.T) {
	legs, err := VerticalPutCredit("btc", "28MAR25", 90000, 1000)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateDefinedRisk(legs); err != nil {
		t.Fatal(err)
	}
	if got := legs[0].Instrument; got != "BTC-28MAR25-90000-P" {
		t.Fatalf("got %q", got)
	}

	legs2, err := VerticalCallDebit("ETH", "27JUN25", 3200, 50)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateDefinedRisk(legs2); err != nil {
		t.Fatal(err)
	}

	ic, err := IronCondor("BTC", "28MAR25", 80000, 82000, 98000, 100000)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateDefinedRisk(ic); err != nil {
		t.Fatal(err)
	}
}

func TestValidateDefinedRisk_nakedVertical_rejected(t *testing.T) {
	legs := []LegSpec{
		{Instrument: "BTC-28MAR25-90000-P", Side: LegSell},
		{Instrument: "BTC-28MAR25-89000-C", Side: LegBuy},
	}
	if err := ValidateDefinedRisk(legs); err == nil {
		t.Fatal("expected mixed put/call to fail vertical check")
	}
}

func TestDevInvariantPanic(t *testing.T) {
	t.Setenv("OPTITRADE_DEV_INVARIANTS", "1")

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when invariant fails in dev")
		}
	}()
	mustAssertDefinedRiskInDev([]LegSpec{
		{Instrument: "BTC-28MAR25-90000-P", Side: LegSell},
		{Instrument: "BTC-28MAR25-89000-P", Side: LegSell},
	})
}
