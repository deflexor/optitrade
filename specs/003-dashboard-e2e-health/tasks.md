# Tasks: Dashboard robustness and operator clarity

**Input**: Design documents from `/home/dfr/optitrade/specs/003-dashboard-e2e-health/`  
**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/](./contracts/), [quickstart.md](./quickstart.md)

**Tests**: Included — feature spec requires automated end-to-end coverage (**FR-004**–**FR-006**, **SC-005**), Go handler tests per constitution, and final full-suite gate (**FR-007**).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependency on incomplete sibling tasks in the same phase where noted)
- **[Story]**: User story label from [spec.md](./spec.md) (`[US1]` … `[US4]`)

## Path Conventions

Repository root: `/home/dfr/optitrade`. Go BFF: `/home/dfr/optitrade/src/`. Web app: `/home/dfr/optitrade/web/`.

---

## Phase 1: Setup (shared infrastructure)

**Purpose**: Wire Makefile E2E target and sync docs so verification matches **FR-007**.

- [x] T001 Add `test-e2e` target and `.PHONY: test-e2e` entry in `/home/dfr/optitrade/Makefile` running `npm run test:e2e` in `$(WEB_DIR)` (after `web-install` or document `npm ci` prerequisite in target comments)
- [x] T002 Update full gate section in `/home/dfr/optitrade/specs/003-dashboard-e2e-health/quickstart.md` to use `make test-e2e` without “until Makefile updated” fallback wording

---

## Phase 2: Foundational (blocking prerequisites)

**Purpose**: Operator-safety awareness and discoverability before story implementation.

**⚠️ No heavy codegen here** — user stories may start after Phase 1 for code paths; complete Phase 2 before merging if your process requires safety sign-off.

- [x] T003 [P] Keep edits scoped to `/home/dfr/optitrade/src/internal/dashboard/`, `/home/dfr/optitrade/web/`, `/home/dfr/optitrade/Makefile`, and `/home/dfr/optitrade/specs/003-dashboard-e2e-health/`; avoid drive-by changes under `/home/dfr/optitrade/src/internal/execution/` unless spec expands (constitution **VII**)
- [x] T004 [P] Extend `help` text in `/home/dfr/optitrade/Makefile` to list `test-e2e` with other test targets

**Checkpoint**: Phase 1–2 complete — proceed with user stories (can overlap US1–US4 work across agents after T001–T002).

---

## Phase 3: User Story 1 — Trustworthy positions views (Priority: P1) 🎯 MVP

**Goal**: Positions UI shows explicit degraded state when BFF returns errors (e.g. 503 `exchange_unavailable`); no perpetual “Loading…” on **Open** / **Closed** when fetch fails.

**Independent Test**: With exchange unset (Playwright default), after login, `/positions` shows operator-visible error/unavailable messaging for lists; Go tests assert 503 JSON for nil exchange.

### Tests for User Story 1

- [x] T005 [P] [US1] Add `/home/dfr/optitrade/src/internal/dashboard/positions_handlers_test.go` using `httptest` + authenticated session cookie pattern from `/home/dfr/optitrade/src/internal/dashboard/overview_test.go`; assert `GET /api/v1/positions/open` and `GET /api/v1/positions/closed` return **503** with `{"error":"exchange_unavailable",...}` when `Exchange: nil`
- [x] T006 [P] [US1] Add `/home/dfr/optitrade/web/e2e/positions-degraded.spec.ts` — sign in as dev operator, open `/positions`, expect clear error/degraded copy (not paired indefinite “Loading…” for both sections)

### Implementation for User Story 1

- [x] T007 [P] [US1] Add `/home/dfr/optitrade/web/src/api/parseApiError.ts` to extract stable `{ error, message }` from Axios errors matching `/home/dfr/optitrade/specs/003-dashboard-e2e-health/contracts/error-response.md`
- [x] T008 [US1] Refactor `/home/dfr/optitrade/web/src/pages/PositionsPage.tsx` to clear loading flags on failure, set sensible `open`/`closed` when applicable, and show operator-facing text (optionally keyed off `exchange_unavailable`) per **FR-001** (depends on T007)

**Checkpoint**: US1 behaves correctly under nil exchange; new Go + Playwright tests pass.

