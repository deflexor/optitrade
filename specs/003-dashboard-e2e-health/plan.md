# Implementation Plan: Dashboard robustness and operator clarity

**Branch**: `003-dashboard-e2e-health` | **Date**: 2026-03-31 | **Spec**: [spec.md](./spec.md)  
**Input**: Feature specification from `/specs/003-dashboard-e2e-health/spec.md`

## Summary

Harden the operator dashboard so **positions** behave correctly when the exchange client is absent or failing (today: HTTP 503 `exchange_unavailable` from `internal/dashboard/positions.go`), **health** shows uptime and heap in human-readable units (today: raw seconds and bytes in `web/src/components/HealthPanel.tsx`), **market mood** uses operator-facing copy instead of the backend placeholder string in `internal/dashboard/overview.go`, and **Playwright** coverage expands to assert key panels, routes, and degraded states aligned with `specs/002-dashboard-operator-trading/spec.md`. Document and add a **Makefile** target so the **full verification gate** matches spec **FR-007** / **SC-006**: `make test`, `make lint`, `make test-web`, and Playwright E2E (`web` `npm run test:e2e`).

## Technical Context

**Language/Version**: Go 1.26 (`src/go.mod`); TypeScript 5.9+ / Node 20+ (`web/`).  
**Primary Dependencies**: Go `net/http` BFF (`src/internal/dashboard`), React 19 + Vite 8 (`web/`), Axios (`web/src/api/client.ts`), Playwright (`@playwright/test`).  
**Storage**: SQLite sessions for operator auth (existing); no new persistent stores for this feature.  
**Testing**: `make test` (Go `go test ./...` + `research/` pytest via `uv`), `make lint` (`go vet`), `make test-web` (`web` build: `tsc -b && vite build`), Playwright `web/npm run test:e2e` (starts BFF + Vite via `web/playwright.config.ts`). Integration: `make test-integration` (tagged, optional creds).  
**Target Platform**: Linux/macOS dev; dashboard served from `optitrade dashboard` + static SPA (embedded or Vite dev proxy).  
**Project Type**: Brownfield monorepo: Go CLI + library + dashboard BFF; React SPA.  
**Performance Goals**: No new hot paths; E2E suite completes locally in minutes; avoid unbounded polling in tests.  
**Constraints**: Constitution v1.1.0 — BFF owns `/api/v1` JSON contracts; UI uses shared Axios client; stable `{ error, message }` error JSON; no duplicated trading rules in SPA.  
**Scale/Scope**: Small operator set; dashboard routes (`/`, `/login`, `/positions`, `/positions/:id`); ≥5 primary E2E journeys after implementation per spec **SC-005**.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-checked after Phase 1 design.*

| Principle | Status |
|-----------|--------|
| I Operator safety | **Pass** — UI clarity on unavailable exchange reduces mistaken assumptions; no change to order sizing/policy in this feature. |
| II Boundaries | **Pass** — Formatting and copy changes stay presentational; positions error semantics remain server-defined. |
| III Code quality | **Pass** — Use existing error helpers (`writeAPIError`); any new logging stays `slog` + redaction patterns. |
| IV Testing | **Pass** — Extend Go handler tests where behavior changes; add Playwright specs; document full gate including E2E (constitution lists `make test`/`test-web`/`lint`; plan adds explicit E2E target — not a contradiction if Makefile becomes canonical). |
| V UX | **Pass** — Tailwind/shell preserved; operator-readable strings; JSON errors unchanged in shape. |
| VI Performance/ops | **Pass** — No new unbounded RPC; health still uses existing `GetServerTime` check when exchange present. |
| VII Focused delivery | **Pass** — Scoped to presentation, copy, tests, and Makefile wiring. |

**Post-design re-check**: No new violations; Complexity Tracking not required.

## Project Structure

### Documentation (this feature)

```text
specs/003-dashboard-e2e-health/
├── plan.md              # This file
├── research.md          # Phase 0
├── data-model.md        # Phase 1
├── quickstart.md        # Phase 1
├── contracts/           # Phase 1 — HTTP JSON contracts
└── tasks.md             # /speckit.tasks (separate command)
```

### Source Code (repository root)

```text
/home/dfr/optitrade/src/
├── cmd/optitrade/           # CLI including dashboard command
├── internal/dashboard/      # BFF: health, overview, positions, auth, tests
└── internal/deribit/        # Exchange client (nil when unconfigured)

/home/dfr/optitrade/web/
├── src/
│   ├── api/client.ts       # Axios instance (VITE_API_BASE, credentials)
│   ├── components/         # HealthPanel, modals
│   └── pages/              # Overview, PositionsPage, PositionDetail, Login
├── e2e/                    # Playwright specs
└── playwright.config.ts    # webServer: Go dashboard + Vite

/home/dfr/optitrade/Makefile # test, lint, test-web (add test-e2e)
```

**Structure Decision**: Single brownfield repo (`src/` + `web/`); feature touches `internal/dashboard`, `web/src`, `web/e2e`, and `Makefile` only unless tests require small shared TS utilities under `web/src/lib/`.

## Complexity Tracking

No constitution MUST violations requiring justification.

## Phase 0 & 1 outputs

| Artifact | Path |
|----------|------|
| Research | [research.md](./research.md) |
| Data model (presentation) | [data-model.md](./data-model.md) |
| Contracts | [contracts/](./contracts/) |
| Verify / full gate | [quickstart.md](./quickstart.md) |

## Implementation notes (for /speckit.tasks)

1. **Positions UX**: On fetch failure (503/5xx/network), set panel-level state so **Open** and **Closed** sections do not stay on “Loading…” while a global error exists; optionally differentiate `exchange_unavailable` with copy aligned to BFF `message` / `error` code ([contracts/error-response.md](./contracts/error-response.md)).
2. **Health formatting**: Add small formatters (e.g. `formatUptime(seconds)`, `formatHeap(bytes)`) in `web/`; keep numeric fields from API unchanged.
3. **Market mood**: Replace placeholder `explanation` in `overview.go` with operator-appropriate “unavailable” text; long-term wiring to real strategy analytics is out of scope unless already available — prefer honest unavailable state per **FR-002**.
4. **Playwright**: Add journeys: authenticated `/positions` with exchange off (expect explicit error/empty policy per FR-001), `/` health panel readable strings, market mood copy check, navigation smoke for spec-critical links; keep **`webServer`** commands in sync with `playwright.config.ts`.
5. **Makefile**: Add `test-e2e` (or equivalent) invoking `cd web && npm run test:e2e`; document **full gate** in [quickstart.md](./quickstart.md) for **FR-007**.

---

*Template workflow: Phase 0 research and Phase 1 design artifacts are filed alongside this plan.*
