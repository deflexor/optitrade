package execution

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/dfr/optitrade/execution/internal/deribit"
	"github.com/dfr/optitrade/execution/internal/session"
	"github.com/dfr/optitrade/execution/internal/state"
	"github.com/dfr/optitrade/execution/internal/state/sqlite"
)

type stubREST struct {
	buyRes   *deribit.PlacedOrderResponse
	buyErr   error
	sellRes  *deribit.PlacedOrderResponse
	open     []deribit.OpenOrder
	openErr  error
	state    *deribit.OrderDetail
	stateErr error
	trades   []deribit.UserTrade
	tradeErr error
	cancelN  int
}

func (s *stubREST) Buy(ctx context.Context, params deribit.PlaceOrderParams) (*deribit.PlacedOrderResponse, error) {
	if s.buyErr != nil {
		return nil, s.buyErr
	}
	return s.buyRes, nil
}

func (s *stubREST) Sell(ctx context.Context, params deribit.PlaceOrderParams) (*deribit.PlacedOrderResponse, error) {
	return s.sellRes, nil
}

func (s *stubREST) CancelAllByInstrument(ctx context.Context, params deribit.CancelAllByInstrumentParams) (int, error) {
	return s.cancelN, nil
}

func (s *stubREST) CancelByLabel(ctx context.Context, params deribit.CancelByLabelParams) (int, error) {
	return s.cancelN, nil
}

func (s *stubREST) GetOpenOrders(ctx context.Context, params *deribit.GetOpenOrdersParams) ([]deribit.OpenOrder, error) {
	if s.openErr != nil {
		return nil, s.openErr
	}
	return s.open, nil
}

func (s *stubREST) GetOrderState(ctx context.Context, orderID string) (*deribit.OrderDetail, error) {
	if s.stateErr != nil {
		return nil, s.stateErr
	}
	return s.state, nil
}

func (s *stubREST) GetUserTradesByOrder(ctx context.Context, params deribit.GetUserTradesByOrderParams) ([]deribit.UserTrade, error) {
	if s.tradeErr != nil {
		return nil, s.tradeErr
	}
	return s.trades, nil
}

func TestPlacerDryRunUpdatesRow(t *testing.T) {
	t.Parallel()
	db, err := sqlite.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := sqlite.NewStore(db)
	ctx := context.Background()
	rec := &state.OrderRecord{
		InternalOrderID: "cl-1",
		InstrumentName:  "BTC-PERPETUAL",
		Label:           "L",
		Side:            "buy",
		OrderType:       "limit",
		Amount:          "1",
		PostOnly:        true,
		ReduceOnly:      false,
		State:           state.OrderStateNew,
		CreatedAt:       1,
		UpdatedAt:       1,
	}
	if err := store.InsertOrder(ctx, rec); err != nil {
		t.Fatal(err)
	}
	p := NewPlacer(&stubREST{}, store)
	p.DryRun = true
	price := 100.0
	if _, err := p.PlaceLimit(ctx, rec, 1, &price, nil, "corr-12345678"); err != nil {
		t.Fatal(err)
	}
	got, err := store.GetOrder(ctx, "cl-1")
	if err != nil {
		t.Fatal(err)
	}
	if got.State != state.OrderStateOpen || got.ExchangeOrderID == nil || *got.ExchangeOrderID != "dry-run" {
		t.Fatalf("got %+v", got)
	}
}

