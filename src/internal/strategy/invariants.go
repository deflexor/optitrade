package strategy

import (
	"fmt"
	"os"
	"strings"
)

// ValidateDefinedRisk checks that legs do not expose naked short options: every
// sell must pair with a protective long on the same option type and expiry.
func ValidateDefinedRisk(legs []LegSpec) error {
	if len(legs) == 0 {
		return fmt.Errorf("no legs")
	}
	switch len(legs) {
	case 2:
		return validateVertical(legs)
	case 4:
		return validateIronCondor(legs)
	default:
		return fmt.Errorf("unsupported leg count %d for defined-risk template", len(legs))
	}
}

func validateVertical(legs []LegSpec) error {
	a, b := legs[0], legs[1]
	t1, k1, c1, err := parseOptionInstrument(a.Instrument)
	if err != nil {
		return err
	}
	t2, k2, c2, err := parseOptionInstrument(b.Instrument)
	if err != nil {
		return err
	}
	if t1 != t2 || c1 != c2 {
		return fmt.Errorf("vertical requires same expiry and call/put flag")
	}
	if k1 == k2 {
		return fmt.Errorf("vertical requires two distinct strikes")
	}
	var sells, buys int
	for _, lg := range legs {
		switch lg.Side {
		case LegSell:
			sells++
		case LegBuy:
			buys++
		default:
			return fmt.Errorf("leg %q: invalid side %q", lg.Instrument, lg.Side)
		}
	}
	if sells != 1 || buys != 1 {
		return fmt.Errorf("vertical requires exactly one buy and one sell")
	}
	return nil
}

func validateIronCondor(legs []LegSpec) error {
	type key struct {
		exp  string
		call bool
	}
	by := make(map[key][]LegSpec)
	for _, lg := range legs {
		exp, strike, call, err := parseOptionInstrument(lg.Instrument)
		if err != nil {
			return err
		}
		_ = strike
		k := key{exp: exp, call: call}
		by[k] = append(by[k], lg)
	}
	if len(by) != 2 {
		return fmt.Errorf("iron condor expects 2 puts + 2 calls (same expiry)")
	}
	for _, group := range by {
		if len(group) != 2 {
			return fmt.Errorf("iron condor: need two legs per wing, got %d", len(group))
		}
		var sells, buys int
		for _, lg := range group {
			switch lg.Side {
			case LegSell:
				sells++
			case LegBuy:
				buys++
			default:
				return fmt.Errorf("leg %q: invalid side", lg.Instrument)
			}
		}
		if sells != 1 || buys != 1 {
			return fmt.Errorf("each wing must have one buy and one sell")
		}
	}
	return nil
}

func parseOptionInstrument(name string) (expiry string, strike int64, call bool, err error) {
	parts := strings.Split(name, "-")
	if len(parts) < 4 {
		return "", 0, false, fmt.Errorf("invalid option instrument %q", name)
	}
	// BASE-DMMMYY-STRIKE-C|P — strike may be last-2; suffix last.
	suffix := parts[len(parts)-1]
	switch suffix {
	case "C":
		call = true
	case "P":
		call = false
	default:
		return "", 0, false, fmt.Errorf("option %q: missing C/P suffix", name)
	}
	var k int64
	if _, scanErr := fmt.Sscanf(parts[len(parts)-2], "%d", &k); scanErr != nil {
		return "", 0, false, fmt.Errorf("option %q: strike: %w", name, scanErr)
	}
	expiry = parts[len(parts)-3]
	return expiry, k, call, nil
}

func mustAssertDefinedRiskInDev(legs []LegSpec) {
	if !devInvariantPanic() {
		return
	}
	if err := ValidateDefinedRisk(legs); err != nil {
		panic(fmt.Sprintf("strategy: defined-risk invariant: %v", err))
	}
}

func devInvariantPanic() bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv("OPTITRADE_DEV_INVARIANTS")))
	return v == "1" || v == "true" || v == "yes"
}
