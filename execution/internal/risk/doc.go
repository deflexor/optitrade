// Package risk merges portfolio snapshots (exchange positions + optional local
// adjustments) and runs pre-trade gates against [config.Policy] limits.
//
// Greeks: when private/get_positions provides delta/vega, aggregates use those
// values. Missing per-leg Greeks are treated as zero (caller should refresh
// positions before decisions; stale Greeks are an operational risk).
//
// Money comparisons use [math/big.Rat] from decimal strings at policy boundaries.
package risk
