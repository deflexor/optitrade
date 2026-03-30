# Feature Specification: Operator dashboard trading and controls

**Feature Branch**: `002-dashboard-operator-trading`  
**Created**: 2026-03-30  
**Status**: Draft  
**Input**: User description: extend web UI for the Deribit trading bot with config-backed sign-in (allowlisted users, username/password, no email or verification codes), then rich trading views: health, mode, balance, performance chart, market mood, positions, strategy stats, position detail, and guided close/rebalance flows.

## Clarifications

### Session 2026-03-30

- Q: What scope should “recently closed positions” use? → A: Last **30 calendar days**, up to **200** closes shown (newest first; older or excess rows omitted from this view).
- Q: What should define the denominator for percent P/L on closes (and other % returns)? → A: Use the **backend’s authoritative** percent; the UI **always names** that basis (tooltip, column legend, or adjacent label) so operators know what the % is relative to.
- Q: What session duration / idle policy should apply? → A: Session remains valid until the operator **signs out**. **No** automatic end from **idle timeout** or **maximum calendar age** alone; server may still invalidate sessions for security or operational reasons (e.g. allowlist removal), see requirements and edge cases.
- Q: What default time range should the P/L history chart use on first load? → A: Last **30 calendar days**, aligned with the closed-positions window unless the operator changes range in a future enhancement.
- Q: How should a large **open** positions list behave? → A: **Pagination** with **25** open positions per page and clear **next/previous** (or equivalent) navigation so every open position remains reachable.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Sign in as an allowlisted operator (Priority: P1)

An operator opens the dashboard, enters username and password, and gains access only if their account appears in the server-maintained allowlist. There is no self-service registration, no email delivery, and no confirmation codes.

**Why this priority**: Access control must exist before any trading or balance information is shown.

**Independent Test**: With a known allowlisted user and correct password, the operator reaches the main dashboard; with an unknown user or wrong password, access is denied with a clear message and no sensitive data is revealed.

**Acceptance Scenarios**:

1. **Given** a username listed in the allowlist with a valid password, **When** the operator submits sign-in, **Then** they are authenticated and see the post-login dashboard.
2. **Given** a username not in the allowlist or an incorrect password, **When** the operator submits sign-in, **Then** access is denied and the response does not confirm which factor failed beyond a generic invalid-credentials message.
3. **Given** an authenticated session, **When** the operator returns later in the same browser (or device storage conditions under which the session is designed to persist), **Then** they remain signed in without re-entering credentials until they **sign out** or the server **invalidates** the session for an allowed reason.
4. **Given** the operator is signed in, **When** they choose **Sign out**, **Then** the session ends and protected dashboard data and actions are no longer available until they sign in again.

---

### User Story 2 - See system health and trading connection mode (Priority: P2)

After sign-in, the operator sees how long the bot service has been running, current memory use, and whether the bot is connected for **test** (paper/simulated) versus **live** trading.

**Why this priority**: Operators must not confuse environments or miss resource pressure before acting on positions.

**Independent Test**: Health and mode values update on refresh or on a defined refresh cycle and are visually distinct so test and live are not mistaken.

**Acceptance Scenarios**:

1. **Given** the bot is in test mode, **When** the operator views the dashboard, **Then** test mode is prominently indicated.
2. **Given** the bot is in live mode, **When** the operator views the dashboard, **Then** live mode is prominently indicated and visually differentiated from test mode.
3. **Given** the service is running, **When** the operator views the health area, **Then** they see uptime and memory usage in operator-meaningful units.

---

### User Story 3 - Portfolio overview: balance, performance, mood, and strategy summary (Priority: P3)

The operator sees account balance, a profit-and-loss chart over time (**default span: last 30 calendar days** on first load), a concise **market mood** indicator (interpretable summary of current market regime or sentiment relevant to the strategy), and strategy metadata including expected profit/loss outlook and historical or modeled win rate as defined by the strategy module.

**Why this priority**: Decisions about closing or rebalancing rely on context: performance trajectory, environment, and strategy expectations.

