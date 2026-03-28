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

Implementation reference: `src/internal/strategy/doc.go` and template builders in the same package.

## 7. IV vs order book sanity (DVOL / index vs touch)

Deribit publishes volatility index (DVOL-style) time series alongside per-instrument order books. Before trusting IV-derived regime or signals:

1. **Clock alignment**: Compare volatility index timestamp to local book update time. Large lag suggests stale IV relative to the touch (veto `iv_stale`).
2. **Realized vs implied band (optional)**: Given a prior underlying mid and a short horizon Δt, a rough check is whether the relative jump in mid stays within a small multiple of σ√Δt when σ is taken from the index (calendar-time heuristic, not a production options calibration).

Implementation: `src/internal/cost` (`IVSanityOptions`, `ivBookConflict`) used from `ScoreCandidate` when `UseIVQuotes` is true.
