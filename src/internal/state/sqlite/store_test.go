package sqlite

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/dfr/optitrade/src/internal/state"
)

func TestOpenAppliesMigrationsEmptyDB(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.db")
	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var n int
	if err := db.QueryRow(`SELECT COUNT(1) FROM schema_migrations`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("migrations recorded: got %d want 2", n)
	}

	if err := db.QueryRow(`SELECT COUNT(1) FROM order_record`).Scan(&n); err != nil {
		t.Fatal(err)
	}
}

func TestMigrationsIdempotent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "idem.db")
	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	_ = db.Close()

	db2, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db2.Close()

	var versions int
	if err := db2.QueryRow(`SELECT COUNT(1) FROM schema_migrations`).Scan(&versions); err != nil {
		t.Fatal(err)
	}
	if versions != 2 {
		t.Fatalf("want 2 migration rows, got %d", versions)
	}
}

func TestOrderInsertSelect(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	db, err := Open(filepath.Join(dir, "orders.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store := NewStore(db)
	ctx := context.Background()
	price := "0.05"
	o := &state.OrderRecord{
		InternalOrderID: "cl-001",
		InstrumentName:  "BTC-28MAR26-90000-P",
		Label:           "test",
		Side:            "buy",
		OrderType:       "limit",
		Price:           &price,
		Amount:          "0.1",
		PostOnly:        true,
		ReduceOnly:      false,
		State:           "open",
		CreatedAt:       time.Now().UnixMilli(),
		UpdatedAt:       time.Now().UnixMilli(),
	}
	if err := store.InsertOrder(ctx, o); err != nil {
		t.Fatal(err)
	}
	got, err := store.GetOrder(ctx, "cl-001")
	if err != nil {
		t.Fatal(err)
	}
	if got.InstrumentName != o.InstrumentName || got.State != o.State {
		t.Fatalf("got %+v", got)
	}
	if got.Price == nil || *got.Price != price {
		t.Fatalf("price: %+v", got.Price)
	}
	if !got.PostOnly || got.ReduceOnly {
		t.Fatalf("flags: post=%v reduce=%v", got.PostOnly, got.ReduceOnly)
	}
}

func TestAuditInsertSelect(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	db, err := Open(filepath.Join(dir, "audit.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store := NewStore(db)
	ctx := context.Background()
	a := &state.AuditDecision{
		ID:               "aud-1",
		Ts:               time.Now().UnixMilli(),
		DecisionType:     "veto_risk",
		RegimeLabel:      "high",
		CostModelVersion: "1",
		RiskGateResults:  `{"delta_cap":false}`,
		Reason:           "over limit",
		CorrelationID:    "corr-1",
	}
	if err := store.InsertAudit(ctx, a); err != nil {
		t.Fatal(err)
	}
	got, err := store.GetAudit(ctx, "aud-1")
	if err != nil {
		t.Fatal(err)
	}
	if got.DecisionType != a.DecisionType || got.Reason != a.Reason || got.RiskGateResults != a.RiskGateResults {
		t.Fatalf("got %+v", got)
	}
	if got.CandidateID != nil {
		t.Fatal("expected nil candidate")
	}
}

func TestOrderWithCandidateFK(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	db, err := Open(filepath.Join(dir, "fk.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()
	candID := "cand-uuid-1"
	_, err = db.ExecContext(ctx, `
INSERT INTO trade_candidate (id, created_at, regime_label, playbook_id, structure_type, legs_json, expected_edge, cost_breakdown_json)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		candID, time.Now().UnixMilli(), "normal", "p1", "vertical", `[]`, "0", `{}`,
	)
	if err != nil {
		t.Fatal(err)
	}

	store := NewStore(db)
	o := &state.OrderRecord{
		InternalOrderID: "cl-002",
		InstrumentName:  "ETH-PERPETUAL",
		Label:           "l",
		Side:            "sell",
		OrderType:       "limit",
		Amount:          "1",
		PostOnly:        false,
		ReduceOnly:      true,
		State:           "new",
		CreatedAt:       1,
		UpdatedAt:       1,
		CandidateID:     &candID,
	}
	if err := store.InsertOrder(ctx, o); err != nil {
		t.Fatal(err)
	}
	got, err := store.GetOrder(ctx, "cl-002")
	if err != nil {
		t.Fatal(err)
	}
	if got.CandidateID == nil || *got.CandidateID != candID {
		t.Fatalf("candidate: %+v", got.CandidateID)
	}
}

func TestOrderUpdateListFills(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	db, err := Open(filepath.Join(dir, "upd.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store := NewStore(db)
	ctx := context.Background()
	ts := time.Now().UnixMilli()
	o := &state.OrderRecord{
		InternalOrderID: "i1",
		InstrumentName:  "BTC-PERPETUAL",
		Label:           "L",
		Side:            "buy",
		OrderType:       "limit",
		Amount:          "1",
		PostOnly:        true,
		ReduceOnly:      false,
		State:           state.OrderStateNew,
		CreatedAt:       ts,
		UpdatedAt:       ts,
	}
	if err := store.InsertOrder(ctx, o); err != nil {
		t.Fatal(err)
	}
	ex := "X-1"
	o.ExchangeOrderID = &ex
	o.State = state.OrderStateOpen
	o.UpdatedAt = ts + 1
	if err := store.UpdateOrder(ctx, o); err != nil {
		t.Fatal(err)
	}
	got, err := store.GetOrder(ctx, "i1")
	if err != nil {
		t.Fatal(err)
	}
	if got.ExchangeOrderID == nil || *got.ExchangeOrderID != ex || got.State != state.OrderStateOpen {
		t.Fatalf("after update: %+v", got)
	}

	rows, err := store.ListOrdersByStates(ctx, []string{state.OrderStateOpen})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].InternalOrderID != "i1" {
		t.Fatalf("list: %+v", rows)
	}

	f := &state.FillRecord{
		ID:             "f1",
		OrderID:        "i1",
		TradeID:        "T-1",
		InstrumentName: "BTC-PERPETUAL",
		Qty:            "1",
		Price:          "100",
		Fee:            "0.0001",
		FilledAt:       ts + 2,
	}
	inserted, err := store.InsertFill(ctx, f)
	if err != nil || !inserted {
		t.Fatalf("first insert: inserted=%v err=%v", inserted, err)
	}
	inserted, err = store.InsertFill(ctx, f)
	if err != nil || inserted {
		t.Fatalf("second insert: inserted=%v err=%v", inserted, err)
	}
	var n int
	if err := db.QueryRow(`SELECT COUNT(1) FROM fill_record WHERE trade_id = ?`, "T-1").Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("duplicate trade insert ignored: count=%d", n)
	}
}
