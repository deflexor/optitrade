# Work Packages: Operator Trading Dashboard

**Inputs**: `kitty-specs/002-operator-trading-dashboard/` (spec.md, plan.md, data-model.md, research.md, contracts/dashboard-api.openapi.yaml, quickstart.md)  
**Prerequisites**: plan.md, spec.md  
**Tests**: Constitution requires automated tests on **auth critical paths** (allowlist messaging, session denial, password hashing); snapshot **staleness (5 s)** with injectable clock; prefer `httptest` for handlers.

**Organization**: Subtasks `T001`..`T047` roll into work packages `WP01`..`WP09`. Prompts live in `kitty-specs/002-operator-trading-dashboard/tasks/` (flat directory).

**Path conventions** (from plan): `web/` (Vite React SPA), `src/internal/dashboard/` (Go BFF), `src/cmd/optitrade`, existing `src/internal/state/sqlite` patterns.

---

## Dependency and execution summary

- **Sequence**: WP01 -> WP02 -> WP03 -> **WP04 and WP05 in parallel** -> WP06 -> **WP07 and WP08 in parallel** -> WP09.
- **MVP slice**: WP01 + WP02 + WP03 + WP04 (snapshot + chart stub) + WP05 (list stub) delivers signed-in health view; full operator value needs WP06 to WP09.
- **Implement commands**:
  - `spec-kitty implement WP01`
  - `spec-kitty implement WP02 --base WP01`
  - `spec-kitty implement WP03 --base WP02`
  - `spec-kitty implement WP04 --base WP03`
  - `spec-kitty implement WP05 --base WP03`
  - `spec-kitty implement WP06 --base WP05`
  - `spec-kitty implement WP07 --base WP06`
  - `spec-kitty implement WP08 --base WP06`
  - `spec-kitty implement WP09 --base WP08` (branch already includes WP07 if merged serially; use merge order from your worktrees)

---

## Work Package WP01: Web and Go dashboard scaffold (Priority: P0)

**Goal**: Create `web/` (Vite, React, TypeScript, Tailwind, Zustand, Axios) and Go `src/internal/dashboard/` skeleton with HTTP mux entry wired from `cmd/optitrade` behind a config flag.  
**Independent Test**: `npm run build` in `web/` succeeds; `go build -o /dev/null ./src/cmd/optitrade` with dashboard packages compiling (handlers may return 501).  
**Prompt**: [tasks/WP01-web-go-scaffold.md](tasks/WP01-web-go-scaffold.md)  
**Estimated prompt size**: ~300 lines

**Requirements Refs**: FR-017

### Included Subtasks
- [x] T001 Create `web/` with Vite, React 18, TS, Tailwind, `index.html`, `src/main.tsx`, `src/App.tsx`
- [x] T002 Add Zustand, React Router, Axios; `src/api/client.ts` with `withCredentials: true` and `/api/v1` baseURL from `import.meta.env`
- [x] T003 [P] Configure `vite.config.ts` `server.proxy` for `/api` -> `http://127.0.0.1:8080` (port documented)
- [x] T004 Create `src/internal/dashboard/` with `Server` struct, route registration stub, `go` doc pointing to OpenAPI
- [x] T005 Wire dashboard listen address into `src/cmd/optitrade/main.go` via flag or env (e.g. `OPTITRADE_DASHBOARD_LISTEN`); no-op when empty
- [x] T006 [P] Add minimal `embed` placeholder or `//go:embed` stub for future `web/dist` (can be empty FS until WP09)

### Implementation Notes
Follow `.kittify/AGENTS.md` UTF-8 rules. Do not commit `node_modules`.

### Dependencies
- None

### Risks and mitigations
Port clash: document in `quickstart.md` updates in WP09.

---

## Work Package WP02: Dashboard SQLite auth and handlers (Priority: P0)

