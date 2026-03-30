---
description: "Task list for operator dashboard trading and controls (002-dashboard-operator-trading)"
---

# Tasks: Operator dashboard trading and controls

**Input**: Design documents from `/specs/002-dashboard-operator-trading/`  
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/openapi.yaml, quickstart.md

**Tests**: Included where the implementation plan requires them — handler tests for **auth** and **one trading read path** (`plan.md`), with optional extension per story.

**Organization**: Phases follow user stories P1–P6 from `spec.md` after shared setup and foundation.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Parallel-safe (different files, no ordering dependency on incomplete tasks in the same phase)
- **[Story]**: `[US1]`…`[US6]` for user-story phases only
- Paths follow `src/` (Go) and `web/` (React) per `plan.md`

## Phase 1: Setup (shared infrastructure)

**Purpose**: Confirm brownfield dashboard dev loop and wiring before feature work.

- [X] T001 Confirm `make run-dashboard`, `make web-dev`, and embed path `src/internal/dashboard/dist` work per `Makefile` and `README.md`
- [X] T002 [P] Verify Vite dev server proxies `/api/v1` to the Go dashboard listener in `web/vite.config.ts` (or equivalent)
- [X] T003 [P] Thread `OPTITRADE_DASHBOARD_AUTH_PATH` (and listen addr flags) from env/CLI into dashboard construction in `src/cmd/optitrade/` dashboard command

---

## Phase 2: Foundational (blocking prerequisites)

**Purpose**: Config-backed allowlist, persistent sessions without idle/max-age auth gates, and API router shape **before** any user story is complete end-to-end.

**⚠️ CRITICAL**: User story UI+API acceptance paths must not ship protected trading data until this phase’s auth boundary exists.

- [X] T004 Revise or add SQLite migration for session-focused schema (no auth identity FK to removed `dashboard_user` pattern) in `src/internal/state/migrations/` per `research.md` R-01/R-02 and plan Complexity Tracking
- [X] T005 Load and validate `DashboardAuthFile` JSON (version, users with `password_hash`) with reload on `SIGHUP` or documented restart in `src/internal/dashboard/` or `src/internal/config/`
- [X] T006 [P] Implement password verification (bcrypt or argon2id) without logging secrets in `src/internal/dashboard/`
- [X] T007 Implement session lifecycle (opaque token, hash-at-rest, create on login, delete on logout) in `src/internal/dashboard/session_store.go` (or adjacent package)
- [X] T008 Implement auth middleware: HTTP-only `optitrade_session` cookie, re-check username ∈ loaded allowlist, **no** 401 solely from idle or max-age clock in `src/internal/dashboard/middleware.go`
- [X] T009 Replace stub `/api/v1/` 404 JSON with a real sub-router: public `POST /auth/login`; protect all other `/api/v1` routes per `contracts/openapi.yaml` in `src/internal/dashboard/server.go`
- [X] T010 Add shared helper for Deribit/BFF calls using `context` with bounded timeouts consistent with `src/internal/deribit/` RPC patterns in `src/internal/dashboard/`
- [X] T011 [P] Centralize JSON error envelopes and redacted `slog` usage for dashboard handlers in `src/internal/dashboard/` (extend `writeJSON` pattern)

**Checkpoint**: Unauthenticated clients receive 401 on protected routes; login route exists but may return errors until US1 finishes handler bodies.

---

## Phase 3: User Story 1 — Sign in as allowlisted operator (Priority: P1) 🎯 MVP

**Goal**: Allowlisted username/password login, persistent session until sign-out or server invalidation, logout, and SPA shell gating.

**Independent Test**: Valid allowlisted credentials reach post-login shell; invalid credentials get generic failure; session survives idle; sign-out clears access.

### Tests (plan-required)

- [X] T012 [P] [US1] Handler tests for `POST /auth/login`, `POST /auth/logout`, `GET /auth/me` (cookie + 401/204 behavior) in `src/internal/dashboard/auth_handlers_test.go`

### Implementation

- [X] T013 [US1] Implement `POST /auth/login` per `contracts/openapi.yaml` in `src/internal/dashboard/` (set session cookie, generic 401 on failure)
- [X] T014 [US1] Implement `POST /auth/logout` and `GET /auth/me` in `src/internal/dashboard/`
- [X] T015 [P] [US1] Add login page UI (no sensitive leakage in errors) in `web/src/pages/` and route in `web/src/App.tsx`
- [X] T016 [US1] Wire Zustand `web/src/stores/authStore.ts` and Axios `web/src/api/client.ts` for login/logout/me with `withCredentials`
- [X] T017 [US1] Add protected layout (redirect unauthenticated users to login) in `web/src/App.tsx`

