package execution

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/dfr/optitrade/src/internal/audit"
	"github.com/dfr/optitrade/src/internal/deribit"
	"github.com/dfr/optitrade/src/internal/session"
	"github.com/dfr/optitrade/src/internal/state"
)

// Placer submits limit orders with post-only/reduce-only taken from the persisted OrderRecord (T046).
type Placer struct {
	API PrivateREST
	// Orders must support GetOrder and UpdateOrder.
	Orders state.OrderRepository
	// Audit, when non-nil, logs a row for each successful RPC submit (WP11).
	Audit  audit.DecisionLogger
	DryRun bool
	// Session, when non-nil, enforces protective mode before RPC submit (WP12 / FR-009).
	Session *session.FSM
}

// NewPlacer returns a placer; callers should set OrderRecord.PostOnly true for entries (Deribit maker default).
func NewPlacer(api PrivateREST, orders state.OrderRepository) *Placer {
	return &Placer{API: api, Orders: orders}
}

// PlaceLimit submits private/buy or private/sell for a limit order already stored with state.OrderStateNew (or re-entrant if no exchange id yet).
// amount is the Deribit order size float parsed from APIs; price/advanced may be nil for non-limit flows later.
func (p *Placer) PlaceLimit(ctx context.Context, rec *state.OrderRecord, amount float64, price *float64, advanced *string, correlationID string) (*deribit.PlacedOrderResponse, error) {
	if p == nil || p.API == nil || p.Orders == nil || rec == nil {
		return nil, errors.New("execution: nil placer or order")
	}
	if p.Session != nil {
		if err := p.Session.AllowSubmit(rec.ReduceOnly); err != nil {
			return nil, err
		}
	}
	if existing, err := p.Orders.GetOrder(ctx, rec.InternalOrderID); err == nil {
		if existing.ExchangeOrderID != nil && *existing.ExchangeOrderID != "" {
			return nil, nil
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	lim := "limit"
	params := deribit.PlaceOrderParams{
		InstrumentName: rec.InstrumentName,
		Amount:         &amount,
		Type:           &lim,
		Label:          &rec.Label,
		Price:          price,
		Advanced:       advanced,
	}
	po := rec.PostOnly
	params.PostOnly = &po
	ro := rec.ReduceOnly
	params.ReduceOnly = &ro

	nowMs := time.Now().UnixMilli()

	if p.DryRun {
		ex := "dry-run"
		rec.ExchangeOrderID = &ex
		rec.State = state.OrderStateOpen
		rec.UpdatedAt = nowMs
		if err := p.Orders.UpdateOrder(ctx, rec); err != nil {
			return nil, err
		}
		return &deribit.PlacedOrderResponse{
			Order: deribit.OrderDetail{OrderID: ex, InstrumentName: rec.InstrumentName},
		}, nil
	}

	var resp *deribit.PlacedOrderResponse
	var err error
	switch rec.Side {
	case "buy":
		resp, err = p.API.Buy(ctx, params)
	case "sell":
		resp, err = p.API.Sell(ctx, params)
	default:
		return nil, errors.New("execution: order side must be buy or sell")
	}
	if err != nil {
		return nil, err
	}

	st := state.OrderStateOpen
	if resp.Order.OrderState != nil {
		st = MapExchangeOrderState(*resp.Order.OrderState)
	}
	oid := resp.Order.OrderID
	rec.ExchangeOrderID = &oid
	rec.State = st
	rec.UpdatedAt = nowMs
	if err := p.Orders.UpdateOrder(ctx, rec); err != nil {
		return nil, err
	}

	if p.Audit != nil && len(correlationID) >= 8 {
		gate := map[string]any{
			"internal_order_id": rec.InternalOrderID,
			"instrument":        rec.InstrumentName,
			"exchange_order_id": oid,
			"side":              rec.Side,
			"post_only":         rec.PostOnly,
			"reduce_only":       rec.ReduceOnly,
		}
		_ = p.Audit.LogDecision(ctx, audit.DecisionRecord{
			CorrelationID:    correlationID,
			DecisionType:     "order_submit",
			CandidateID:      rec.CandidateID,
			RegimeLabel:      "-",
			CostModelVersion: "-",
			GateResults:      gate,
			Reason:           audit.ReasonOrderSubmit,
			TsMs:             nowMs,
		})
	}

	return resp, nil
}

// ParseAmount parses rec.Amount as float for Deribit placement.
func ParseAmount(rec *state.OrderRecord) (float64, error) {
	if rec == nil {
		return 0, errors.New("nil order")
	}
	return strconv.ParseFloat(rec.Amount, 64)
}

// PriceFloat returns rec.Price as float for limit placement, or (nil, nil) if absent.
func PriceFloat(rec *state.OrderRecord) (*float64, error) {
	if rec == nil || rec.Price == nil || *rec.Price == "" {
		return nil, nil
	}
	v, err := strconv.ParseFloat(*rec.Price, 64)
	if err != nil {
		return nil, err
	}
	return &v, nil
}