**Goal**: Migrations for `dashboard_user` / `dashboard_session`, password hashing, allowlist-gated login per FR-014 to FR-018 and SC-006.  
**Independent Test**: `go test ./src/internal/dashboard/...` covers register, allowlisted login session cookie, non-allowlisted returns exact JSON `Sorry, feature not ready`, wrong password 401 without not-ready text.  
**Prompt**: [tasks/WP02-dashboard-auth-sqlite.md](tasks/WP02-dashboard-auth-sqlite.md)  
**Estimated prompt size**: ~380 lines

**Requirements Refs**: FR-014, FR-015, FR-016, FR-018, SC-006

### Included Subtasks
- [x] T007 Add SQLite migrations for `dashboard_user` and `dashboard_session` per `data-model.md`; wire into existing migration runner
- [x] T008 Implement password set/verify with Argon2id or bcrypt; document cost params; never log passwords
- [x] T009 Implement `Register` (username normalize, unique constraint, reject empty weak patterns if minimal validation added)
- [x] T010 Load allowlist from env `OPTITRADE_DASHBOARD_ALLOWLIST` (comma-separated); document in plan/quickstart
- [x] T011 Session issue: random token, store hash only, httpOnly+Secure+SameSite=Lax cookie `optitrade_session`; expiry
- [x] T012 HTTP: `POST /api/v1/auth/register`, `POST /api/v1/auth/login`, `POST /api/v1/auth/logout` matching OpenAPI error shapes
- [x] T013 [P] Table-driven `httptest` tests for full auth matrix and rate-limit stub hook on login

### Implementation Notes
Constitution: parameterized SQL only; redact secrets in logs.

### Dependencies
- Depends on WP01

### Risks and mitigations
Timing side channels: use constant-time compare for passwords; session timing same as spec for not-ready path.

---

## Work Package WP03: Auth middleware and summary snapshot API (Priority: P1)

**Goal**: Protect `/api/v1/*` (except auth); implement `GET /api/v1/snapshot` with uptime, RSS, Deribit mode, equity, paired regime/mood, `stale` per FR-019.  
**Independent Test**: Authenticated `GET /snapshot` returns 200 JSON matching OpenAPI; halted updater makes `stale=true` within 5 s rule in unit test with fake clock.  
**Prompt**: [tasks/WP03-snapshot-middleware.md](tasks/WP03-snapshot-middleware.md)  
**Estimated prompt size**: ~400 lines

**Requirements Refs**: FR-001, FR-002, FR-003, FR-005, FR-017, FR-019, SC-001, SC-002, SC-009, SC-008

### Included Subtasks
- [x] T014 Auth middleware: validate session; `401` on missing/invalid; attach `user_id` to context
- [x] T015 Implement snapshot assembler reading process stats (`runtime`, `memstats` or psutil-equivalent in Go), Deribit test/live from existing client config
- [x] T016 Read equity from existing account summary path or temporary stub with `TODO` bounded behind feature flag
- [x] T017 Read canonical regime from `internal/regime` or latest `regime_state`; set `regime_label` and `market_mood_label` identically
- [x] T018 Compute `stale`: true when `now - snapshot_utc > 5s` or subsystem freshness flags; document clock source
- [x] T019 Expose `GET /api/v1/snapshot`; return RFC3339 `snapshot_utc`
- [x] T020 [P] Tests: staleness threshold, paired labels equality, middleware 401

### Implementation Notes
Integrate without blocking trading hot path: snapshot build should read caches; if data missing return explicit `stale` or degraded fields per spec edge cases.

### Dependencies
- Depends on WP02

### Risks and mitigations
Deribit round-trip too slow: use last cached account summary with age in DTO.

---

## Work Package WP04: P/L series API and chart UI (Priority: P2)

**Goal**: `GET /api/v1/pl-series` default 30d; React chart with range control and short-history messaging (FR-004, SC-007).  
**Independent Test**: API returns monotonic timestamps for window; UI shows 30d default label; with <30d data backend indicates span in JSON.  
**Prompt**: [tasks/WP04-pl-series-chart.md](tasks/WP04-pl-series-chart.md)  
**Estimated prompt size**: ~320 lines

