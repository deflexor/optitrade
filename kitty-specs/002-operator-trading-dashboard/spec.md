# Feature Specification: Operator Trading Dashboard

**Feature Branch**: `002-operator-trading-dashboard`  
**Created**: 2026-03-28  
**Status**: Draft  
**Input**: User description: lightweight web UI for the Deribit trading bot: uptime, RAM, connection mode (test vs live), balance, P/L chart, market mood, open positions with close actions, recently closed positions with USD and percent P/L, strategy metadata (expected P/L, win %), position detail with legs, liquidity, metrics, Greeks, close modal with estimated exit P/L and wait vs close guidance, rebalance modal with suggested adjustments and outcomes. Operator-stated preference for a small, debuggable SPA stack (implementation detail captured in Assumptions only).

## Clarifications

### Session 2026-03-28

- Q: For this dashboard, how should access be secured in the first shipped version? -> A: Simple username/password registration and login only (no email, no verification codes). Login succeeds only for usernames on an operator-maintained allowlist (example: `opti`) after correct password verification; all other usernames receive the exact message `Sorry, feature not ready` and MUST NOT access dashboard data. Passwords MUST NOT be stored in plaintext.
- Q: The main total balance figure on the dashboard should match which meaning for Deribit? -> A: **Equity** (portfolio value: cash plus unrealized profit and loss; not wallet cash alone). The UI label MUST identify this figure as equity so operators do not confuse it with isolated margin or cash-only balances.
- Q: What should the default time range be for the P/L chart when the operator first opens the dashboard? -> A: **Last 30 days** (rolling calendar-day window ending at snapshot time). Other ranges remain optional if the backend supports them.
- Q: For health and portfolio snapshots (uptime, RAM, connection, test vs live, equity, mood), what maximum age should a reading have before the UI treats it as stale? -> A: **5 seconds** after the snapshot timestamp (clock per server or agreed time source). Older data MUST show a visible stale or refreshing state for that summary strip.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - At-a-glance operational health (Priority: P1)

An operator opens the dashboard and immediately sees whether the bot process is healthy, how much memory it uses, whether the exchange connection is up, and whether the session is test or real-money.

**Why this priority**: Wrong environment or silent failure is a capital and safety risk; this is the minimum trust surface.

**Independent Test**: With a running bot in each mode (test and live), compare dashboard indicators to ground truth (process monitor, configured Deribit base URL or flags) and verify mismatches are visible within the UI refresh interval.

**Acceptance Scenarios**:

1. **Given** the bot is running, **when** the operator loads the dashboard, **then** they see process uptime and memory usage that track actual values within an agreed tolerance and update on a predictable schedule.
2. **Given** Deribit connectivity, **when** the dashboard loads, **then** the operator sees explicit indication of test vs real-money mode with no ambiguity.
3. **Given** a broken or missing exchange connection, **when** the dashboard refreshes, **then** the connection status shows degraded or disconnected and the operator can tell what failed without reading logs.
4. **Given** an unauthenticated visitor, **when** they open any dashboard URL, **then** they are required to register or sign in before seeing health, balances, or positions.
5. **Given** a username on the operator allowlist and correct password, **when** the user signs in, **then** they obtain a session and reach the main dashboard.
6. **Given** a username not on the allowlist, **when** the user attempts sign-in after registration or login, **then** they see the message `Sorry, feature not ready` and MUST NOT see portfolio or health data.
7. **Given** an allowlisted username and wrong password, **when** the user attempts sign-in, **then** access is denied with a generic failed-login outcome (MUST NOT display portfolio data; MUST NOT use the not-ready message reserved for non-allowlisted users).
8. **Given** a signed-in operator and a snapshot whose timestamp is more than **5 seconds** old relative to the agreed time source, **when** they view the summary strip, **then** they see an explicit stale or refreshing indicator for health, connection mode, equity, and mood fields rather than silent presentation as fresh.

---

### User Story 2 - Portfolio snapshot and market context (Priority: P2)

An operator wants total account balance, a profit and loss view over time, and the current market mood (the same notion of mood or regime the bot uses for decisions) so they can interpret positions in context.

**Why this priority**: Supports supervision and aligns human understanding with automated behavior.

**Independent Test**: Compare dashboard balance and P/L series to exchange-reported or internally reconciled figures; verify mood label matches the engine's published mood or regime for the same timestamp.