**Independent Test**: Each of balance, chart, mood, and strategy summary can be verified against authoritative backend figures in a controlled scenario.

**Acceptance Scenarios**:

1. **Given** connected exchange data, **When** the operator opens the overview, **Then** balance and P/L chart reflect the same figures the backend reports for that account for the **default 30-day** chart window.
2. **Given** strategy metadata is available, **When** the operator views the strategy section, **Then** expected P/L and win percentage (or equivalent) are shown with units and time horizon if applicable.
3. **Given** market mood is computed, **When** the operator views the mood indicator, **Then** they see a short label or score and optional tooltip-level explanation without requiring raw model internals.
4. **Given** historical P/L exists spanning the default window, **When** the operator first loads the overview, **Then** the chart’s time axis covers the **trailing 30 calendar days** (or all available history if shorter), labeled so the range is obvious.

---

### User Story 4 - Open and recently closed positions (Priority: P4)

The operator can reach **every** open position via a **paginated** list (**25** per page) with obvious page navigation, each row with enough context to act, including a way to start closing. They also see a list of closed positions from the **last 30 calendar days**, showing up to **200** most recent closes, each with realized P/L in USD and as a percentage **whose meaning is explicit** via an operator-visible label of the backend-defined basis.

**Why this priority**: Open positions are the primary operational surface; recent closes validate strategy behavior.

**Independent Test**: Open positions match backend state across pages; close action is only available for open positions; closed list shows correct USD and percent P/L for each closed trade, and every percent is paired with a visible basis label consistent with backend definitions.

**Acceptance Scenarios**:

1. **Given** open positions exist, **When** the operator views the positions list, **Then** at most **25** open positions appear per page, page navigation is available when there are more, and every open position can be reached by paging without omission.
2. **Given** an open position, **When** the operator chooses close, **Then** they are taken into the close flow (modal or dedicated step) before execution.
3. **Given** closes exist within the last 30 days, **When** the operator views history, **Then** each row shows USD P/L and percent P/L consistent with backend calculations, and the view includes at most 200 rows ordered newest-first.
4. **Given** a row shows percent P/L, **When** the operator views it, **Then** the basis for that percentage (as defined by the backend) is visible in the same area or via an obvious on-screen hint (e.g. column header legend or tooltip).

---

### User Story 5 - Position detail: structure, liquidity, metrics, and Greeks (Priority: P5)

From a position, the operator drills into detail: individual legs (instruments and sides), liquidity-related information, risk or execution metrics, and option Greeks where applicable.

**Why this priority**: Multi-leg and options positions need decomposition for competent risk management.

**Independent Test**: Legs and Greeks for a known test position match exchange or internal pricing outputs used elsewhere in the stack.

**Acceptance Scenarios**:

1. **Given** a multi-leg position, **When** the operator opens detail, **Then** all legs are listed with quantities and instruments.
2. **Given** the position has computed Greeks, **When** the operator views detail, **Then** Greeks are shown with definitions or labels so operators know units and meaning.
3. **Given** liquidity metrics exist, **When** the operator views detail, **Then** liquidity is expressed in terms the operator can use to judge execution risk (e.g. depth, spread proxy, or similar).

---

### User Story 6 - Guided close and rebalance (Priority: P6)

When closing, the operator sees an estimated exit P/L and plain-language guidance on **wait versus close now** (e.g. based on slippage, time decay, or policy). When rebalancing, the operator sees suggested adjustments and the projected portfolio outcomes if applied.

**Why this priority**: Reduces impulsive actions and aligns UI with operator safety goals.

**Independent Test**: Estimates and suggestions are labeled as estimates, update when inputs change, and require explicit confirmation before sending orders.

**Acceptance Scenarios**:

1. **Given** the close modal is open, **When** the operator reviews it, **Then** they see estimated exit P/L and wait-vs-close guidance before confirming.
2. **Given** the rebalance modal is open, **When** the operator reviews it, **Then** they see suggested adjustments and projected outcomes before confirming.
3. **Given** the operator cancels either modal, **When** they dismiss it, **Then** no order is placed and the previous screen state is restored.

---

### Edge Cases