**Requirements Refs**: FR-004, SC-007

### Implementation Subtasks
- [x] T021 Define P/L storage or query: new `dashboard_pnl_snapshot` table or reuse bot aggregates; seed minimal migration
- [x] T022 Implement `GET /api/v1/pl-series?range=` with `30d` default; validate enum per OpenAPI
- [x] T023 React: chart component (reuse lightweight lib or SVG) and Zustand fetch on interval aligned with dashboard poll
- [x] T024 UX: show active range; when history < 30d display helper text from API metadata

### Implementation Notes
If bot has no history yet, return empty `points` plus `range_actual` metadata.

### Dependencies
- Depends on WP03

### Parallel opportunities
- [P] T023 UI can start against mock JSON before T021 is complete (fixture file).

### Risks and mitigations
Precision: decimals as strings end-to-end.

---

## Work Package WP05: Open and closed positions list API and UI (Priority: P2)

**Goal**: List endpoints and dashboard tables for open positions and 10 closed (FR-006, FR-007, FR-008).  
**Independent Test**: Lists match SQLite/exchange reconciliation source in dev; sort newest-first closed; row shows strategy, expected P/L, win rate, closed P/L USD and %.  
**Prompt**: [tasks/WP05-positions-lists.md](tasks/WP05-positions-lists.md)  
**Estimated prompt size**: ~340 lines

**Requirements Refs**: FR-006, FR-007, FR-008

### Included Subtasks
- [x] T025 `GET /api/v1/positions/open` mapping from internal position models to OpenAPI `PositionSummaryOpen`
- [x] T026 `GET /api/v1/positions/closed?limit=10` capped at 10, newest first
- [x] T027 Strategy stats: wire win rate and expected P/L fields from backend calculators or placeholders with clear `TODO` and feature flag
- [x] T028 React list views with loading and error states; **Close** button navigates or opens flow WP07
- [x] T029 Empty and partial-history states per spec edge cases

### Implementation Notes
Reuse `internal/risk` / execution reconciliation types where possible.

### Dependencies
- Depends on WP03

### Risks and mitigations
Multi-leg partial fills: list row shows aggregate state; detail in WP06.

---

## Work Package WP06: Position detail, legs, and Greeks (Priority: P3)

**Goal**: `GET /api/v1/positions/{id}` with legs, liquidity notes, greeks (FR-009).  
**Independent Test**: Detail view renders all legs for a multi-leg test fixture; unknown greeks show null-safe UI.  
**Prompt**: [tasks/WP06-position-detail.md](tasks/WP06-position-detail.md)  
**Estimated prompt size**: ~330 lines

**Requirements Refs**: FR-009

### Included Subtasks
- [x] T030 Implement position detail handler with stable `position_id` scheme matching list endpoints
- [x] T031 Map each leg: instrument, side, size, liquidity string from book cache or stub
- [x] T032 Populate leg and summary greeks when available from risk engine or exchange
- [x] T033 React route `/positions/:id` with sections for legs and metrics
- [x] T034 Navigation from list rows; deep-link refresh

### Implementation Notes
502/503 when exchange disconnected: show degraded message, no fake numbers.

### Dependencies
- Depends on WP05

### Risks and mitigations
Large books: cap leg list display with expand if ever needed (YAGNI for BTC/ETH defined-risk).

---

## Work Package WP07: Close preview and close execution (Priority: P3)

**Goal**: Close modals with estimated P/L, quote time, recommendation, confirm; `POST` close and preview (FR-010, FR-012, FR-013, SC-005).  
**Independent Test**: Preview returns `quote_as_of`; stale quote blocks confirm button; successful close updates list after refresh.  
**Prompt**: [tasks/WP07-close-flow.md](tasks/WP07-close-flow.md)  
**Estimated prompt size**: ~360 lines

**Requirements Refs**: FR-010, FR-012, FR-013, SC-005, SC-003