**Acceptance Scenarios**:

1. **Given** the bot has a live Deribit session and the operator is authenticated to the dashboard, **when** the operator views the summary, **then** the primary balance shown is **account equity** (cash plus unrealized P/L) reconciled to the exchange at the same timestamp as the rest of the snapshot, and the label makes clear it is equity rather than cash-only.
2. **Given** historical P/L snapshots exist, **when** an authenticated operator opens the chart without changing settings, **then** the default horizon is the **last 30 days** (rolling), axes are labeled, and the active range is visible in the UI. If fewer than 30 days of history exist, the chart spans all available history and states that fact.
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
- P/L chart: history shorter than 30 days MUST still render with correct span and MUST NOT imply missing data beyond the stated range.
- Summary snapshots older than **5 seconds** MUST surface staleness; clock skew between client and server MUST be handled (for example by trusting server-emitted snapshot age or synchronized time) so operators are not falsely warned on every tick.
- Auth: brute-force or credential-stuffing attempts SHOULD be rate-limited or delayed per deployment policy; session tokens MUST expire or be revocable on restart as documented for operators.
- Auth: duplicate registration, weak passwords, or recovery: v1 has no email recovery; operators reset accounts via deployment procedures.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The dashboard MUST present bot process uptime and memory usage on each visit and on periodic refresh.
- **FR-002**: The dashboard MUST show Deribit connection status and MUST distinguish test environment from real-money trading in a single unambiguous indicator or label pair.
- **FR-003**: The dashboard MUST show the operator's primary total as **account equity** (portfolio value including unrealized P/L, not wallet cash alone), reconciled to the same Deribit-sourced snapshot as other portfolio fields, and MUST label it clearly as equity on screen.
- **FR-004**: The dashboard MUST render a profit-and-loss time series whose **default** range is the **last 30 days** (rolling, end-aligned to the latest snapshot). If the backend exposes additional windows, the operator MUST be able to switch among them without losing access to the default. If stored history is shorter than 30 days, the chart MUST cover available history only and MUST indicate the actual span.
- **FR-005**: The dashboard MUST display current market mood or regime consistent with the bot's published classification for the same moment.
- **FR-006**: The dashboard MUST list all currently open positions with strategy name, modeled expected P/L at open, and strategy win-rate statistic (per operator-configured statistics window defined by the backend).
- **FR-007**: The dashboard MUST list the ten most recently closed positions in reverse chronological order with realized P/L in USD and as a percentage.
- **FR-008**: From an open position row, the operator MUST be able to start a close-now flow that leads to the confirmation dialog described in FR-010.
- **FR-009**: Selecting a position MUST open a detail view showing legs, per-leg liquidity context, summary metrics, and Greeks needed to supervise options risk.
- **FR-010**: The close-now flow MUST show estimated P/L if the position were closed immediately, and MUST include a system-generated recommendation to close now or wait when rules support such guidance, with short rationale.
- **FR-011**: The dashboard MUST offer a rebalance entry point on the position detail view that opens a dialog with ranked or numbered suggestions, each describing the adjustment and the likely effect on risk or P/L.
- **FR-012**: Destructive actions (close, rebalance execution) MUST require explicit confirmation after the operator has seen the latest estimates in the dialog.
- **FR-013**: The dashboard MUST reflect errors from the trading layer without implying success; after failures, displayed positions MUST reconcile to backend state within one full refresh cycle or show a reconciling state.
- **FR-014**: The product MUST provide registration using username and password only; it MUST NOT require email addresses, SMS, or verification codes in v1.
- **FR-015**: The product MUST provide sign-in that verifies password for allowlisted usernames only. Non-allowlisted usernames MUST receive the exact user-visible message `Sorry, feature not ready` and MUST NOT receive any authenticated session or protected API payload.
- **FR-016**: Allowlisted usernames with incorrect passwords MUST be denied login without issuing a session; the response MUST NOT reveal allowlist membership to unauthenticated callers beyond the distinct not-ready message for non-allowlisted usernames at attempted login.
- **FR-017**: All dashboard data and control actions (health, balances, positions, close, rebalance) MUST require an authenticated session established per FR-015.
- **FR-018**: Passwords MUST be stored using a salted, slow, one-way password hash suitable for verifier-based login (MUST NOT store plaintext passwords).
- **FR-019**: The UI MUST treat the **summary snapshot** (uptime, memory, Deribit connection and environment mode, equity, mood label) as **stale** when its bundled timestamp is more than **5 seconds** behind the current time on the agreed reference clock (normally server time supplied with the payload). Stale summaries MUST show a clear stale or refreshing state and MUST NOT be labeled or styled as live-current without that qualification.