- Invalid credentials, removed allowlist user, server-invalidated session, or signed-out client must not expose balances, positions, or strategy details.
- Allowlist or credential change: if an identity is no longer allowed or verifiers are rotated in configuration, the next protected request MUST fail authentication and MUST NOT leak sensitive data.
- If the trading connection or data feed is unavailable, the dashboard shows a clear degraded state per panel rather than stale numbers without warning.
- Empty portfolio: open and closed sections show helpful empty states.
- Partial leg fills or positions in transition: UI reflects backend state and does not offer close on already-closed legs.
- More than **25** open positions: the list is **paginated** (**25** per page); navigation shows that additional pages exist and does not drop positions between pages.
- More than **200** closes in **30** days: the closed-positions view shows the **200** newest; the UI indicates that the list is capped (without implying the cap is hit when it is not).
- Percent P/L shown without a supplied basis label: treat as incomplete data—do not display a misleading bare **%**; show an explicit **not available** state or USD-only until basis and percent are both supplied.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST restrict dashboard access to operators who authenticate with username and password AND whose username is present in the server-managed allowlist.
- **FR-002**: The system MUST NOT send email, SMS, or other out-of-band verification codes for this sign-in flow.
- **FR-003**: The system MUST NOT permit self-service account creation through the UI; only allowlisted identities may sign in.
- **FR-004**: Credentials MUST be verified on the server using secure password practices; raw passwords MUST NOT appear in logs or operator-visible diagnostics.
- **FR-005**: After authentication, the operator MUST see service uptime and memory usage, and the active trading connection mode as **test** or **live**, unambiguously.
- **FR-006**: After authentication, the operator MUST see account balance and a P/L history chart whose **default** span is the **last 30 calendar days** (or all available history if fewer than 30 days exist), with values alignable to backend-reported totals for that window.
- **FR-007**: After authentication, the operator MUST see a market mood summary tied to the strategy or market data the bot uses.
- **FR-008**: After authentication, the operator MUST see strategy metadata including expected P/L and win rate (or strategy-defined equivalents when exact win rate is unavailable).
- **FR-009**: The operator MUST be able to list open positions and initiate a close action that leads to confirmation. Open positions MUST be shown **paginated** at **25** per page with **next/previous** (or equivalent) controls so the operator can reach **every** open position when the count exceeds **25**.
- **FR-010**: The operator MUST be able to view closed positions from the **last 30 calendar days** with realized P/L in USD and as a percentage, showing up to **200** most recent closes (newest first).
- **FR-011**: Wherever percentage P/L or percentage return is shown for positions or closes, the value MUST match the backend’s authoritative calculation, and the UI MUST make the **basis** of that percentage visible (label, legend, or tooltip naming what the % is relative to, as defined by the backend). If the backend does not supply both the value and its basis, the UI MUST NOT imply a precise % return (see edge cases).
- **FR-012**: The operator MUST be able to open position detail showing legs, liquidity context, metrics, and Greeks when the position type supports them.
- **FR-013**: The close flow MUST present an estimated exit P/L and wait-versus-close guidance prior to order submission.
- **FR-014**: The rebalance flow MUST present suggested adjustments and projected outcomes prior to order submission.
- **FR-015**: Destructive or capital-moving actions (close, rebalance execution) MUST require explicit operator confirmation.
- **FR-016**: The system MUST enforce session boundaries so unauthenticated clients cannot read protected dashboard data or trigger protected actions.
- **FR-017**: Authenticated sessions MUST remain valid until the operator **signs out** or the server **invalidates** the session for an auditable reason (e.g. identity no longer allowlisted, credential verifier rotation, explicit revocation). The system MUST NOT end sessions based solely on **idle timeout** or **fixed maximum session age**.
- **FR-018**: The authenticated dashboard MUST expose a clear **Sign out** (or **Log out**) action that ends the session server-side and leaves the client unable to access protected data until the next successful sign-in.

### Key Entities *(include if feature involves data)*