### Included Subtasks
- [x] T035 `POST /api/v1/positions/{id}/close-preview` calling execution layer or simulator; include `recommendation` enum
- [x] T036 `POST /api/v1/positions/{id}/close` with `confirm` body; idempotent guard; return 409 on conflict
- [x] T037 React modal: show P/L, recommendation, rationale; disable confirm if preview stale vs FR-019 freshness rules for quotes
- [x] T038 Error display and list reconciliation after mutation

### Implementation Notes
Destructive path must mirror spec: no silent success.

### Dependencies
- Depends on WP06

### Risks and mitigations
Exchange reject: surface message; audit log if available.

---

## Work Package WP08: Rebalance preview and execution (Priority: P4)

**Goal**: Rebalance suggestions modal and endpoints (FR-011, FR-012).  
**Independent Test**: Preview returns ranked suggestions; execute returns 202; failures explicit.  
**Prompt**: [tasks/WP08-rebalance-flow.md](tasks/WP08-rebalance-flow.md)  
**Estimated prompt size**: ~320 lines

**Requirements Refs**: FR-011, FR-012, FR-013

### Included Subtasks
- [x] T039 `POST /api/v1/positions/{id}/rebalance-preview` returning ranked items
- [x] T040 `POST /api/v1/positions/{id}/rebalance` with validation and async-friendly response
- [x] T041 React modal listing suggestions with expected_effect text
- [x] T042 Confirm step and post-condition UI refresh

### Implementation Notes
If engine lacks rebalance v1, return empty suggestions with clear copy (still satisfies structural API; wire real logic when bot supports).

### Dependencies
- Depends on WP06

### Risks and mitigations
Over-promising outcomes: copy as "estimated" per spec.

---

## Work Package WP09: SPA embed, build pipeline, and hardening (Priority: P2)

**Goal**: Production `go:embed web/dist`, same-origin static+API, update `quickstart.md`, login/register pages polish, CSRF notes, UTF-8 validation.  
**Independent Test**: Single binary serves UI and API; smoke: register, allowlist login, snapshot poll; `spec-kitty validate-encoding --feature 002-operator-trading-dashboard`.  
**Prompt**: [tasks/WP09-embed-and-hardening.md](tasks/WP09-embed-and-hardening.md)  
**Estimated prompt size**: ~350 lines

**Requirements Refs**: FR-017, SC-001, SC-003, SC-004, constitution UX and security

### Included Subtasks
- [ ] T043 `//go:embed all:dist` (or build tag) bundle `web/dist` into binary; fallback 404 message if embed empty in dev
- [ ] T044 Root Makefile or CI step: `npm ci && npm run build --prefix web` before `go build` for release
- [ ] T045 React auth pages: register, login, route guard for protected layout; **User Story 1** acceptance scenarios 4-7
- [ ] T046 Document CSRF strategy (SameSite, POST-only mutations); security self-review checklist
- [ ] T047 Run encoding validator; align copy with exact `Sorry, feature not ready` string

### Implementation Notes
Cross-link repo `docs/quickstart.md` only if product owner wants duplication; prefer feature `quickstart.md`.

### Dependencies
- Depends on WP08 (and transitively full UI/API). **Also requires WP04 and WP07 outputs present before release candidate.**

### Risks and mitigations
CORS: same-origin eliminates most issues; document dev proxy.

---

## Subtask index (reference)

| Subtask | Summary | WP |
|---------|---------|-----|
| T001-T006 | Scaffold web+Go | WP01 |
| T007-T013 | Auth SQL + handlers | WP02 |
| T014-T020 | Middleware + snapshot | WP03 |
| T021-T024 | P/L API + chart | WP04 |
| T025-T029 | Position lists | WP05 |
| T030-T034 | Position detail | WP06 |
| T035-T038 | Close flow | WP07 |
| T039-T042 | Rebalance flow | WP08 |
| T043-T047 | Embed + hardening | WP09 |
