package risk

import (
	"fmt"
	"math/big"
	"strings"
)

// ParseDecimalRat parses a finite decimal string into a *big.Rat.
func ParseDecimalRat(s string) (*big.Rat, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("risk: empty decimal")
	}
	r := new(big.Rat)
	if _, ok := r.SetString(s); !ok {
		return nil, fmt.Errorf("risk: invalid decimal %q", s)
	}
	return r, nil
}

func ratFromFloat64(f float64) *big.Rat {
	return new(big.Rat).SetFloat64(f)
}

func ratAbsCopy(r *big.Rat) *big.Rat {
	if r == nil {
		return new(big.Rat)
	}
	out := new(big.Rat).Set(r)
	if out.Sign() < 0 {
		out.Neg(out)
	}
	return out
}

func ratLessOrEqual(a, b *big.Rat) bool {
	return a.Cmp(b) <= 0
}
