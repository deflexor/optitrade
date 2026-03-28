# Feature Specification: Operator Trading Dashboard

**Feature Branch**: `002-operator-trading-dashboard`  
**Created**: 2026-03-28  
**Status**: Draft  
**Input**: User description: lightweight web UI for the Deribit trading bot: uptime, RAM, connection mode (test vs live), balance, P/L chart, market mood, open positions with close actions, recently closed positions with USD and percent P/L, strategy metadata (expected P/L, win %), position detail with legs, liquidity, metrics, Greeks, close modal with estimated exit P/L and wait vs close guidance, rebalance modal with suggested adjustments and outcomes. Operator-stated preference for a small, debuggable SPA stack (implementation detail captured in Assumptions only).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - At-a-glance operational health (Priority: P1)

An operator opens the dashboard and immediately sees whether the bot process is healthy, how much memory it uses, whether the exchange connection is up, and whether the session is test or real-money.

**Why this priority**: Wrong environment or silent failure is a capital and safety risk; this is the minimum trust surface.

**Independent Test**: With a running bot in each mode (test and live), compare dashboard indicators to ground truth (process monitor, configured Deribit base URL or flags) and verify mismatches are visible within the UI refresh interval.

**Acceptance Scenarios**:

1. **Given** the bot is running, **when** the operator loads the dashboard, **then** they see process uptime and memory usage that track actual values within an agreed tolerance and update on a predictable schedule.
2. **Given** Deribit connectivity, **when** the dashboard loads, **then** the operator sees explicit indication of test vs real-money mode with no ambiguity.
3. **Given** a broken or missing exchange connection, **when** the dashboard refreshes, **then** the connection status shows degraded or disconnected and the operator can tell what failed without reading logs.

---

### User Story 2 - Portfolio snapshot and market context (Priority: P2)

An operator wants total account balance, a profit and loss view over time, and the current market mood (the same notion of mood or regime the bot uses for decisions) so they can interpret positions in context.

**Why this priority**: Supports supervision and aligns human understanding with automated behavior.

**Independent Test**: Compare dashboard balance and P/L series to exchange-reported or internally reconciled figures; verify mood label matches the engine's published mood or regime for the same timestamp.

**Acceptance Scenarios**:

1. **Given** an authenticated exchange session, **when** the operator views the dashboard, **then** total balance reflects the same sourcing the bot uses for trading decisions (per documented definition: equity, wallet balance, or margin balance).
2. **Given** historical P/L snapshots exist, **when** the operator views the chart, **then** they can read P/L over a clear time window with labeled axes and a sensible default range.
3. **Given** the classifier produces a mood or regime, **when** the operator views the dashboard, **then** the displayed mood matches the bot's current classification label (or explicitly shows unknown if data is stale).

---

### User Story 3 - Open and recently closed positions (Priority: P3)

An operator lists all open positions and the ten most recent closed positions. Each row shows strategy name, expected profit and loss at open (as modeled by the bot), win rate, and for closed trades realized P/L in USD and percent. Open rows expose a control to close now.

**Why this priority**: Core operational view for risk and performance.

**Independent Test**: Open and close positions in a test account; verify lists, counts, sorting, and field values match backend records within rounding rules.

**Acceptance Scenarios**:

1. **Given** open positions exist, **when** the operator views the list, **then** each row includes strategy name, expected P/L (per bot definition), and win-rate statistic relevant to that strategy.
2. **Given** at least ten closed positions in history, **when** the operator views recent closes, **then** exactly the ten most recent closed positions appear in newest-first order with realized P/L in USD and percent.
3. **Given** an open position, **when** the operator chooses Close from the list, **then** they are guided into the same confirmation flow as the detailed position view (see User Story 4).

---

### User Story 4 - Position detail, exit guidance, and rebalance (Priority: P4)

An operator selects a position to inspect legs with liquidity, risk metrics, and Greeks. They can request close or rebalance. Closing opens a confirmation dialog that states estimated P/L if closed now and recommends waiting or proceeding based on bot rules. Rebalance opens a dialog with suggested adjustments and possible outcomes.

**Why this priority**: Deep control and education without forcing log diving; depends on list views being correct.

**Independent Test**: For a controlled book, verify leg-level fields, Greek magnitudes, modal estimates, and recommendations match offline or API-sourced calculations within tolerance; execute a close and confirm behavior matches confirmation.

**Acceptance Scenarios**:

1. **Given** a multi-leg position, **when** the operator opens the detail view, **then** they see each leg with liquidity indicators and key metrics including Greeks required for options oversight.
2. **Given** an open position, **when** the operator chooses Close, **then** a confirmation dialog shows estimated P/L at immediate exit, states the reasoning if the system recommends waiting, and requires explicit confirmation to proceed.
3. **Given** an open position, **when** the operator chooses Rebalance, **then** a dialog lists concrete suggested adjustments (instruments, sides, approximate sizing where applicable) and describes expected effect on risk or P/L in plain language.
4. **Given** the operator confirms close, **when** the action completes or fails, **then** the UI reflects the new position state or a clear error without silent partial updates.

