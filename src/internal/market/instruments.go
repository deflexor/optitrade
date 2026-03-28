package market

import (
	"context"
	"fmt"
	"strings"

	"github.com/dfr/optitrade/src/internal/deribit"
)

const kindOption = "option"

// InstrumentsSource is the minimal surface needed for discovery (implemented by *deribit.REST).
type InstrumentsSource interface {
	GetInstruments(ctx context.Context, params *deribit.GetInstrumentsParams) ([]deribit.Instrument, error)
}

// InstrumentFilter narrows public/get_instruments results to an options universe.
type InstrumentFilter struct {
	// BaseCurrencies lists underlying currencies (e.g. "BTC", "ETH"). Empty defaults to BTC and ETH.
	BaseCurrencies []string
	// IncludeInactive allows instruments with is_active false.
	IncludeInactive bool
	// MaxInstruments caps the returned slice (0 = no cap).
	MaxInstruments int
	// NameWhitelist, if non-empty, keeps only instrument_name keys present in the map.
	NameWhitelist map[string]struct{}
}

// DiscoverActiveOptions fetches instruments per currency with kind=option and applies filters.
// Futures and other kinds never appear because the API request uses kind=option only.
func DiscoverActiveOptions(ctx context.Context, api InstrumentsSource, f InstrumentFilter) ([]deribit.Instrument, error) {
	if api == nil {
		return nil, fmt.Errorf("market: nil instruments source")
	}
	bases := f.BaseCurrencies
	if len(bases) == 0 {
		bases = []string{"BTC", "ETH"}
	}
	wantBase := make(map[string]struct{}, len(bases))
	for _, c := range bases {
		wantBase[strings.ToUpper(strings.TrimSpace(c))] = struct{}{}
	}

	kind := kindOption
	expired := false
	var out []deribit.Instrument
	for _, cur := range bases {
		cur = strings.ToUpper(strings.TrimSpace(cur))
		if cur == "" {
			continue
		}
		params := &deribit.GetInstrumentsParams{
			Currency: &cur,
			Kind:     &kind,
			Expired:  &expired,
		}
		batch, err := api.GetInstruments(ctx, params)
		if err != nil {
			return nil, err
		}
		for i := range batch {
			inst := batch[i]
			if !f.keep(inst, wantBase) {
				continue
			}
			out = append(out, inst)
			if f.MaxInstruments > 0 && len(out) >= f.MaxInstruments {
				return out, nil
			}
		}
	}
	return out, nil
}

func (f InstrumentFilter) keep(inst deribit.Instrument, wantBase map[string]struct{}) bool {
	if !f.IncludeInactive && !isActive(inst) {
		return false
	}
	if inst.Kind != nil && strings.ToLower(*inst.Kind) != kindOption {
		return false
	}
	bc := baseCurrency(inst)
	if _, ok := wantBase[bc]; !ok {
		return false
	}
	if len(f.NameWhitelist) > 0 {
		if _, ok := f.NameWhitelist[inst.InstrumentName]; !ok {
			return false
		}
	}
	return true
}

func isActive(inst deribit.Instrument) bool {
	if inst.IsActive == nil {
		return true
	}
	return *inst.IsActive
}

func baseCurrency(inst deribit.Instrument) string {
	if inst.BaseCurrency != nil {
		return strings.ToUpper(strings.TrimSpace(*inst.BaseCurrency))
	}
	return ""
}
