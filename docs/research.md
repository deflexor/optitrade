# Research notes

## Deribit option instrument names (BTC / ETH linear)

Linear options reference the underlying index (e.g. `BTC` / `ETH`) with this pattern:

`{BASE}-{D}{MON}{YY}-{STRIKE}-{C|P}`

- **BASE**: `BTC`, `ETH`, …
- **Expiry**: day + English month abbreviation + two-digit year (e.g. `28MAR25`). Deribit’s API returns this token in `instrument_name`; format it consistently when synthesizing names for strategy legs.
- **STRIKE**: integer
- **Suffix**: `C` (call) or `P` (put).

Examples: `BTC-28MAR25-90000-C`, `ETH-27JUN25-3200-P`.

Vertical spreads use two names sharing BASE, expiry, and option type. Iron condors use four names (two puts, two calls) on one expiry with four ascending strikes.

Implementation reference: `execution/internal/strategy/doc.go` and template builders in the same package.
