# Feature Specification: Autonomous Deribit Options Robot

**Feature Branch**: `001-autonomous-deribit-options-robot`  
**Created**: 2026-03-28  
**Status**: Draft  
**Input**: Derived from project `plan.md` (product brief for autonomous Deribit options trading with regime-aware, defined-risk execution).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Safe risk envelope (Priority: P1)

An operator runs the system against a live account and needs absolute certainty that no trade can violate configured loss, exposure, or Greeks limits, and that the system stops opening risk when connectivity or data quality fails.

**Why this priority**: Capital preservation and compliance with the stated non-goals (no naked short options, no runaway sizing) outweigh profitable trading.

**Independent Test**: Configure tight limits and simulated bad data; verify no new risk-increasing activity occurs after gates trigger and that exits use reduce-only style behavior where applicable.

**Acceptance Scenarios**:

1. **Given** active positions and live connectivity, **when** a hypothetical new trade would breach max loss per trade, max portfolio delta, max portfolio vega, or max open premium at risk, **then** the trade is rejected and existing risk is unchanged by that decision.
2. **Given** a running session, **when** market data feed is lost, authentication fails, or order book quality crosses a configured “gap” threshold, **then** the system enters a protective mode: no new risk-increasing orders, and flatten or freeze proceeds per operator policy.
3. **Given** open orders, **when** max open orders per instrument is reached, **then** additional orders for that instrument are not submitted until capacity is freed.

---

### User Story 2 - Cost-aware edge execution (Priority: P2)

An operator wants the system to open and close **only** liquid BTC/ETH option structures when expected edge after all trading costs and adverse selection exceeds a threshold, using conservative sizing.

**Why this priority**: This is the core value proposition after safety is guaranteed.

**Independent Test**: Replay recorded market conditions with known fee/spread/slippage assumptions; verify trades only occur when net edge is positive under the cost model and that position size respects caps.

**Acceptance Scenarios**:

1. **Given** a candidate structure on a liquid expiry/strike, **when** expected value after fees, half-spread, and estimated slippage is not positive, **then** the system does not open the position.
2. **Given** approval to trade, **when** placing orders, **then** the system prefers passive (maker-style) execution first and only crosses aggressively within defined rules when residual edge still clears the cost hurdle.
3. **Given** an open defined-risk position, **when** exit is required, **then** the system uses spread/combination style exits where possible and avoids blasting full notional as an uncontrolled market-style exit unless policy allows and risk checks pass.

---

### User Story 3 - Regime-aware playbooks (Priority: P3)

An operator wants volatility regime (low, normal, high) to drive which defined-risk playbook applies—for example income-style structures in calm markets and directional/vol structures appropriate for stressed markets—without manual intervention each session.

**Why this priority**: Adapts behavior to market conditions; builds on P1/P2.

**Independent Test**: Feed historical or synthetic series classified into regimes; verify playbook selection and audit logs show regime label at time of decision.

**Acceptance Scenarios**:

1. **Given** a low-volatility regime label, **when** the system selects a structure, **then** allowed templates include conservative credit spreads and iron condors (not naked shorts).
2. **Given** a high-volatility regime label, **when** the system selects a structure, **then** allowed templates favor defined-risk debit spreads appropriate to high vol (within the same liquidation and risk gates).
3. **Given** a normal regime, **when** the system operates, **then** behavior follows the configured neutral playbook without contradicting P1 limits.

---

### User Story 4 - Observability and session control (Priority: P4)

An operator needs continuous visibility into positions, working orders, fills, exposure, PnL, and system state (running, paused, flattening, frozen), and confidence that anomalies surface clearly.

**Why this priority**: Operational trust and manual override paths.

**Independent Test**: Trigger state transitions; verify logs or operator-visible status reflect fills, cancels, and risk snapshots within agreed freshness.

**Acceptance Scenarios**:

1. **Given** ongoing trading, **when** orders fill partially or fully, **then** position and PnL views update and remain consistent with exchange-reported state after reconciliation.
2. **Given** operator command to cancel by label or flatten an instrument, **then** the system issues coordinated cancels and does not leave unintended orphan legs beyond a bounded reconciliation window.

---

### Edge Cases