### Key Entities

- **Operator user**: Username, password verifier material, allowlist flag or allowlist source reference, registration timestamp; no email or phone in v1.
- **Dashboard session**: Authenticated subject bound to one allowlisted operator user, issuance and expiry rules set in the implementation plan.
- **Dashboard snapshot**: Timestamped bundle of health metrics, connection mode, **equity** (primary balance), mood label, and freshness flags; **staleness threshold 5 seconds** on the reference clock for the summary strip.
- **P/L series**: Time-ordered points with currency and optional benchmark reference if provided by backend; default request horizon **30 days** unless operator selects another supported window.
- **Position summary**: Identifier, strategy name, open vs closed state, expected P/L at open, win-rate statistic source, unrealized or realized P/L.
- **Position leg**: Instrument, side, size, entry context, liquidity indicator, and leg-level Greeks or references.
- **Close preview**: Estimated exit P/L, scenario assumptions, recommendation (close, wait, or neutral), and optional countdown or quote timestamps.
- **Rebalance suggestion**: Proposed action set, expected risk or P/L impact description, and prerequisite conditions.

## Out of scope

- Mobile-native applications and push notifications.
- Enterprise IAM, OAuth2/OIDC providers, or multi-customer tenancy beyond username allowlist plus password auth in FR-014 to FR-018.
- Email or SMS verification, multi-factor authentication, and self-service password recovery workflows in v1 (operators recover via deployment procedures).
- Building a full general-purpose exchange front end or charting terminal beyond the described P/L and position views.
- Guaranteeing profitable outcomes from rebalance or close recommendations (guidance is probabilistic or rule-based only).

## Assumptions

- The trading engine and persistence layer expose stable operator APIs or events for health, balances, positions, closes, and rebalance proposals; this specification treats those contracts as deliverables of sibling work if not yet present.
- The operator maintains a username allowlist (including at least the exemplar `opti`) alongside credential storage; only allowlisted users may complete sign-in to the dashboard. Registration MAY accept credentials for usernames not yet allowlisted; those users remain blocked at login with the not-ready message until the operator adds them. Network exposure (localhost vs VPN vs public) is chosen in deployment; authentication in FR-014 to FR-018 is always required before protected data.
- Win-rate and "expected P/L" strings use definitions published by the backend (e.g., win rate over last N closed trades per strategy); the UI displays engine-provided numbers without recomputing strategy analytics client-side.
- The operator prefers a lightweight, easy-to-debug single-page client for the first version (stated preference: React with Tailwind, Zustand, Axios). Technology choices are not success criteria; they inform planning and implementation only.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Operators can identify test vs live mode and connection health in under 5 seconds of opening the dashboard when the bot is running.
- **SC-002**: **Equity** and open position counts shown on the dashboard match backend or exchange reconciliation sources in 100% of sampled checks during acceptance testing.
- **SC-003**: For scripted test accounts, open and close flows initiated from the UI succeed or fail with an explicit message in at least 95% of trials without silent inconsistency between UI and exchange state after a full refresh.
- **SC-004**: Operators report they can locate strategy name, win rate, and expected P/L for any open position without leaving the dashboard (validated via acceptance checklist or supervised session).
- **SC-005**: Close and rebalance dialogs always show when estimates were computed or when data is stale, so operators are not asked to confirm on unknown-age pricing (zero tolerance for missing freshness indicator during acceptance).
- **SC-006**: In acceptance testing, every non-allowlisted username receives only the not-ready message and zero protected responses across API and UI probes; every allowlisted username with wrong password receives denial without session and without that not-ready message.
- **SC-007**: On first open of the P/L chart after sign-in, operators always see the **last 30 days** as the active range when at least that much history exists; when less history exists, they see the true shorter span called out (100% conformance in scripted acceptance checks).
- **SC-008**: In fault-injection tests where the snapshot source stops updating, the summary strip shows stale or refreshing within **6 seconds** of the last good snapshot (allowing one client poll cycle) in 100% of trials.