**Checkpoint**: US1 acceptance scenarios from `spec.md` hold without any trading panels populated.

---

## Phase 4: User Story 2 — Health and trading mode (Priority: P2)

**Goal**: After sign-in, operator sees process uptime, memory use, and **test** vs **live** mode unambiguously.

**Independent Test**: Values refresh on reload; test/live visually distinct; numbers match `runtime.MemStats` + uptime source of truth from BFF.

### Implementation

- [X] T018 [US2] Implement `GET /api/v1/health` returning health snapshot + trading connection profile DTOs per `data-model.md` and `contracts/openapi.yaml` in `src/internal/dashboard/health.go`
- [X] T019 [P] [US2] Add health/mode panel to main dashboard/overview shell in `web/src/` (Tailwind, prominent test/live styling)

**Checkpoint**: FR-005 satisfied in UI backed by live JSON.

---

## Phase 5: User Story 3 — Portfolio overview (Priority: P3)

**Goal**: Balance, default **30 calendar day** P/L series, market mood, strategy metadata with honest degraded states.

**Independent Test**: Balance and P/L series match BFF for the window; mood/strategy panels show degraded messaging when backend unavailable (no fabricated values).

### Tests (plan-required trading read path)

- [X] T020 [US3] Handler test for `GET /api/v1/overview` JSON shape (decimal strings, mood `available`, strategy fields) in `src/internal/dashboard/overview_test.go`

### Implementation

- [X] T021 [US3] Implement `GET /api/v1/overview`: account snapshot + P/L points (`from`/`to` window) in `src/internal/dashboard/overview.go`
- [X] T022 [P] [US3] Integrate Deribit/account reads for balance and series using T010 helper in `src/internal/dashboard/`
- [X] T023 [P] [US3] Add market mood + strategy metadata aggregation with `available: false` fallback per `research.md` R-09 in `src/internal/dashboard/overview.go`
- [X] T024 [P] [US3] Build overview UI: balance, 30d chart, mood, strategy sections with empty/error states in `web/src/`

**Checkpoint**: FR-006–FR-008 and SC-008 satisfied for overview.

---

## Phase 6: User Story 4 — Open and closed positions lists (Priority: P4)

**Goal**: Paginated open positions (25/page, cursor/page nav), closed positions (30 days, max 200, newest first), USD + **labeled** percent basis.

**Independent Test**: Backend totals match across pages; closed rows always pair `%` with `percent_basis_label`; cap messaging when 200 rows hit.

### Implementation

- [X] T025 [US4] Implement `GET /api/v1/positions/open` with `limit=25` and cursor pagination + `total_count` per `research.md` R-05 in `src/internal/dashboard/positions.go`
- [X] T026 [US4] Implement `GET /api/v1/positions/closed` (30d filter, max 200) with `realized_pnl_pct` + `percent_basis_label` per R-06 in `src/internal/dashboard/positions.go`
- [X] T027 [P] [US4] Positions list UI: pagination controls, closed table, basis labels/tooltips, empty states in `web/src/pages/` (or `web/src/components/`)
- [X] T028 [US4] Gate “start close” entry only on open rows; honor degraded feed messaging per `spec.md` edge cases in `web/src/`

**Checkpoint**: FR-009–FR-011 and SC-006/SC-009 for list behaviors.

---

## Phase 7: User Story 5 — Position detail (Priority: P5)

**Goal**: Drill-in shows legs, liquidity context, metrics, and Greeks with explicit N/A when data missing.

**Independent Test**: Multi-leg position lists all legs; Greeks labeled; liquidity fields match BFF.

### Implementation

- [X] T029 [US5] Implement `GET /api/v1/positions/{id}` mapping legs, metrics, `greeks` in `src/internal/dashboard/positions.go`
- [X] T030 [P] [US5] Add position detail route/view and navigation from list rows in `web/src/`

**Checkpoint**: FR-012 and related acceptance scenarios for detail.

---

## Phase 8: User Story 6 — Guided close and rebalance (Priority: P6)