- Very fast volatility spikes: implied-vol-based quoting can lag; decisions must not rely solely on slowly updating vol quotes when the order book shows dislocation.
- Wide spreads or thin depth at otherwise “listed” strikes: candidate generator must skip illiquid strikes/expiry even if nominally listed.
- Partial fills on multi-leg structures: risk and hedge integrity must be monitored; legs may need coordinated cancel/replace.
- Exchange partial outages: risk posture must default safe (no new risk) until health checks pass.
- Clock skew or session rollover: expiry selection and time-in-trade limits must behave deterministically across session boundaries.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST connect to the operator’s Deribit account using credentials and permissions supplied by the operator and MUST refuse to trade if authentication cannot be established securely.
- **FR-002**: The system MUST discover tradable option expiries and strikes from the exchange and MUST restrict candidates to liquid BTC and ETH options only (per configurable liquidity thresholds).
- **FR-003**: The system MUST ingest continuous pricing, order book, and volatility-index features sufficient to classify regime (low / normal / high) and to estimate transaction costs.
- **FR-004**: The system MUST maintain authoritative views of open orders, positions, and account summaries sufficient for pre-trade and post-trade risk checks.
- **FR-005**: The system MUST classify each decision interval into a volatility regime and MUST select strategies only from the playbook allowed for that regime, using defined-risk structures only (no naked short options).
- **FR-006**: The system MUST score each candidate trade net of explicit fees, spread, and slippage assumptions and MUST veto trades that do not clear a positive net-edge threshold after costs.
- **FR-007**: The system MUST enforce hard limits including: max loss per trade, max daily loss, max open premium at risk, max portfolio delta, max portfolio vega, max open orders per instrument, and max time-in-trade (values operator-configured within safe bounds).
- **FR-008**: The system MUST implement execution preferences: prioritize passive limit orders, use combination orders for spreads where supported, and use reduce-only semantics for exits where applicable.
- **FR-009**: The system MUST detect anomaly conditions including feed loss, auth failure, and abnormal book gaps and MUST enter mandatory flatten or freeze behavior per operator policy without opening new risk.
- **FR-010**: The system MUST provide operator-visible auditability: enough logged context to reconstruct regime, cost assumptions, risk gate outcomes, and order lifecycle for each trade (without logging secrets).
- **FR-011**: The system MUST NOT increase position size using loss-recovery (“martingale”) rules, MUST NOT pursue cross-exchange arbitrage, MUST NOT optimize leverage beyond configured caps, and MUST NOT “trade every signal” without passing cost and risk gates.

### Key Entities

- **Instrument**: Underlying (BTC/ETH), type (option), expiry, strike, exchange identifiers; liquidity metadata used for candidate filtering.
- **Market snapshot**: Top-of-book and depth summaries, vol index features, timestamps, and data-quality flags feeding regime and cost models.
- **Regime state**: Label (low/normal/high), confidence or persistence rules, and time range valid for playbook selection.
- **Trade candidate**: Structure template (e.g., spread types permitted in current regime), legs, intended size, expected edge after costs, and references to quotes used.
- **Order/working order**: Client order id, label, side, price/type, post-only flags, instrument or combo reference, and linkage to parent strategy.
- **Position / portfolio risk snapshot**: Per-leg and net Greeks, premium at risk, unrealized PnL aggregates used in gates.
- **Risk policy**: Limit thresholds, anomaly triggers, flatten vs freeze preferences, and session constraints.
- **Session / system state**: Running, paused, protective mode (flatten/frozen), and reason codes for operator visibility.

## Out of scope

- Naked short options, martingale sizing, cross-exchange arbitrage, deliberate trading of illiquid strikes, leverage chasing beyond caps, and firing on every raw signal without cost and risk gating.

## Assumptions

- Primary market is Deribit; scope is BTC and ETH listed options as described in the input brief.
- Operators are qualified to configure API keys, limits, and playbook parameters; regulatory and tax obligations remain with the operator.
- A separate research/backtesting workflow may exist to validate strategies before live parameters are used, but this specification treats **live** risk and execution behavior as in scope for P1–P4.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: In certification scenarios with seeded market data, **100%** of structures that complete execution are defined-risk templates permitted by the active regime playbook (no naked short option outcomes).
- **SC-002**: In stress scenarios where cumulative losses reach the configured daily cap, **100%** of test runs show no additional risk-increasing orders after the cap is hit until operator reset or the next defined session boundary.
- **SC-003**: In feed-loss and auth-failure simulations, protective mode activates within **60 seconds** of detection and no new risk-increasing orders are confirmed afterward in **100%** of runs.
- **SC-004**: Operators can reconcile positions and orders against the exchange after any trading window with **no unexplained orphan legs** beyond a bounded reconciliation procedure documented in runbooks (target: zero orphans in acceptance tests).
- **SC-005**: Audit records tie at least **90%** of live entries/exits (by count) to a recorded regime label, cost model version, and risk gate outcome in acceptance sampling—rising to **100%** for trades that breached a soft warning threshold.

## Notes

- Detailed numeric thresholds (edge after costs, liquidity scores, book-gap definitions) are operator-tunable parameters and SHOULD be finalized before live deployment; defaults must be conservative.
