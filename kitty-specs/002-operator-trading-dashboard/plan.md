# Implementation Plan: Operator Trading Dashboard

*Path: [.kittify/missions/software-dev/templates/plan-template.md](.kittify/missions/software-dev/templates/plan-template.md)*

**Branch**: `master` (planning repo; feature slug `002-operator-trading-dashboard`) | **Date**: 2026-03-28 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification at `kitty-specs/002-operator-trading-dashboard/spec.md` plus clarifications (sessions 2026-03-28): v1 auth with username allowlist, equity as primary balance, default P/L chart 30 days, summary staleness 5 seconds, paired **Regime** / **Market mood** labels.

**Engineering alignment**: Ship a **React + Vite** SPA with **Tailwind CSS**, **Zustand**, and **Axios** (per spec Assumptions). The **Go execution binary** (`src/cmd/optitrade`) gains an **embedded HTTP API** (JSON, cookie sessions) that reads the same process state and SQLite as the bot, serves the SPA in production (embedded `web/dist`), and exposes endpoints matching `contracts/dashboard-api.openapi.yaml`. No separate microservice for MVP.

## Summary

Build an operator web dashboard that shows process health, Deribit mode (test vs live), **equity**, paired regime/mood, P/L series (default **last 30 days**), open positions, ten latest closed positions, and position detail with close/rebalance flows. **v1 auth**: username/password registration (no email), allowlist-gated login with exact `Sorry, feature not ready` for others, salted password hashing, httpOnly session cookie. **Freshness**: summary bundle older than **5 seconds** (server clock) surfaces explicit stale UI. Implementation splits **web/** (frontend) and **src/internal/dashboard/** (HTTP handlers, auth, aggregation) with contract tests at the REST boundary.

## Technical Context

**Language/Version**: Go 1.26+ (API and auth, shared with existing `src/`); TypeScript 5.x / React 18+ (dashboard UI)  
**Primary Dependencies**: Go: `net/http`, chi or std mux; `golang.org/x/crypto/bcrypt` or `argon2` for passwords; existing `internal/deribit`, `internal/regime`, `internal/state`, `internal/risk`. Web: Vite, React, Tailwind, Zustand, Axios, React Router (or equivalent minimal routing).  
**Storage**: New SQLite migrations under `dashboard_` tables for users and sessions; existing bot SQLite for positions/audit where applicable; time-series for P/L may be aggregated table or materialized snapshots (see `data-model.md`).  
**Testing**: Go: `go test` for auth allowlist matrix, session validation, snapshot TTL math (fake clock), handler tests with `httptest`. Frontend: Vitest + Testing Library for stores and critical modals; Playwright optional later. Contract: schemathesis or openapi examples against mock server.  
**Target Platform**: Linux (same as bot); browser: modern evergreen (Chrome/Firefox).  
**Project Type**: **Monorepo extension**: existing `src/` Go service plus new `web/` SPA; single deployable binary embeds SPA assets.  
**Performance Goals**: Summary **GET** p95 under **150 ms** on idle bot when cache warm; UI poll interval **2 to 3 seconds** while focused so staleness beyond 5 s is rare when healthy; no unbounded fan-out on Deribit from dashboard handlers (reuse bot caches).  
**Constraints**: Constitution: validate all auth inputs; never log passwords or raw session tokens; CSRF strategy documented (same-site cookies + optional double-submit for mutations); 5 s staleness rule enforced using server-issued `snapshot_utc` (see FR-019).  
**Scale/Scope**: Single operator deployment; tens of concurrent dashboard tabs max; allowlist typically one usernames.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle (from `.kittify/memory/constitution.md`) | How this plan complies |
|----------------------------------------------------|-------------------------|
| Testing MUST cover critical paths (auth, safety) | Table-driven tests for allowlist vs generic deny vs not-ready message; session fixation and expiry; close/rebalance endpoints require auth. |
| MUST NOT log secrets | Redact credentials; audit login success/fail without password material; session IDs logged as truncated hash only if needed. |
| Parameterized DB access | All dashboard SQL via prepared statements; same patterns as `internal/state/sqlite`. |
| Performance: no unbounded work | Snapshot builder reads bounded fields; P/L query capped to requested window; rate limit login endpoint. |
| UX consistency | Loading and stale states on all async regions; paired Regime/Market mood always; errors actionable. |
| Public APIs documented | OpenAPI for `/api/v1/*`; handler docstrings reference error shapes. |

**Post-Phase 1 re-check**: PASS. Design adds explicit error JSON schema and staleness field on snapshot; no new MUST violations.

## Project Structure

### Documentation (this feature)

```
kitty-specs/002-operator-trading-dashboard/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
└── tasks.md             # Phase 2 only: created by /spec-kitty.tasks
```

### Source Code (repository root)

```
web/                              # NEW: Vite + React dashboard
├── package.json
├── vite.config.ts
├── index.html
├── tailwind.config.js
└── src/
    ├── main.tsx
    ├── App.tsx
    ├── api/client.ts             # Axios instance, cookie creds
    ├── stores/authStore.ts       # Zustand
    ├── stores/dashboardStore.ts
    ├── components/
    └── pages/                    # list, position detail, login/register

src/                              # EXISTING Go tree
├── cmd/optitrade/main.go         # start HTTP dashboard listener when flag set
└── internal/
    └── dashboard/                # NEW
        ├── server.go             # mux, middleware, static embed
        ├── auth_handlers.go
        ├── snapshot.go           # health + equity + regime pair + mode
        ├── positions_handlers.go
        └── session/sqlite        # or under internal/state/
```

**Structure Decision**: **Monorepo web + Go BFF**: new `web/` for the SPA; dashboard HTTP and auth live under `src/internal/dashboard/` and wire into `optitrade` when `-dashboard.listen=:PORT` (or config) is enabled. Production build copies `web/dist` into `embed.FS`.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| N/A | | |

## Parallel Work Analysis

### Dependency Graph

```
Contracts + data-model (shared) -> Go auth + session store -> Go snapshot/positions API
                              \-> React shell + auth pages -> dashboard widgets -> position modals
Integration: cookie login + polling snapshot after API stable
```

### Work Distribution

- **Sequential work**: SQLite migrations for dashboard users/sessions before login E2E.  
- **Parallel streams**: Frontend layout and styling vs Go HTTP skeleton vs OpenAPI fixture tests.  
- **Agent assignments**: One track on `web/`, one on `src/internal/dashboard/`, merge via shared OpenAPI types (optional codegen later).

### Coordination Points

- **Contract freeze**: First usable milestone when `/api/v1/auth/login` and `/api/v1/snapshot` match OpenAPI.  
- **Integration tests**: httptest + headless login flow before wiring real Deribit fields (mocks/fakes first).

## Phase Outputs (this command)

| Phase | Artifact | Path |
|-------|-----------|------|
| 0 | Research decisions | `kitty-specs/002-operator-trading-dashboard/research.md` |
| 1 | Data model | `kitty-specs/002-operator-trading-dashboard/data-model.md` |
| 1 | Contracts | `kitty-specs/002-operator-trading-dashboard/contracts/` |
| 1 | Quickstart | `kitty-specs/002-operator-trading-dashboard/quickstart.md` |

**Stop**: Task generation is **not** performed here. Run `/spec-kitty.tasks` next.