---

### Edge Cases

- Stale market data: Greeks and estimated exit P/L must be labeled with timestamps or freshness; if data is too old, show blocking warning before destructive actions.
- Partially filled or broken multi-leg positions: detail view must show each leg state; close/rebalance may be disabled with reason when the book is inconsistent.
- Very rapid updates: lists must not reorder confusingly; stable sort keys or brief loading states are acceptable.
- Exchange or API errors during close/rebalance: operator sees failure message and book unchanged until reconciled.
- Zero or fewer than ten historical closes: recent-closes section shows available rows and empty state messaging.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The dashboard MUST present bot process uptime and memory usage on each visit and on periodic refresh.
- **FR-002**: The dashboard MUST show Deribit connection status and MUST distinguish test environment from real-money trading in a single unambiguous indicator or label pair.
- **FR-003**: The dashboard MUST show total balance using the same balance definition the trading engine documents for operator-facing totals.
- **FR-004**: The dashboard MUST render a profit-and-loss time series for an operator-relevant default window, with capability to adjust range if the backend exposes multiple windows.
- **FR-005**: The dashboard MUST display current market mood or regime consistent with the bot's published classification for the same moment.
- **FR-006**: The dashboard MUST list all currently open positions with strategy name, modeled expected P/L at open, and strategy win-rate statistic (per operator-configured statistics window defined by the backend).
- **FR-007**: The dashboard MUST list the ten most recently closed positions in reverse chronological order with realized P/L in USD and as a percentage.
- **FR-008**: From an open position row, the operator MUST be able to start a close-now flow that leads to the confirmation dialog described in FR-010.
- **FR-009**: Selecting a position MUST open a detail view showing legs, per-leg liquidity context, summary metrics, and Greeks needed to supervise options risk.
- **FR-010**: The close-now flow MUST show estimated P/L if the position were closed immediately, and MUST include a system-generated recommendation to close now or wait when rules support such guidance, with short rationale.
- **FR-011**: The dashboard MUST offer a rebalance entry point on the position detail view that opens a dialog with ranked or numbered suggestions, each describing the adjustment and the likely effect on risk or P/L.
- **FR-012**: Destructive actions (close, rebalance execution) MUST require explicit confirmation after the operator has seen the latest estimates in the dialog.
- **FR-013**: The dashboard MUST reflect errors from the trading layer without implying success; after failures, displayed positions MUST reconcile to backend state within one full refresh cycle or show a reconciling state.

### Key Entities

- **Dashboard snapshot**: Timestamped bundle of health metrics, connection mode, balance, mood label, and freshness flags.
- **P/L series**: Time-ordered points with currency and optional benchmark reference if provided by backend.
- **Position summary**: Identifier, strategy name, open vs closed state, expected P/L at open, win-rate statistic source, unrealized or realized P/L.
- **Position leg**: Instrument, side, size, entry context, liquidity indicator, and leg-level Greeks or references.
- **Close preview**: Estimated exit P/L, scenario assumptions, recommendation (close, wait, or neutral), and optional countdown or quote timestamps.
- **Rebalance suggestion**: Proposed action set, expected risk or P/L impact description, and prerequisite conditions.

## Out of scope

- Mobile-native applications and push notifications.
- Multi-tenant access control beyond what the deployment already provides (see Assumptions).
- Building a full general-purpose exchange front end or charting terminal beyond the described P/L and position views.
- Guaranteeing profitable outcomes from rebalance or close recommendations (guidance is probabilistic or rule-based only).

## Assumptions

- The trading engine and persistence layer expose stable operator APIs or events for health, balances, positions, closes, and rebalance proposals; this specification treats those contracts as deliverables of sibling work if not yet present.
- The operator runs the dashboard in a trusted environment (for example localhost or private network). Fine-grained authentication and authorization are assumed to match the chosen deployment pattern documented in the implementation plan if remote access is required.
- Win-rate and "expected P/L" strings use definitions published by the backend (e.g., win rate over last N closed trades per strategy); the UI displays engine-provided numbers without recomputing strategy analytics client-side.
- The operator prefers a lightweight, easy-to-debug single-page client for the first version (stated preference: React with Tailwind, Zustand, Axios). Technology choices are not success criteria; they inform planning and implementation only.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Operators can identify test vs live mode and connection health in under 5 seconds of opening the dashboard when the bot is running.
- **SC-002**: Balance and open position counts shown on the dashboard match backend reconciliation sources in 100% of sampled checks during acceptance testing.
- **SC-003**: For scripted test accounts, open and close flows initiated from the UI succeed or fail with an explicit message in at least 95% of trials without silent inconsistency between UI and exchange state after a full refresh.
- **SC-004**: Operators report they can locate strategy name, win rate, and expected P/L for any open position without leaving the dashboard (validated via acceptance checklist or supervised session).
- **SC-005**: Close and rebalance dialogs always show when estimates were computed or when data is stale, so operators are not asked to confirm on unknown-age pricing (zero tolerance for missing freshness indicator during acceptance).
