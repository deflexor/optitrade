package strategy

import (
	"fmt"
	"strings"
)

// BuildLegsForStructure returns deterministic legs for a policy playbook
// allowed_structures token (WP13 / SC-001 certification). Names match
// config/examples/policy.example.json.
func BuildLegsForStructure(structureName string) ([]LegSpec, error) {
	switch strings.ToLower(strings.TrimSpace(structureName)) {
	case "credit_spread":
		return VerticalPutCredit("BTC", "27JUN26", 95000, 500)
	case "debit_spread":
		return VerticalCallDebit("BTC", "27JUN26", 95000, 500)
	case "iron_condor":
		return IronCondor("BTC", "27JUN26", 88000, 90000, 100000, 102000)
	default:
		return nil, fmt.Errorf("strategy: unknown structure name %q", structureName)
	}
}