---

## Phase 4: User Story 2 — Readable health at a glance (Priority: P2)

**Goal**: Health panel shows uptime and heap allocation in human-meaningful units (**FR-003**).

**Independent Test**: Visual/regression: health area shows formatted duration and capacity-style memory, not raw `Ns` + `bytes` only.

### Tests for User Story 2

- [x] T009 [P] [US2] Add `/home/dfr/optitrade/web/e2e/health-format.spec.ts` asserting process health line matches formatted uptime/heap (e.g. contains unit labels like `h`, `m`, `MB`/`GB`/`MiB` per chosen convention — avoid asserting only trailing `bytes`)

### Implementation for User Story 2

- [x] T010 [P] [US2] Add `/home/dfr/optitrade/web/src/lib/formatHealth.ts` with `formatUptimeSeconds(number): string` and `formatHeapBytes(number): string` (pure functions; one consistent byte convention)
- [x] T011 [US2] Update `/home/dfr/optitrade/web/src/components/HealthPanel.tsx` to render formatted uptime and heap using `/home/dfr/optitrade/web/src/lib/formatHealth.ts` (depends on T010)

**Checkpoint**: US2 satisfies readable health; E2E passes.

---

## Phase 5: User Story 3 — Meaningful market mood (Priority: P3)

**Goal**: Remove developer-placeholder mood copy; operators see appropriate unavailable text (**FR-002**).

**Independent Test**: Overview JSON and Playwright deny old placeholder string.

### Tests for User Story 3

- [x] T012 [P] [US3] Extend `/home/dfr/optitrade/src/internal/dashboard/overview_test.go` (or add focused test) asserting `market_mood.explanation` in JSON does not contain internal wiring phrase `"strategy modules not wired"` after handler change
- [x] T013 [P] [US3] Add `/home/dfr/optitrade/web/e2e/market-mood-copy.spec.ts` (or merge into existing e2e file) asserting **Market mood** section text excludes dev placeholder substring after login

### Implementation for User Story 3

- [x] T014 [US3] Edit `/home/dfr/optitrade/src/internal/dashboard/overview.go` — set `market_mood.explanation` to operator-facing unavailable message while `available: false` per `/home/dfr/optitrade/specs/003-dashboard-e2e-health/contracts/overview-fragments.md`

**Checkpoint**: US3 complete; Go + UI tests green.

---

## Phase 6: User Story 4 — Dashboard specification conformance regression (Priority: P4)

**Goal**: Broaden Playwright coverage toward **SC-005** (≥5 primary journeys) aligned with `/home/dfr/optitrade/specs/002-dashboard-operator-trading/spec.md` routes/panels.

**Independent Test**: `npm run test:e2e` includes distinct scenarios: sign-in, overview health + mood/strategy presence, positions degraded, health formatting, navigation smoke.

### Tests for User Story 4

- [x] T015 [P] [US4] Add `/home/dfr/optitrade/web/e2e/dashboard-conformance.spec.ts` with additional `test()` blocks so the **total** suite covers at least **five** distinct operator journeys (reuse login helper pattern from `/home/dfr/optitrade/web/e2e/auth-overview.spec.ts`); include navigation to `/`, `/positions`, and assertions on key headings/regions (Balance, P/L, Market mood, Strategy, Open/Closed sections as applicable)

### Implementation for User Story 4

- [x] T016 [US4] Deduplicate or factor shared Playwright login steps in `/home/dfr/optitrade/web/e2e/` (e.g. `/home/dfr/optitrade/web/e2e/fixtures.ts` or `helpers/login.ts`) if **T015** duplicates more than ~5 lines — keep DRY without over-engineering

**Checkpoint**: US4 raises E2E breadth; no flaky timing (use `expect(...).toBeVisible` with defaults).

---

## Phase 7: Polish & cross-cutting concerns

**Purpose**: Constitution gates, full verification, minor docs.

