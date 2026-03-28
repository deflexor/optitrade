package risk

import (
	"math/big"

	"github.com/dfr/optitrade/src/internal/deribit"
)

// PortfolioSnapshot aggregates exchange-derived fields for limit gates (T040).
type PortfolioSnapshot struct {
	DeltaTotal    *big.Rat
	VegaTotal     *big.Rat
	PremiumAtRisk *big.Rat

	OpenOrdersByInstrument map[string]int
}

// BuildPortfolioSnapshot merges RPC positions and open orders with optional
// delta/vega adjustments (local overlays). Premium-at-risk uses initial_margin
// per position when present; otherwise abs(size)*mark_price as a documented proxy.
func BuildPortfolioSnapshot(
	positions []deribit.Position,
	openOrders []deribit.OpenOrder,
	deltaAdj, vegaAdj *big.Rat,
) PortfolioSnapshot {
	d := new(big.Rat)
	v := new(big.Rat)
	par := new(big.Rat)

	for i := range positions {
		p := &positions[i]
		if p.Delta != nil && !isBadFloat(*p.Delta) {
			d.Add(d, ratFromFloat64(*p.Delta))
		}
		if p.Vega != nil && !isBadFloat(*p.Vega) {
			v.Add(v, ratFromFloat64(*p.Vega))
		}
		var leg *big.Rat
		if p.InitialMargin != nil && *p.InitialMargin > 0 && !isBadFloat(*p.InitialMargin) {
			leg = ratFromFloat64(*p.InitialMargin)
		} else if p.MarkPrice != nil && *p.MarkPrice > 0 && !isBadFloat(*p.MarkPrice) && !isBadFloat(p.Size) {
			leg = new(big.Rat).Mul(ratFromFloat64(p.Size), ratFromFloat64(*p.MarkPrice))
			leg = ratAbsCopy(leg)
		}
		if leg != nil {
			par.Add(par, leg)
		}
	}

	if deltaAdj != nil {
		d.Add(d, deltaAdj)
	}
	if vegaAdj != nil {
		v.Add(v, vegaAdj)
	}

	byInst := map[string]int{}
	for i := range openOrders {
		name := openOrders[i].InstrumentName
		if name == "" {
			continue
		}
		byInst[name]++
	}

	return PortfolioSnapshot{
		DeltaTotal:             d,
		VegaTotal:              v,
		PremiumAtRisk:          par,
		OpenOrdersByInstrument: byInst,
	}
}

func isBadFloat(f float64) bool {
	return f != f || f+1 == f // NaN or ±Inf
}
