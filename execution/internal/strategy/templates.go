package strategy

import (
	"fmt"
	"strings"
)

// LegSide is exchange order side for one option leg.
type LegSide string

const (
	LegBuy  LegSide = "buy"
	LegSell LegSide = "sell"
)

// LegSpec is one option leg with Deribit instrument name and side.
type LegSpec struct {
	Instrument string
	Side       LegSide
}

// OptionInstrumentName returns a Deribit option name: BASE-DMMMYY-STRIKE-C|P.
func OptionInstrumentName(base, expiry string, strike int64, call bool) string {
	sfx := "P"
	if call {
		sfx = "C"
	}
	return fmt.Sprintf("%s-%s-%d-%s", strings.ToUpper(strings.TrimSpace(base)), strings.ToUpper(strings.TrimSpace(expiry)), strike, sfx)
}

// VerticalPutCredit is a defined-risk short put vertical (collect premium): sell
// higher strike put, buy lower strike put. shortStrike is the sold (higher) leg.
func VerticalPutCredit(base, expiry string, shortStrike int64, width int) ([]LegSpec, error) {
	if width <= 0 {
		width = DefaultStrikeWidth
	}
	if shortStrike <= int64(width) {
		return nil, fmt.Errorf("put credit: short strike too low for width %d", width)
	}
	longStrike := shortStrike - int64(width)
	legs := []LegSpec{
		{Instrument: OptionInstrumentName(base, expiry, shortStrike, false), Side: LegSell},
		{Instrument: OptionInstrumentName(base, expiry, longStrike, false), Side: LegBuy},
	}
	mustAssertDefinedRiskInDev(legs)
	return legs, nil
}

// VerticalCallDebit is a defined-risk long call vertical: buy lower strike
// call, sell higher strike call. longStrike is the purchased (lower) leg.
func VerticalCallDebit(base, expiry string, longStrike int64, width int) ([]LegSpec, error) {
	if width <= 0 {
		width = DefaultStrikeWidth
	}
	shortStrike := longStrike + int64(width)
	legs := []LegSpec{
		{Instrument: OptionInstrumentName(base, expiry, longStrike, true), Side: LegBuy},
		{Instrument: OptionInstrumentName(base, expiry, shortStrike, true), Side: LegSell},
	}
	mustAssertDefinedRiskInDev(legs)
	return legs, nil
}

// IronCondor builds four legs: long put K1, short put K2, short call K3, long call K4
// with K1 < K2 < K3 < K4 (classic defined-risk iron condor).
func IronCondor(base, expiry string, k1, k2, k3, k4 int64) ([]LegSpec, error) {
	if !(k1 < k2 && k2 < k3 && k3 < k4) {
		return nil, fmt.Errorf("iron condor: require K1<K2<K3<K4, got %d,%d,%d,%d", k1, k2, k3, k4)
	}
	legs := []LegSpec{
		{Instrument: OptionInstrumentName(base, expiry, k1, false), Side: LegBuy},
		{Instrument: OptionInstrumentName(base, expiry, k2, false), Side: LegSell},
		{Instrument: OptionInstrumentName(base, expiry, k3, true), Side: LegSell},
		{Instrument: OptionInstrumentName(base, expiry, k4, true), Side: LegBuy},
	}
	mustAssertDefinedRiskInDev(legs)
	return legs, nil
}