- [x] T017 Run `make lint` from `/home/dfr/optitrade` and fix issues in touched Go packages under `/home/dfr/optitrade/src/internal/dashboard/`
- [x] T018 [P] Run `make test` from `/home/dfr/optitrade` ensuring `/home/dfr/optitrade/src/internal/dashboard/` tests pass
- [x] T019 [P] Run `make test-web` from `/home/dfr/optitrade` after `/home/dfr/optitrade/web/` TypeScript changes
- [x] T020 Run full gate at `/home/dfr/optitrade`: `make test && make lint && make test-web && make test-e2e` per **FR-007** / **SC-006** in `/home/dfr/optitrade/specs/003-dashboard-e2e-health/spec.md`; iterate until green

---

## Dependencies & execution order

### Phase dependencies

| Phase | Depends on |
|-------|------------|
| Phase 1 | — |
| Phase 2 | Phase 1 recommended (docs accurate for `make test-e2e`) |
| Phase 3–6 (US1–US4) | Phase 1 complete; Phase 2 parallel-friendly |
| Phase 7 | All desired user stories done |

### User story dependencies

| Story | Depends on |
|-------|------------|
| **US1** | T001 (E2E runner exists for local dev); implementation tasks T007→T008 |
| **US2** | Independent of US1 code paths (different files); T010→T011 |
| **US3** | Independent (Go overview); T014 after T012–T013 if following TDD order |
| **US4** | Benefits from all prior e2e stability; T016 optional cleanup after T015 |

### Within each user story

- Prefer **failing test first** where tasks list tests before implementation (US1: T005–T006 before T007–T008; US3: T012–T013 before T014).
- **T008** depends on **T007**.
- **T011** depends on **T010**.

### Parallel opportunities

- **T005 ∥ T006 ∥ T007** after Phase 1: shared story US1 but different files (Go test vs e2e vs TS helper) — run T008 after T007.
- **US1 ∥ US2 ∥ US3** can proceed in parallel across developers after Phase 1 (touch disjoint files until US4 consolidates e2e).
- **T012 ∥ T013** parallel before **T014**.
- **T018 ∥ T019** after code complete.

---

## Parallel example: User Story 1

```bash
# After Phase 1, launch in parallel:
T005  # Go handler tests positions_handlers_test.go
T006  # Playwright positions-degraded.spec.ts (may fail until T008 lands)
T007  # parseApiError.ts

# Then sequential:
T008  # PositionsPage.tsx consumes T007
```

---

## Parallel example: User Story 2

```bash
T010  # formatHealth.ts
T009  # health-format.spec.ts (fails until T011)
# Then:
T011  # HealthPanel.tsx
```

---

## Implementation strategy

### MVP first (User Story 1 only)

1. Complete Phase 1–2  
2. Complete Phase 3 (US1)  
3. Run `make test`, `make test-web`, `make test-e2e` — stop and demo if sufficient for ops trust on positions  

### Incremental delivery

1. US1 → US2 → US3 → US4 in priority order, or parallelize US1–US3 then add US4 breadth  
2. Phase 7 full gate before declaring **FR-007** satisfied  

### Parallel team strategy

- Developer A: US1 (Go + PositionsPage + e2e)  
- Developer B: US2 (formatHealth + HealthPanel + e2e)  
- Developer C: US3 (overview.go + tests)  
- Integrator: US4 + Phase 7  

---

## Notes

- All tasks use absolute paths under `/home/dfr/optitrade/…` for clarity.  
- **Constitution**: handler JSON shape and `writeAPIError` patterns in `/home/dfr/optitrade/src/internal/dashboard/errors.go`; no secrets in logs.  
- **DoD**: **T020** must pass with zero failures before marking feature tasks complete (**SC-006**).

---

## Task summary

| Metric | Count |
|--------|--------|
| **Total tasks** | 20 (T001–T020) |
| **Phase 1 — Setup** | 2 |
| **Phase 2 — Foundational** | 2 |
| **US1 (Phase 3)** | 4 (T005–T008) |
| **US2 (Phase 4)** | 3 (T009–T011) |
| **US3 (Phase 5)** | 3 (T012–T014) |
| **US4 (Phase 6)** | 2 (T015–T016) |
| **Phase 7 — Polish** | 4 (T017–T020) |
| **Parallel-friendly [P]** | T003, T004, T005, T006, T007, T009, T010, T012, T013, T015, T018, T019 |
</think>
Fixing task T003 to include explicit paths and cleaning the summary section.

<｜tool▁calls▁begin｜><｜tool▁call▁begin｜>
StrReplace