func TestPlacerBlockedWhenSessionProtectiveNonReduceOnly(t *testing.T) {
	t.Parallel()
	db, err := sqlite.Open(filepath.Join(t.TempDir(), "guard.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := sqlite.NewStore(db)
	ctx := context.Background()
	rec := &state.OrderRecord{
		InternalOrderID: "cl-g",
		InstrumentName:  "BTC-PERPETUAL",
		Label:           "L",
		Side:            "buy",
		OrderType:       "limit",
		Amount:          "1",
		PostOnly:        true,
		ReduceOnly:      false,
		State:           state.OrderStateNew,
		CreatedAt:       1,
		UpdatedAt:       1,
	}
	if err := store.InsertOrder(ctx, rec); err != nil {
		t.Fatal(err)
	}
	sess := session.NewFSM()
	sess.NotifyRPCAuthFailure()
	p := NewPlacer(&stubREST{
		buyRes: &deribit.PlacedOrderResponse{
			Order: deribit.OrderDetail{OrderID: "should-not-run"},
		},
	}, store)
	p.Session = sess
	price := 1.0
	if _, err := p.PlaceLimit(ctx, rec, 1, &price, nil, "corr-12345678"); err == nil {
		t.Fatal("expected session guard error")
	}
}

func TestPlacerIdempotentWhenExchangeIDSet(t *testing.T) {
	t.Parallel()
	db, err := sqlite.Open(filepath.Join(t.TempDir(), "t2.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := sqlite.NewStore(db)
	ctx := context.Background()
	ex := "X-99"
	rec := &state.OrderRecord{
		InternalOrderID: "cl-2",
		ExchangeOrderID: &ex,
		InstrumentName:  "BTC-PERPETUAL",
		Label:           "L",
		Side:            "buy",
		OrderType:       "limit",
		Amount:          "1",
		PostOnly:        true,
		State:           state.OrderStateOpen,
		CreatedAt:       1,
		UpdatedAt:       1,
	}
	if err := store.InsertOrder(ctx, rec); err != nil {
		t.Fatal(err)
	}
	api := &stubREST{
		buyRes: &deribit.PlacedOrderResponse{
			Order: deribit.OrderDetail{OrderID: "should-not-run"},
		},
	}
	p := NewPlacer(api, store)
	price := 1.0
	if _, err := p.PlaceLimit(ctx, rec, 1, &price, nil, "corr-12345678"); err != nil {
		t.Fatal(err)
	}
	got, _ := store.GetOrder(ctx, "cl-2")
	if *got.ExchangeOrderID != "X-99" {
		t.Fatalf("exchange id changed: %+v", got.ExchangeOrderID)
	}
}

func TestReconcileInsertsMissingLocal(t *testing.T) {
	t.Parallel()
	db, err := sqlite.Open(filepath.Join(t.TempDir(), "r.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := sqlite.NewStore(db)
	ctx := context.Background()
	api := &stubREST{
		open: []deribit.OpenOrder{{
			OrderID: "ETH-1", InstrumentName: "ETH-PERPETUAL",
			Label: strPtr("lb"), Direction: strPtr("buy"),
		}},
	}
	r := Reconciler{API: api, Orders: store}
	if err := r.Run(ctx, nil); err != nil {
		t.Fatal(err)
	}
	got, err := store.GetOrder(ctx, SyncOrderIDPrefix+"ETH-1")
	if err != nil {
		t.Fatal(err)
	}
	if got.InstrumentName != "ETH-PERPETUAL" {
		t.Fatalf("%+v", got)
	}
}

func TestReconcileOrphanMarksCancelledOnStateError(t *testing.T) {
	t.Parallel()
	db, err := sqlite.Open(filepath.Join(t.TempDir(), "o.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := sqlite.NewStore(db)
	ctx := context.Background()
	ex := "ETH-2"
	loc := &state.OrderRecord{
		InternalOrderID: "local-1",
		ExchangeOrderID: &ex,
		InstrumentName:  "ETH-PERPETUAL",
		Label:           "x",
		Side:            "buy",
		OrderType:       "limit",
		Amount:          "1",
		PostOnly:        true,
		State:           state.OrderStateOpen,
		CreatedAt:       time.Now().UnixMilli(),
		UpdatedAt:       time.Now().UnixMilli(),
	}
	if err := store.InsertOrder(ctx, loc); err != nil {
		t.Fatal(err)
	}
	api := &stubREST{open: nil, stateErr: context.Canceled}
	r := Reconciler{API: api, Orders: store}
	if err := r.Run(ctx, nil); err != nil {
		t.Fatal(err)
	}
	got, err := store.GetOrder(ctx, "local-1")
	if err != nil {
		t.Fatal(err)
	}
	if got.State != state.OrderStateCanceled {
		t.Fatalf("got %q", got.State)
	}
}

func TestPollUserTradesIngests(t *testing.T) {
	t.Parallel()
	db, err := sqlite.Open(filepath.Join(t.TempDir(), "p.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := sqlite.NewStore(db)
	ctx := context.Background()
	ex := "ETH-3"
	if err := store.InsertOrder(ctx, &state.OrderRecord{
		InternalOrderID: "io",
		ExchangeOrderID: &ex,
		InstrumentName:  "ETH-PERPETUAL",
		Label:           "z",
		Side:            "buy",
		OrderType:       "limit",
		Amount:          "1",
		PostOnly:        true,
		State:           state.OrderStateOpen,
		CreatedAt:       1,
		UpdatedAt:       1,
	}); err != nil {
		t.Fatal(err)
	}
	api := &stubREST{
		trades: []deribit.UserTrade{{
			TradeID: "TR-1", OrderID: ex, InstrumentName: "ETH-PERPETUAL",
			Timestamp: 10,
			Amount:    f64p(2), Price: f64p(3), Fee: f64p(0.0001), Direction: strPtr("buy"),
		}},
	}
	exp := NewExposureBook()
	ing := &FillIngestor{Fills: store, Exposure: exp}
	if err := PollUserTradesForOrders(ctx, api, store, ing, []string{state.OrderStateOpen}); err != nil {
		t.Fatal(err)
	}
	if exp.Net("ETH-PERPETUAL") != 2 {
		t.Fatalf("exposure %v", exp.Net("ETH-PERPETUAL"))
	}
	inserted, err := store.InsertFill(ctx, &state.FillRecord{
		ID: "x", OrderID: "io", TradeID: "TR-1", InstrumentName: "ETH-PERPETUAL",
		Qty: "1", Price: "1", Fee: "0", FilledAt: 1,
	})
	if err != nil || inserted {
		// duplicate trade_id should no-op
		t.Fatalf("duplicate? inserted=%v err=%v", inserted, err)
	}
}

func strPtr(s string) *string { return &s }
func f64p(f float64) *float64 { return &f }

// SC-004: after reconcile, no local row stays "open" when the exchange reports the order filled.
func TestSC004_ReconcileMarksFilledWithoutOrphanOpen(t *testing.T) {
	t.Parallel()
	db, err := sqlite.Open(filepath.Join(t.TempDir(), "sc4.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := sqlite.NewStore(db)
	ctx := context.Background()
	ex := "EX-7"
	filled := "filled"
	if err := store.InsertOrder(ctx, &state.OrderRecord{
		InternalOrderID: "loc-sc4",
		ExchangeOrderID: &ex,
		InstrumentName:  "ETH-PERPETUAL",
		Label:           "x",
		Side:            "buy",
		OrderType:       "limit",
		Amount:          "1",
		PostOnly:        true,
		State:           state.OrderStateOpen,
		CreatedAt:       1,
		UpdatedAt:       1,
	}); err != nil {
		t.Fatal(err)
	}
	api := &stubREST{
		open: nil,
		state: &deribit.OrderDetail{
			OrderState: &filled,
		},
	}
	r := Reconciler{API: api, Orders: store}
	if err := r.Run(ctx, nil); err != nil {
		t.Fatal(err)
	}
	got, err := store.GetOrder(ctx, "loc-sc4")
	if err != nil {
		t.Fatal(err)
	}
	if got.State != state.OrderStateFilled {
		t.Fatalf("want filled, got %q", got.State)
	}
	left, err := store.ListOrdersByStates(ctx, []string{state.OrderStateOpen, state.OrderStatePartiallyFilled})
	if err != nil {
		t.Fatal(err)
	}
	for _, o := range left {
		if o.InternalOrderID == "loc-sc4" {
			t.Fatalf("orphan open row after reconcile: %+v", o)
		}
	}
}
