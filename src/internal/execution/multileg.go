package execution

import (
	"context"
	"fmt"

	"github.com/dfr/optitrade/src/internal/deribit"
	"github.com/dfr/optitrade/src/internal/state"
	"github.com/dfr/optitrade/src/internal/strategy"
)

// MultiLegMVP documents combo execution for options structures built from [strategy] templates (T047).
//
// Deribit exposes native combo/block flows for some products; for MVP we place each [strategy.LegSpec]
// sequentially with private/buy|sell. Every leg shares the same user label so [deribit.REST.CancelByLabel]
// can unwind the batch if a later leg fails (best-effort rollback: already-matched legs may need manual handling).
//
// Pair with exits that set OrderRecord.ReduceOnly when flattening legs so the exchange enforces reduction.
const MultiLegMVP = "sequential_legs_shared_label"

// LegOrder is one leg ready for placement (maps from strategy.LegSpec + economics).
type LegOrder struct {
	Instrument string
	Side       string // buy | sell
	Amount     float64
	Price      *float64
	Advanced   *string
}

// FromLegSpecs converts template legs into placement descriptors (same amount/price for each leg — refine per structure elsewhere).
func FromLegSpecs(legs []strategy.LegSpec, amount float64, price *float64, advanced *string) []LegOrder {
	out := make([]LegOrder, 0, len(legs))
	for _, l := range legs {
		out = append(out, LegOrder{
			Instrument: l.Instrument,
			Side:       string(l.Side),
			Amount:     amount,
			Price:      price,
			Advanced:   advanced,
		})
	}
	return out
}

// PlaceLegsSequential places each persisted order row in order; on failure, cancels by comboLabel (T047 rollback MVP).
func PlaceLegsSequential(ctx context.Context, p *Placer, comboLabel string, legs []LegOrder, recs []*state.OrderRecord, correlationID string) error {
	if p == nil {
		return fmt.Errorf("nil placer")
	}
	if len(legs) != len(recs) {
		return fmt.Errorf("legs and order records length mismatch")
	}
	for i := range legs {
		rec := recs[i]
		leg := legs[i]
		rec.Label = comboLabel
		if rec.InstrumentName != leg.Instrument || rec.Side != leg.Side {
			return fmt.Errorf("leg %d: record does not match leg spec", i)
		}
		if _, err := p.PlaceLimit(ctx, rec, leg.Amount, leg.Price, leg.Advanced, correlationID); err != nil {
			if !p.DryRun && p.API != nil {
				_, _ = p.API.CancelByLabel(ctx, deribit.CancelByLabelParams{Label: comboLabel})
			}
			return fmt.Errorf("leg %d: %w", i, err)
		}
	}
	return nil
}