- **Allowlisted operator**: Username, password verifier reference, optional display metadata; maintained in server configuration or equivalent admin-controlled store.
- **Session**: Authenticated operator identity bound to the client after successful sign-in; persists until **sign-out** or **server invalidation** (not idle-only or max-age-only expiry).
- **System health snapshot**: Uptime, memory usage, timestamp of measurement.
- **Trading connection profile**: Mode (test vs live), connectivity status.
- **Account snapshot**: Balance, equity or margin context as relevant to the bot.
- **P/L series**: Time-ordered points for charting performance; default dashboard request covers **30** trailing calendar days unless a later release adds range controls.
- **Market mood**: Compact indicator plus human-readable explanation text.
- **Strategy metadata**: Expected P/L, win rate or substitute metrics, definitions of time horizon.
- **Position**: Open or closed state, identifiers, legs, summaries for list and detail views.
- **Open-position list (UI)**: Paginated view, **25** rows per page, stable ordering consistent with the backend for the active sort (default sort defined at implementation time if not otherwise specified).
- **Closed-position history (UI scope)**: Rows for exits in the trailing **30** days, surfaced as up to **200** newest records for the dashboard list.
- **Percent return display**: Authoritative percent from the backend plus an operator-visible **basis** descriptor so the denominator is never ambiguous.
- **Leg**: Instrument, side, quantity, and pricing or mark fields as applicable.
- **Greeks set**: Delta, gamma, vega, theta (and others if used) with units.
- **Close estimate**: Estimated exit P/L, assumptions, and wait-vs-close recommendation text.
- **Rebalance proposal**: Suggested trades or deltas, projected portfolio outcome summary.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: At least **95%** of sign-in attempts by allowlisted operators with correct credentials complete successfully in under **60 seconds** under normal network conditions.
- **SC-002**: **100%** of authenticated dashboard views expose test vs live mode in a single screen area without navigation deeper than the main overview.
- **SC-003**: For a sample set of **10** benchmark positions, open/close lists and detail fields match backend-authoritative values with **no mismatches** in P/L, leg count, or mode flags during acceptance testing.
- **SC-004**: At least **90%** of surveyed operators (or internal reviewers standing in for operators) correctly identify whether they are in test or live mode when shown a screenshot of the overview within **5 seconds**.
- **SC-005**: Every close and rebalance submission path includes a confirmation step that surfaces estimated or projected outcome text **100%** of the time before orders can be sent.
- **SC-006**: In acceptance review of the closed-positions list and open-position summaries, **100%** of displayed percentage P/L fields have a visible basis label that matches the backend’s stated definition for those rows.
- **SC-007**: In scripted acceptance, an operator who remains signed in over a **2-hour** idle gap (no dashboard interaction) still has full protected access **without** re-entering password, and after **Sign out**, a protected request **100%** of the time fails until sign-in succeeds again.
- **SC-008**: In acceptance, the default P/L chart shown on first overview load uses a **30-calendar-day** trailing window (or shorter when history is insufficient) and matches backend series for that span with **no** date-span mismatches.
- **SC-009**: In acceptance with more than **25** mock open positions, paging reveals **100%** of rows exactly once across pages (**25** per page), with working next/previous (or equivalent), and totals match the backend open-position count.

## Assumptions

- Operators are a small, trusted set; allowlist changes are made by administrators through configuration deployment, not through the dashboard UI in this release.
- Operators expect to stay signed in across long work sessions; sessions end on **explicit sign-out** or **server invalidation**, not automatic idle or calendar limits.
- “Market mood” is produced by existing or planned bot analytics (volatility regime, skew, or composite score) and is not a third-party social sentiment feed unless later specified.
- P/L chart **defaults to 30 calendar days** to match the closed-positions window; **interactive range switching** (7D / 90D / custom) is optional follow-up unless added in the same release.
- Greek and liquidity values depend on live or recent market data; when unavailable, the UI shows explicit “not available” rather than zeros that imply certainty.
- Mobile layout optimization is desirable but not required for initial acceptance if desktop-first operators are the primary audience.
- **Open-position** list ordering (e.g. by instrument, opened time, or P/L) follows a single backend-defined default unless a later release adds sort controls; pagination semantics remain **25** per page regardless.
