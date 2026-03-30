# Feature Specification: Dashboard robustness and operator clarity

**Feature Branch**: `003-dashboard-e2e-health`  
**Created**: 2026-03-31  
**Status**: Draft  
**Input**: User description: expand automated end-to-end verification for the operator dashboard; eliminate confusing behavior when viewing open and closed positions (including failed data fetches); ensure every dashboard element defined in the operator trading dashboard specification is present and either shows correct information or a clear explanation when something is unavailable; present service uptime and memory use in human-friendly units; replace placeholder “market mood” messaging with a real product behavior.

**Depends on**: Operator dashboard trading specification (`002-dashboard-operator-trading`) for the authoritative list of screens, panels, and behaviors the dashboard must support.

## Clarifications

### Session 2026-03-31

- Q: When may this feature be treated as fully done? → A: Only after **every** implementation task for this feature is marked complete, **and** the project’s **full** automated test suite has been run once on that state with **all** tests passing (no failing tests; tests that must not be skipped for release remain unskipped).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Trustworthy positions views (Priority: P1)

An authenticated operator opens the positions experience (open and recently closed lists). They see accurate lists when the system can provide them, or a clear in-product explanation when data cannot be loaded (for example temporary backend or trading-link unavailability). They do not encounter a “working” screen that hides failures or dumps low-level failures only into technical tooling.

**Why this priority**: Positions are the primary operational surface; silent or confusing failure erodes trust and can delay risk decisions.

**Independent Test**: With data available, list content matches authoritative records; with data intentionally unavailable, the operator sees an explicit panel-level state and no implied success.

**Acceptance Scenarios**:

1. **Given** open-position data is available, **When** the operator views open positions, **Then** the list and pagination behave per the operator dashboard specification (including page size and navigation).
2. **Given** closed-position data is available for the defined history window, **When** the operator views closed positions, **Then** rows show the required amounts and percentage basis labeling per that specification.
3. **Given** the server cannot supply position data temporarily, **When** the operator opens the positions views, **Then** each affected area shows a clear unavailable or retry-oriented message consistent with the dashboard’s degraded-state rules, and the experience does not present empty success where data failed.

---

### User Story 2 - Readable health at a glance (Priority: P2)

On the main dashboard, the operator sees how long the service has been running and how much memory it is using, expressed in everyday units (for example duration broken into days and hours where helpful, and memory in common capacity units rather than raw low-level numbers alone).

**Why this priority**: Operators use health to spot instability before trading; opaque numbers increase misreads and support burden.

**Independent Test**: Values shown can be mapped to the underlying measurements in acceptance tooling without the operator needing a calculator.

**Acceptance Scenarios**:

1. **Given** the service reports uptime and memory use, **When** the operator views the health area, **Then** uptime is readable as a duration and memory use is readable as a capacity (not only as an unlabeled large integer).
2. **Given** health data is missing, **When** the operator views the health area, **Then** the panel states that health is unavailable rather than showing misleading placeholders.

---

### User Story 3 - Meaningful market mood (Priority: P3)

The operator sees a market mood summary that either reflects real strategy or market analytics, or a deliberate “not available” state with a short, operator-appropriate reason—not an internal “not wired yet” placeholder.

**Why this priority**: Mood is part of the overview story in the operator specification; placeholder text signals an unfinished product and blocks confident interpretation.

**Independent Test**: In a controlled environment, when analytics are configured to return a mood, the UI matches; when not, the UI shows an explicit unavailable state without developer jargon.

**Acceptance Scenarios**:

1. **Given** mood analytics are available from the product, **When** the operator views the overview, **Then** they see a concise label and optional short explanation appropriate to operators (per dashboard specification).
2. **Given** mood analytics are not available, **When** the operator views the overview, **Then** they see an explicit unavailable or pending-integration message written for operators, not implementation-side placeholder wording.

---

### User Story 4 - Dashboard specification conformance is regression-tested (Priority: P4)

The team can run automated end-to-end verification that exercises the dashboard the way an operator would and proves the product still satisfies the visible and behavioral intent of the operator trading dashboard specification: required screens and panels exist, critical interactions still work, and failure paths show the required messaging.

**Why this priority**: Prevents regressions as the stack evolves and backs up manual acceptance with repeatable checks.

**Independent Test**: A standard end-to-end run exercises sign-in (where applicable), overview, positions, and other primary routes and confirms presence of key elements and acceptable empty or error states.

**Acceptance Scenarios**:

