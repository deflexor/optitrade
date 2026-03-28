# Research: Autonomous Deribit Options Robot

**Feature**: `001-autonomous-deribit-options-robot`  
**Date**: 2026-03-28

## 1. Execution language and runtime

**Decision**: Go 1.22+ for the live execution service.

**Rationale**: Order placement, concurrent market feeds, and risk evaluation benefit from lightweight goroutines, static binaries, and predictable GC for a sub-minute anomaly response target.

**Alternatives considered**:

- **Python asyncio only**: Faster research-to-prod iteration but weaker ergonomics for strict latency and typed concurrent order state machines in production.
- **Rust**: Excellent for HFT; rejected for MVP due to higher implementation cost and team assumptions embedded in existing brief (Go recommended).

## 2. Research and backtest stack

**Decision**: Python 3.12+ with `pandas` / `numpy` and `pytest` for offline studies; outputs are **versioned parameters** (thresholds, playbook weights) consumed by execution via config files.

**Rationale**: Aligns with root product brief; isolates experimental code from the live key path.

**Alternatives considered**:

- **Go-only backtests**: Possible later; rejected for MVP to preserve analyst velocity.

## 3. Exchange integration pattern

**Decision**: Deribit **JSON-RPC 2.0** over HTTPS for request/response; **WebSocket** for subscriptions (ticker, book, user events) with a single **reconnect + state reconciliation** policy documented in execution code.

**Rationale**: Matches Deribit API model; WS required for low-latency book updates relative to polling.

**Alternatives considered**:

- **REST-only polling**: Simpler but fails book freshness and protective-mode timing goals.

## 4. Regime classifier (MVP)

**Decision**: Start with a **rule-based** classifier driven by exchange vol index features plus realized/spread stress flags; persist **regime label + version** on each decision audit row. Allow substitution with ML later without changing risk gate interfaces.

**Rationale**: Spec requires three labels and auditability; opaque ML is harder to certify for SC-005.

**Alternatives considered**:

- **ML-first** (HMM / classifier on custom features): Defer until baseline rule set proves insufficient.

## 5. Cost and slippage model

**Decision**: Explicit **fee schedule** from config, **half-spread** proxy from top of book, **slippage** bound from depth sweep or fixed conservative bps per underlying; **adverse selection** as optional additive bps per regime. Veto if net edge `<= 0` after all terms.

**Rationale**: Maps directly to FR-006 and root brief; testable with table-driven inputs.

**Alternatives considered**:

- **Full microstructure model**: Out of scope for MVP.

## 6. Storage

**Decision**: **SQLite** file per deployment with WAL mode; migrations for schema; nightly backup optional via ops.

**Rationale**: Root brief; fits single-account MVP; satisfies constitution parameterized query requirement.

**Alternatives considered**:

- **Embedded etcd / Postgres**: Only if multi-instance HA required later.

## 7. IV order staleness (spec edge case)

**Decision**: If IV-based order type is used, execution MUST cross-check against **recent book mid/spread** and **short-horizon move detectors** (e.g., jump in implied vol from book vs last IV quote ts); on conflict, **downgrade to plain limits** or **skip** trade.

**Rationale**: Directly addresses spec edge case on fast spikes.

**Alternatives considered**: IV orders disabled entirely for MVP; viable if timeline tight (document as config flag).

## 8. Open items for tasks phase

- Exact Deribit **combo / instrument name** conventions for each playbook leg set.
- Default numeric thresholds for liquidity filters (min depth, max spread bps per underlying).
- Whether MVP ships with **testnet-only** default in `quickstart.md` (recommended).
