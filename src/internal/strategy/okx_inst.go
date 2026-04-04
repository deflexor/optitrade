package strategy

import (
	"fmt"
	"strings"
)

// OKXOptionInstID returns an OKX option instrument id:
// {BASE}-USD-{YYMMDD}-{strike}-{C|P}. base is uppercased; expiryYYYYMMDD must be
// 8 digits (YYYYMMDD); YYMMDD is expiry[2:8].
func OKXOptionInstID(base, expiryYYYYMMDD string, strike int64, call bool) string {
	b := strings.ToUpper(strings.TrimSpace(base))
	exp := strings.TrimSpace(expiryYYYYMMDD)
	yyMMdd := exp
	if len(exp) == 8 {
		yyMMdd = exp[2:8]
	}
	sfx := "P"
	if call {
		sfx = "C"
	}
	return fmt.Sprintf("%s-USD-%s-%d-%s", b, yyMMdd, strike, sfx)
}

// VerticalPutCreditOKX is a defined-risk short put vertical (collect premium)
// using OKX option instIds: sell higher strike put, buy lower strike put.
// shortStrike is the sold (higher) leg.
func VerticalPutCreditOKX(base, expiryYYYYMMDD string, shortStrike int64, width int) ([]LegSpec, error) {
	if width <= 0 {
		width = DefaultStrikeWidth
	}
	if shortStrike <= int64(width) {
		return nil, fmt.Errorf("put credit: short strike too low for width %d", width)
	}
	longStrike := shortStrike - int64(width)
	legs := []LegSpec{
		{Instrument: OKXOptionInstID(base, expiryYYYYMMDD, shortStrike, false), Side: LegSell},
		{Instrument: OKXOptionInstID(base, expiryYYYYMMDD, longStrike, false), Side: LegBuy},
	}
	mustAssertDefinedRiskInDev(legs)
	return legs, nil
}