1. **Given** a maintained list of journeys derived from the operator dashboard specification, **When** the automation suite runs, **Then** each journey completes with expected elements visible and no false “all good” when backend dependencies are down in test doubles.
2. **Given** a release candidate build, **When** stakeholders review results, **Then** they can tell from the run outcome whether the dashboard still meets the specification’s UI coverage bar.
3. **Given** every implementation task for this feature is marked done, **When** maintainers run the repository’s full automated test suite once on that state, **Then** the feature is accepted only if **all** tests pass.

---

### Edge Cases

- Operator signed out or session invalidated while on positions: protected data and actions remain blocked per existing session rules.
- Partial failure: one panel’s data unavailable while others succeed; each panel fails independently with clear messaging.
- Very large open-position sets: pagination and navigation remain correct (per existing specification).
- Closed history cap and percent basis rules: unchanged from operator dashboard specification; this feature must not weaken those rules.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The product MUST ensure open and closed position experiences either display data consistent with the operator dashboard specification or show a clear degraded-state message when loading fails, without implying success.
- **FR-002**: The product MUST avoid exposing implementation-only or developer-placeholder text to operators for market mood; mood MUST either reflect supplied analytics or a deliberate operator-facing unavailable explanation.
- **FR-003**: The authenticated overview MUST present service uptime and memory usage in human-meaningful forms (durations and capacity-style units), or an explicit unavailable state if values cannot be shown.
- **FR-004**: The product MUST verify, through automated end-to-end checks maintained with the codebase, that all dashboard elements relevant to the operator dashboard specification are present on their routes and either show required information or an acceptable error or empty state as that specification requires.
- **FR-005**: Automated journeys MUST cover the positions routes and health and mood areas at minimum, including scenarios where dependencies are unavailable in test, so that failure presentation—not silent success—is asserted.
- **FR-006**: Where the operator dashboard specification defines labels, caps, pagination, confirmation steps, and basis labeling for percentages, the conformance tests MUST include checks that those behaviors remain observable (or explicitly deferred only if the specification is formally updated).
- **FR-007**: Final acceptance of this feature MUST NOT be declared until every planned implementation task is marked done **and** the repository’s **full** automated test suite has been executed **once after** that point with **all** included tests passing.

### Key Entities *(include if feature involves data)*

- **Health presentation**: Uptime duration, memory footprint, measurement time; formatting is part of the operator-visible contract.
- **Positions presentation**: Open list, closed list, loading and error states; must align with operator dashboard specification windows and caps.
- **Market mood presentation**: Indicator label, optional explanation, unavailable state; must not use internal placeholder copy.
- **Conformance journey set**: Named operator paths and assertions used for regression (derived from the operator dashboard specification).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: In acceptance, **100%** of automated conformance journeys derived from the operator dashboard specification’s primary flows pass on the release candidate, or failures are documented as specification amendments with stakeholder agreement.
- **SC-002**: In supervised review, **100%** of sampled visits to positions views under simulated backend unavailability show an explicit panel-level problem state, not an empty list presented as if no positions exist.
- **SC-003**: In supervised review, **100%** of dashboard screenshots taken after sign-in show uptime and memory in human-meaningful units or an explicit “health unavailable” style message—never only raw opaque numbers with no unit context.
- **SC-004**: In supervised review, **zero** operator-visible instances of internal placeholder mood wording remain on production-oriented builds; mood is either analytics-backed or clearly labeled as unavailable for operators.
- **SC-005**: Repeatable automated verification adds or extends coverage so that at least **five** distinct primary operator journeys (or equivalent breadth agreed from the operator specification) are exercised each release cycle, including positions and overview health.
- **SC-006**: Feature completion is evidenced by **100%** pass rate on the repository’s **full** automated test suite in a single run started **after** all tasks for this feature are marked complete; **zero** failing tests are allowed for that completion claim.

## Assumptions

- The operator dashboard trading specification remains the source of truth for which UI elements and behaviors must exist unless explicitly revised.
- Test environments can simulate unavailable trading or data dependencies without affecting production.
- “Human-meaningful” units follow common expectations (e.g. combined day/hour uptime, MB/GB-style memory) and may be rounded for display while remaining traceable to underlying values in acceptance checks.
- Session, authentication, and destructive-action confirmation rules from the prior specification are unchanged and are included only where conformance tests exercise those paths.
- “Full automated test suite” means the union of all routine automated test targets the project expects before merge or release (as documented in the implementation plan or project conventions), not an ad-hoc subset; running tests during development remains encouraged—**FR-007** / **SC-006** add a **final** gate after work is marked done.
