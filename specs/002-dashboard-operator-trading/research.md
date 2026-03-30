# Research: Operator dashboard trading and controls

**Feature**: `002-dashboard-operator-trading`  
**Date**: 2026-03-30

Consolidated decisions for unknowns and integration patterns. All items below unblock implementation and match `spec.md` clarifications.

## R-01 — Allowlisted operators: config vs database

**Decision**: Primary source of truth for **username + password verifier** is a **validated JSON configuration** document (dedicated schema under `src/internal/config` or `src/internal/dashboard`), loaded at process start (path via env e.g. `OPTITRADE_DASHBOARD_AUTH_PATH`). Each entry: `username`, `password_hash` (bcrypt or argon2id string in canonical encoding).

**Rationale**: Matches spec FR-001/003 and clarified “list in backend config”; avoids drift between DB and deployed config; aligns with existing `config.LoadBytes` + JSON Schema pattern.

**Alternatives considered**:

- SQLite `dashboard_user` only (`0003_dashboard_auth.sql`) — rejected as primary store: spec calls out **config allowlist**, not operator-managed DB rows.
- Hybrid: config seeds DB — rejected: doubles sources of truth without strong benefit for small operator sets.

**Follow-up**: Existing migration `0003_dashboard_auth.sql` is **superseded for auth identity**; a follow-on migration should model **sessions only** keyed by `username` (text) + `token_hash`, or drop `dashboard_user` and adjust `dashboard_session` FK (see plan Complexity Tracking).

## R-02 — Session semantics (until sign-out)

**Decision**: Issue an opaque session token (random bytes, **hash at rest**). Persist session row for revocation and logout. **Do not** reject requests based on `expires_at` or idle clock **alone**; only **logout**, **server invalidation** (allowlist/user removed from config → reject on next request after reload), **token delete**, or **process restart** (if tokens are memory-only **optional** — prefer DB/memory store documented in plan) ends the session.

**Rationale**: FR-017/018 and clarifications; avoids surprising re-auth during long watches.

**Alternatives considered**:

- Rolling idle timeout — rejected by product decision.
- JWT without server store — rejected: explicit logout and allowlist revocation require server-side invalidation list or session rows.

**Implementation note**: If `dashboard_session.expires_at` remains in schema, set to a **sentinel far-future** value or ignore column in auth middleware; prefer migration to drop the column to avoid confusion.

## R-03 — Session transport

**Decision**: **HTTP-only**, **Secure** (when TLS), **SameSite=Lax** cookie carrying opaque session id **or** double-submit cookie pattern; `web` already uses `withCredentials: true` on Axios.

**Rationale**: Cookie aligns with BFF same-origin in production embed; reduces XSS exfil vs localStorage token.

**Alternatives**: Bearer header from memory — acceptable for SPA if cookie blocked; document in quickstart if used for API tooling only.

## R-04 — Where trading truth lives

**Decision**: All balances, positions, P/L series, Greeks, and **order placement** go through **Go BFF** calling existing **`internal/deribit`** (and future strategy/risk packages). SPA renders and confirms; **no** Deribit keys in the browser.

**Rationale**: Constitution II — server is source of truth; capital safety.

**Alternatives**: Direct browser to Deribit — rejected.

## R-05 — Pagination and list contracts

**Decision**: Open positions: **`limit=25`**, **`cursor`** or **`page`** + **`total_count`** in JSON; default sort defined in BFF (documented in OpenAPI). Closed positions: single response up to **200** rows, **30-day** filter server-side.

**Rationale**: FR-009/010 and SC-009.

## R-06 — Percent P/L presentation

**Decision**: API returns `percent_pl` (string decimal) + `percent_basis` (short machine key) + `percent_basis_label` (human string). UI always shows label adjacent or in header legend.

**Rationale**: FR-011, SC-006.

## R-07 — Close / rebalance safety

**Decision**: Two-step API: `POST .../preview` (idempotent read of estimate) then `POST .../confirm` with **client echo** of key fields or server-issued **preview token** to prevent stale double-submit. UI must show preview payload before confirm (FR-013–FR-015).

**Rationale**: Operator safety; aligns with constitution I.

**Alternatives**: Single POST with `?confirm=true` — weaker against accidental resubmit; preview token is stronger.

## R-08 — P/L chart window

**Decision**: Default query **`from=now-30d`** (calendar) in exchange / ledger time; chart component receives ordered `{t, pnl}` points.

**Rationale**: FR-006, SC-008.

## R-09 — Market mood + strategy metadata

**Decision**: BFF aggregates from **internal strategy / regime** modules when available; if not yet implemented, return **`degraded`** panel state with `available: false` and message per spec edge cases — **no fake numbers**.

**Rationale**: FR-007/008; avoid lying UI.

## R-10 — Health: uptime and RAM

**Decision**: BFF reports **process** uptime (monotonic or start timestamp) and **`runtime.MemStats`** (e.g. `Alloc` / `Sys`) with units; test vs live from existing env / deribit base URL config (`testnet` vs `mainnet`).

**Rationale**: FR-005; keep off hot path, cache ≤ few seconds if needed.
