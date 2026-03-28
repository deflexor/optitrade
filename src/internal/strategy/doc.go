// Package strategy builds defined-risk option structure candidates from policy,
// regime, and top-of-book liquidity. It never emits naked short legs from the
// bundled templates; see [ValidateDefinedRisk] for the structural checks.
//
// Deribit instrument name patterns (options on linear BTC/ETH perps index):
//
//	{BASE}-{D}{MON}{YY}-{STRIKE}-{C|P}
//
// BASE is the currency code (BTC, ETH). MON is English month abbreviation
// (JAN..DEC). STRIKE is an integer (no decimals). Suffix C or P is call/put.
//
// Examples:
//
//	BTC-28MAR25-90000-C
//	ETH-27JUN25-3200-P
//
// Vertical spreads reference two distinct strikes with the same BASE, expiry,
// and option type. An iron condor uses four instruments: two puts (different
// strikes) and two calls (different strikes) sharing one expiry.
package strategy