**Goal**: Preview (estimates/guidance) then confirm for close and rebalance; cancel leaves state unchanged; preview token or field echo prevents stale double-submit.

**Independent Test**: Cannot confirm without preview payload visible; cancel sends no orders; estimates labeled as estimates.

### Implementation

- [X] T031 [US6] Implement `POST /api/v1/positions/{id}/close/preview` and `.../close/confirm` with preview token/echo pattern in `src/internal/dashboard/close.go`
- [X] T032 [US6] Implement `POST /api/v1/rebalance/preview` and `POST /api/v1/rebalance/confirm` in `src/internal/dashboard/rebalance.go`
- [X] T033 [P] [US6] Close modal: preview → confirm/cancel in `web/src/components/`
- [X] T034 [P] [US6] Rebalance modal: preview → confirm/cancel in `web/src/components/`

**Checkpoint**: FR-013–FR-015 and SC-005.

---

## Phase 9: Polish & cross-cutting concerns

**Purpose**: Constitution alignment, contract accuracy, and verification.

- [X] T035 Review order/risk UX against `docs/trader-safety-cheatsheet.md` for close/rebalance flows (no silent exposure increase; clear labels)
- [X] T036 [P] Update `docs/runbook-incident.md` if new operational paths or env vars require on-call notes
- [X] T037 [P] Reconcile `specs/002-dashboard-operator-trading/contracts/openapi.yaml` with implemented handlers and DTO field names
- [X] T038 Run manual `quickstart.md` workflow (auth file, curl, two-terminal dev) and fix gaps in referenced paths only as needed
- [X] T039 `make lint`, `make test`, and `make test-web` clean at repo root for touched packages

---

## Dependencies & execution order

### Phase dependencies

- **Phase 1** → **Phase 2** → **Phases 3–8 (user stories)** → **Phase 9**
- **Phase 2** blocks all user stories: no protected trading data without session + middleware (US2–US6 all depend on Phase 2 + US1 for realistic UI flows)

### User story dependencies

| Story | Depends on | Notes |
|-------|------------|--------|
| US1 (P1) | Phase 2 | MVP; no other story |
| US2 (P2) | US1 | Needs authenticated shell |
| US3 (P3) | US1 | Can parallelize backend pieces after Phase 2 with US1 tests using cookies |
| US4 (P4) | US1 | List UI logically after overview shell (US2/US3 optional for routing only) |
| US5 (P5) | US4 | Navigation from list to detail |
| US6 (P6) | US4, US5 | Close from list/detail; rebalance from overview or dedicated entry |

### Parallel opportunities

- **Phase 1**: T002, T003 in parallel
- **Phase 2**: T006, T011 in parallel after T005 exists; T010 parallel once mux skeleton stable
- **US1**: T012 (tests) parallel with T015 after T013–T014 stubs exist; T015 vs T016 different files
- **US3**: T022, T023, T024 parallel once T021 API contract is stable
- **US4**: T027 parallel with late tweaks to T025–T026
- **US6**: T033, T034 parallel after T031–T032 API stable
- **Phase 9**: T036, T037 in parallel

---

## Parallel example: User Story 3

```bash
# After T021 overview handler exists:
Task T022 "Deribit/account reads in src/internal/dashboard/"
Task T023 "Mood + strategy degraded flags in src/internal/dashboard/overview.go"
Task T024 "Overview UI sections in web/src/"
```

---

## Implementation strategy

### MVP first (US1 only)

1. Complete Phase 1–2  
2. Complete Phase 3 (US1)  
3. **Stop** — validate login/session/logout and 401 boundaries  

### Incremental delivery

1. US1 → US2 (environment clarity) → US3 (context) → US4 (operations surface) → US5 (risk drill-down) → US6 (mutable actions)  
2. After each story, run story **Independent Test** from this file and `spec.md`  

### Suggested team split (after Phase 2)

- **Dev A**: US3 backend + chart  
- **Dev B**: US4–US5 positions  
- **Dev C**: US6 modals + preview/confirm wiring  

---

## Notes

- Money fields: **decimal strings** in JSON; no `float64` for policy limits on new paths (`plan.md`).  
- Session: valid until logout or **server invalidation** (allowlist removal, revocation); never 401 from idle/max-age alone (`research.md` R-02).  
- All tasks use checklist format with sequential IDs **T001–T039**.
