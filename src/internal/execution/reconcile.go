package execution

import (
	"context"
	"fmt"
	"time"

	"github.com/dfr/optitrade/src/internal/deribit"
	"github.com/dfr/optitrade/src/internal/state"
)

// SyncOrderIDPrefix tags rows we learned from get_open_orders but had no local client id (T050).
const SyncOrderIDPrefix = "sync:"

// Reconciler aligns SQLite order_record with exchange truth (T050).
type Reconciler struct {
	API    PrivateREST
	Orders state.OrderRepository
}

// Run pulls open orders, inserts missing locals, and closes orphans using get_order_state.
func (r *Reconciler) Run(ctx context.Context, filter *deribit.GetOpenOrdersParams) error {
	if r == nil || r.API == nil || r.Orders == nil {
		return fmt.Errorf("execution: nil reconciler")
	}
	open, err := r.API.GetOpenOrders(ctx, filter)
	if err != nil {
		return err
	}
	openByID := make(map[string]deribit.OpenOrder, len(open))
	for _, o := range open {
		openByID[o.OrderID] = o
	}

	locals, err := r.Orders.ListOrdersByStates(ctx, state.NonTerminalOrderStates())
	if err != nil {
		return err
	}
	now := time.Now().UnixMilli()

	localByEx := make(map[string]state.OrderRecord)
	for _, loc := range locals {
		if loc.ExchangeOrderID != nil && *loc.ExchangeOrderID != "" {
			localByEx[*loc.ExchangeOrderID] = loc
		}
	}

	for exID, remote := range openByID {
		if _, ok := localByEx[exID]; ok {
			continue
		}
		rec := orderRecordFromOpen(remote, now)
		if err := r.Orders.InsertOrder(ctx, rec); err != nil {
			return fmt.Errorf("reconcile insert %s: %w", exID, err)
		}
	}

	for _, loc := range locals {
		if loc.ExchangeOrderID == nil || *loc.ExchangeOrderID == "" {
			continue
		}
		exID := *loc.ExchangeOrderID
		if _, stillOpen := openByID[exID]; stillOpen {
			continue
		}
		detail, err := r.API.GetOrderState(ctx, exID)
		if err != nil {
			loc.State = state.OrderStateCanceled
			loc.UpdatedAt = now
			if err := r.Orders.UpdateOrder(ctx, &loc); err != nil {
				return err
			}
			continue
		}
		if detail != nil && detail.OrderState != nil {
			loc.State = MapExchangeOrderState(*detail.OrderState)
		} else {
			loc.State = state.OrderStateCanceled
		}
		loc.UpdatedAt = now
		if err := r.Orders.UpdateOrder(ctx, &loc); err != nil {
			return err
		}
	}
	return nil
}

func orderRecordFromOpen(o deribit.OpenOrder, now int64) *state.OrderRecord {
	label := ""
	if o.Label != nil {
		label = *o.Label
	}
	side := ""
	if o.Direction != nil {
		side = *o.Direction
	}
	ot := "limit"
	if o.OrderType != nil {
		ot = *o.OrderType
	}
	st := state.OrderStateOpen
	if o.OrderState != nil {
		st = MapExchangeOrderState(*o.OrderState)
	}
	var price *string
	if o.Price != nil {
		p := formatOrderPrice(o.Price)
		price = &p
	}
	amt := "0"
	if o.Amount != nil {
		amt = formatFloat(*o.Amount)
	}
	post := false
	if o.PostOnly != nil {
		post = *o.PostOnly
	}
	created := now
	if o.CreationTimestamp != nil && *o.CreationTimestamp > 0 {
		created = *o.CreationTimestamp
	}
	ex := o.OrderID
	return &state.OrderRecord{
		InternalOrderID: SyncOrderIDPrefix + o.OrderID,
		ExchangeOrderID: &ex,
		InstrumentName:  o.InstrumentName,
		Label:           label,
		Side:            side,
		OrderType:       ot,
		Price:           price,
		Amount:          amt,
		PostOnly:        post,
		ReduceOnly:      false,
		State:           st,
		CreatedAt: created,
		UpdatedAt: now,
	}
}

func formatOrderPrice(p *float64) string {
	if p == nil {
		return ""
	}
	return formatFloat(*p)
}
