# Research: Dashboard robustness and operator clarity

**Feature**: `003-dashboard-e2e-health` | **Date**: 2026-03-31

## R-1: Why `/positions` shows HTTP 503 in dev E2E and browser console

**Decision**: Treat **503** with `error: exchange_unavailable` as a **supported degraded mode** when `s.xchg == nil` (`positions.go`), including Playwright’s deliberate `env -u DERIBIT_*` startup. The product shall surface this in the **UI** per **FR-001**, not only via network logs.

**Rationale**: Console network entries are expected when Axios receives non-2xx; reducing **operator confusion** means consistent panel messaging and no infinite “Loading…” when both requests fail.

**Alternatives considered**:

- Mock exchange in E2E — heavier setup; deferred unless tests need non-degraded lists.
- Change API to 200 with empty body on nil exchange — would misrepresent “cannot load from exchange” vs “empty portfolio”; rejected vs spec edge cases.

## R-2: Human-readable uptime and heap

**Decision**: Format **client-side** from existing `health.uptime_seconds` and `health.memory_heap_alloc_bytes` (`handleHealthAPI`); no BFF schema break. Use clear compound duration (e.g. `Xd Yh` or `Yh Zm` below one day) and binary IEC or decimal MB/GB with unit suffix — pick one style and apply consistently in `HealthPanel`.

**Rationale**: Constitution **II** discourages uncoordinated API shape changes; presentation belongs in SPA.

**Alternatives considered**:

- Server sends pre-formatted strings — duplicates localization logic and couples copy to Go releases.

## R-3: Market mood placeholder text

**Decision**: Replace `"strategy modules not wired to dashboard yet"` with **operator-facing** unavailable text (e.g. “Market mood is not available yet.”) while `available: false`; keep JSON keys stable for `Overview.tsx`.

**Rationale**: Satisfies **FR-002** without requiring a full strategy analytics integration in this feature.

**Alternatives considered**:

- Hide mood section when unavailable — conflicts with spec expectation of visible summary/degraded state.

## R-4: Full automated test suite (**FR-007**)

**Decision**: Define the release gate as:

```bash
make test && make lint && make test-web && make test-e2e
```

after adding `test-e2e` to the root `Makefile` (wrapper around `web` `npm run test:e2e`). Document in [quickstart.md](./quickstart.md).

**Rationale**: Current `make test-web` only runs production build, not Playwright; spec and stakeholders expect E2E in the final gate.

**Alternatives considered**:

- Rely on CI only — weakens local “all tasks done” verification from **Clarifications**.

## R-5: E2E breadth vs. spec 002

**Decision**: Map journeys to **P1–P4** flows in spec 002: sign-in, overview shell + health + mood/strategy panels, positions list degraded state, optional position detail route smoke. Minimum **five** distinct scenarios in Playwright after implementation (**SC-005**), counting login, overview, positions-unavailable, health formatting, and one navigation/negative case.

**Rationale**: Balances runtime with **FR-004** / **FR-005** without duplicating entire spec in browser tests.

**Alternatives considered**:

- Full visual regression — out of scope; not requested.
