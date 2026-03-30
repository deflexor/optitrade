# Implementation Plan: Operator dashboard trading and controls

**Branch**: `002-dashboard-operator-trading` | **Date**: 2026-03-30 | **Spec**: [spec.md](./spec.md)  
**Input**: Feature specification from `/specs/002-dashboard-operator-trading/spec.md` (incl. clarifications session 2026-03-30)

## Summary

Deliver an **authenticated operator dashboard**: allowlisted **username/password** from **validated config** (no email flow), **sessions until explicit sign-out** (no idle/max-age expiry), and **trading/observability UI** backed by the **Go BFF** (`src/internal/dashboard`) with **React + Vite** (`web/`). Scope includes **health + test/live mode**, **balance**, **30-day P/L chart**, **market mood**, **strategy metadata**, **paginated open positions (25/page)**, **closed positions (30 days, max 200)** with **labeled % basis**, **position detail** (legs, liquidity, Greeks), and **close / rebalance** flows with **preview + confirm** and operator-safe copy. Exchange and capital logic stay in **Go**; SPA is presentation and confirmation only.

## Technical Context

**Language/Version**: Go **1.26** (`src/go.mod`), TypeScript/React (**web/**, Node 20+).  
**Primary Dependencies**: `net/http` BFF; `modernc.org/sqlite` for sessions; `internal/deribit` for exchange; React Router, Axios, Tailwind (existing scaffold), Zustand stub in `web/src/stores/authStore.ts`.  
**Storage**: SQLite (**sessions**); **JSON file** for dashboard allowlist + password hashes (new schema; see `research.md`).  
**Testing**: `make test`, `make lint`, `make test-web`; new `internal/dashboard` handler tests for auth + one trading read path before “production-ready” claim.  
**Target Platform**: Linux (WSL/server); dashboard listens on configurable `OPTITRADE_DASHBOARD_LISTEN` / `-listen`.  
**Project Type**: **Web application** — split `src/` (CLI + BFF + domain) and `web/` (SPA embedded via `go:embed`).  
**Performance Goals**: Operator-scale (single-digit concurrent users); BFF Deribit calls bounded with **context timeouts** (align with `internal/deribit/rpc` patterns).  
**Constraints**: No secrets in browser logs; **slog** + redaction; money as **decimal strings** in JSON; **no float64** for policy-style limits in new code paths.  
**Scale/Scope**: ~6 primary UI flows (auth, overview, positions list/detail, close modal, rebalance modal); **25** open rows per page; **200** closes cap.

## Constitution Check

*GATE: Passed before Phase 0. Re-checked after Phase 1 design below.*

Aligned with `.specify/memory/constitution.md` (v1.1.0):

| Principle | Status |
|-----------|--------|
| **I Safety** | Pass — preview/confirm for close & rebalance; no silent exposure increase; estimates labeled. |
| **II Boundaries** | Pass — trading/Deribit in Go; SPA consumes `/api/v1` only. |
| **III Code quality** | Pass — decimal strings for money fields in contracts; structured errors; no password logging. |
| **IV Testing** | Pass — plan mandates handler tests as behavior lands; integration stays env-gated. |
| **V UX** | Pass — Tailwind/shell in `App.tsx`; stable JSON errors; cookie + `withCredentials` per existing client. |
| **VI Performance/ops** | Pass — context on outbound RPC; `/healthz` remains; panel-level degraded states. |
| **VII Scope** | Pass — minimal new packages; incremental endpoints. |

## Post-design constitution re-check

- **OpenAPI + data model** keep money as **strings** and **preview/confirm** for mutable operations — still aligned.
- **Session “until logout”** requires **no time-based auth gate**; if SQLite schema retains `expires_at`, it must not drive 401s (see `research.md` R-02) — **documented**, not a constitution violation.

## Project Structure

### Documentation (this feature)

```text
specs/002-dashboard-operator-trading/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── openapi.yaml
├── spec.md
├── checklists/
│   └── requirements.md
└── tasks.md              # from /speckit.tasks (not created here)
```

### Source code (repository root)

```text
src/
├── cmd/optitrade/              # dashboard command, signal shutdown
├── internal/
│   ├── dashboard/              # BFF: server.go, handlers, auth, embed
│   ├── config/                 # policy load; add dashboard auth schema or sibling package
│   ├── deribit/                # RPC/WS — reused by dashboard read/write
│   ├── state/
│   │   └── migrations/         # 0003_dashboard_auth.sql — revise per Complexity Tracking
│   └── ...
web/
├── src/
│   ├── App.tsx                 # shell, routes, protected layout
│   ├── api/client.ts
│   ├── stores/authStore.ts
│   ├── pages/                  # login, overview, positions (to add)
│   └── components/           # modals, charts (to add)
```

**Structure decision**: **Brownfield split** — extend `internal/dashboard` from stub (`dashboard API not implemented`) and grow `web/` under existing Vite + Tailwind setup per README.

## Phase 0 — Research

**Output**: [`research.md`](./research.md) — resolved items: config allowlist, session semantics, cookie transport, BFF-only Deribit, pagination, % basis fields, preview/confirm pattern, chart window, degraded mood/strategy, health metrics.

No remaining `NEEDS CLARIFICATION` blockers for planning.

## Phase 1 — Design & contracts

**Outputs**:

- [`data-model.md`](./data-model.md) — entities and DTOs.
- [`contracts/openapi.yaml`](./contracts/openapi.yaml) — endpoint skeleton (auth, overview, positions, preview/confirm).
- [`quickstart.md`](./quickstart.md) — dev workflow + env vars.

**Agent context**: Run `.specify/scripts/bash/update-agent-context.sh cursor-agent` after committing plan artifacts.

## Implementation sequencing (recommended)

1. **Auth foundation**: Config schema + load; session store + cookie; login/logout/me; middleware on `/api/v1/**` except login; tests.
2. **Read-only overview**: `/overview` + `/health` (uptime, mem, mode); wire testnet/mainnet detection from env/config already used elsewhere.
3. **Positions read**: open paginated + closed list + detail; Deribit mapping layer in Go.
4. **Charts & labels**: 30d P/L series; closed row **percent_basis_label**; empty/error states.
5. **Close / rebalance**: preview + confirm endpoints; SPA modals with blocking confirmation.
6. **Polish**: market mood + strategy meta from real modules or explicit degraded flags.

## Complexity Tracking

| Item | Why needed | Simpler alternative rejected because |
|------|------------|----------------------------------------|
| **Revise `0003_dashboard_auth.sql` / new migration** | Spec requires **config** allowlist, not `dashboard_user` rows; sessions must **not** expire on idle/clock; existing table has `expires_at` + FK pattern that fights FR-017. | Keeping user table as source of truth **violates** spec (config allowlist). |
| **Preview + confirm API** | Two-step pattern reduces mistaken live orders and satisfies FR-013–015. | Single POST is weaker for operator safety under stress. |

No constitution MUST violations requiring waiver.
