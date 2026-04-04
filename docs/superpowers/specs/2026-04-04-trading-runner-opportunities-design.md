# Trading runner, bot modes, and Opportunities UI

**Status:** Draft — pending product review  
**Scope:** `src/internal/dashboard` (BFF + SQLite), **`internal/strategy` maturity** and a new **Opportunities selector** (runner core), new runner package(s) under `src/internal/`, `web/` (routes, header, Opportunities page, settings)  
**Date:** 2026-04-04

## Summary

Add a **per-operator trading runner** started as **goroutine(s) inside the dashboard process**. The runner’s **core** is a real **Opportunities selector**: mature the **strategy layer** so it can **expand** the traded universe (instruments × allowed structures), **price** candidates with venue data, **score / rank** them, and **veto** via existing **cost** and **risk** modules. The runner keeps the ranked list **fresh for the UI**, respects **risk limits** (e.g. max loss as % of equity), and **auto-enters** only when **bot mode is `auto`**. Operators control **bot mode** (`manual` / `auto` / `paused`) from the **header**; **admins** can **disable** an account (`account_status`) so **no runner** starts. Replace the current **exchange positions list** with an **Opportunities** page that shows **live candidates** and **bot-managed positions** (open / opening / partial) with **Open**, **Cancel**, and **Close** actions. When **paused**, the Opportunities page shows a **Paused** banner and **no live list** (empty body + short copy to resume).

## Brainstorming decisions

| Topic | Choice |
|--------|--------|
| Manual vs auto vs paused (placement) | **Header** controls for bot mode (alongside Settings / user). |
| Runner topology | **Goroutine(s) inside `dashboard`** — not a separate binary for v1. |
| Multi-user | **One runner per operator** when eligible; goroutines are cheap; rate limits are the practical cap. |
| Runner eligibility | Start only if **valid exchange keys** AND **`account_status = active`** AND **`bot_mode` ∈ {`manual`, `auto`}**. |
| Stopped | **No runner** if admin **`account_status ≠ active`** OR user **`bot_mode = paused`**. |
| Paused UX | Opportunities page: **Paused banner**; **empty** list; short **“Resume to scan the market”** (or equivalent) copy. |
| Legacy Positions page | **Remove** the current **`/positions`** list (and linked flows) in favor of **Opportunities**; see **Positions parity** below. |
| Runner core | **Mature `internal/strategy` + implement Opportunities selector** (not a mocked list); see dedicated section below. |

## Goals

1. **Discovery loop** — The **selector** periodically evaluates a **generated candidate set** from the options universe (plus playbook / regime), produces **ranked opportunities** with: strategy name, legs with **bid/ask** (live), **greeks**, **max profit / max loss**, **recommendation** (open / pass) and **short rationale**.
2. **Realtime UI** — Opportunities list **updates continuously** while the runner is active (`manual` or `auto`).
3. **Execution** — **Open** submits the multi-leg plan; **Cancel** while **Opening**; **Close** when **Active** or **Partial**; status machine: default → **Opening** → **Active** | **Partial**.
4. **Risk** — Per-strategy **max loss** must not exceed operator-configured **fraction of equity** (default e.g. **10%**, editable on **Settings**).
5. **Modes** — **`manual`**: runner **never** auto-opens new entries; **`auto`**: runner **may** open when checks pass; **`paused`**: **no runner**, banner-only UX on Opportunities.

## Non-goals (v1)

- Separate worker process or horizontal scaling story beyond “one dashboard replica.”
- Full **admin UI** for `account_status` (may be **DB / migration default** + documented SQL until an admin surface exists).
- **Light theme** changes (follow existing dashboard tokens).

## Architecture

### Process model

- **`optitrade dashboard`** starts an internal **`RunnerManager`** after HTTP server and DB are ready.
- **RunnerManager** maintains a map **username → runner context** (cancel func + metadata).
- On **startup**, enumerate operators with **eligible** rows (keys + `account_status` + `bot_mode`).
- On **settings change** (PUT `/settings` or future admin mutation), **reconcile**: start/stop/restart affected runners.

### Concurrency and safety

- **Per-user `recover`** on the runner’s top loop; log and continue or backoff on panic.
- **Context cancellation** per user on stop/disable/pause.
- **Timeouts** on all exchange RPC from the runner; avoid blocking HTTP handlers (runner never holds dashboard-global locks across I/O).
- **Isolation** — Runner state is **per username**; no shared mutable global for trading decisions.

### Data flow (conceptual)

```text
Runner (per user) → in-memory snapshot + optional SQLite rows for opportunity state
                 → BFF exposes GET (snapshot) and/or SSE for UI
HTTP handlers     → POST open / cancel / close delegate to same execution path as runner auto-open
```

**Alternatives considered:** (1) **Separate binary** polling SQLite — rejected for v1 operational simplicity. (2) **WebSocket** for realtime — possible later. **Preference for v1:** **SSE** from BFF for push updates; **short-interval polling** acceptable as a first milestone if SSE is deferred.

## Strategy layer maturity and Opportunities selector (runner core)

Today, **`internal/strategy`** is strong at **building** defined-risk legs from **given** expiry/strikes and at **playbook allow-lists**, but **production selection** is not implemented: `BuildLegsForStructure` uses **hardcoded** example strikes/expiry for certification-style wiring, and **`cost.ScoreCandidate`** expects an **already supplied** `ExpectedEdge`. **Closing that gap is in scope** for this feature.

### Responsibilities of the Opportunities selector

1. **Universe** — For each operator (currencies, venue), fetch or cache **listed options** (e.g. Deribit `get_instruments` filtered by kind/expiry window). Bound work: **max expiries ahead**, **strikes near spot** (configurable bands), **liquidity pre-filters** (min OI/volume optional, v2).
2. **Structure expansion** — For each **regime-resolved** allowed structure from policy (`AllowedStructures`), **expand** to concrete `[]LegSpec` using **parameterized** builders (not fixed `27JUN26` demos). Strategy package grows **pure functions**: e.g. “put credit vertical: base, expiry, short strike, width → legs.”
3. **Market data** — Batch or throttle **order books** (and **greeks** from venue if available) per leg; reuse **`market.MarketSnapshot`** / quality flags where applicable.
4. **Economics** — Derive **max loss / max profit** (defined-risk templates), **mid / bid / ask** aggregate for the structure, and an **expected edge** (or score) model documented in the implementation plan — e.g. credit received − estimated fair − haircuts, or a simpler v1 proxy (credit / width) **as long as** it feeds **`cost.ScoreCandidate`** honestly (no fake constant edge).
5. **Ranking** — Sort/filter by a **documented primary key** (e.g. edge after costs, then liquidity). Cap **top N** opportunities exposed to UI and auto mode to control load.
6. **Gates** — Run **`cost.ScoreCandidate`** and **`risk.Engine.Check`** (and max-loss-% of equity) before a row is **recommended** or **auto-opened**; attach **human-readable rationale** from gate results / veto reasons.
7. **Tests** — **Table-driven** tests for expansion (known chain → expected legs), scoring invariants, and selector integration with **stub books** (no live venue in CI).

### Code organization (suggested)

- **`internal/strategy`** — Parameterized **structure builders** + validation; **remove or quarantine** hardcoded `BuildLegsForStructure` from the **live selector path** (keep for tests/certification if needed behind a test-only or explicit “example” entrypoint).
- **`internal/opportunities`** (or `internal/runner/selector`) — **Orchestration**: universe query, loop, ranking, snapshot type consumed by runner + API. Depends on `strategy`, `cost`, `risk`, `regime`, `market`, `deribit`.
- **Runner** — Ticks selector at operator cadence; merges **user-initiated** state (Opening / Active) with **fresh** candidate rows.

### Phasing inside the project

- **Milestone 1** — One underlying (e.g. BTC), **one structure** (e.g. put credit spread), **strike grid** from policy width + spot band, **ranking v1**, full **pipeline to UI** with manual mode.
- **Milestone 2** — Additional structures from playbook, second currency, tighter **rate-limit** batching, **auto** mode.
- **Milestone 3** — Richer edge model, optional liquidity/OI filters, performance tuning.

Exact milestone boundaries belong in the **writing-plans** breakdown.

## State model

### `account_status` (admin)

- **`active`** (default for new rows) — runner **may** run if other conditions hold.
- **`disabled`** — **never** start runner; **stop** if running. Operator may still use dashboard for configuration unless product forbids (default: **allow login**).

### `bot_mode` (user)

| Mode | Runner | Auto new entries | UI |
|------|--------|------------------|-----|
| `manual` | On | No | Live opportunities; **Open** is user-driven |
| `auto` | On | Yes (if risk + gates pass) | Same list; highlight auto actions in audit/logs |
| `paused` | Off | No | **Paused** banner; empty body + resume hint |

### Opportunity / position row (UI + server)

- **Candidate** — Runner-produced, not yet ordered; **Open** visible.
- **Opening** — Orders in flight; **Open** hidden/disabled; **Cancel** visible.
- **Active** — All legs filled as planned; show **live P/L**-oriented fields; **Close** visible.
- **Partial** — Some legs failed (undesired); **Close** still available; surface **warning** copy.

**Identity** — Stable **server-assigned id** per opportunity row (not only exchange instrument id), so one strategy maps to one row through its lifecycle.

## API surface (v1 sketch)

Exact paths and payloads belong in the implementation plan; shape should include:

- **`GET /trading/status`** — `{ account_status, bot_mode, runner_running: bool }` for header + Opportunities shell.
- **`PUT /trading/mode`** — body `{ "bot_mode": "manual"|"auto"|"paused" }` (authenticated operator only).
- **`GET /opportunities`** — JSON snapshot for initial load.
- **`GET /opportunities/stream`** (SSE) **or** polling **`GET /opportunities?since=…`** — push updates while runner active.
- **`POST /opportunities/{id}/open`**, **`POST /opportunities/{id}/cancel`**, **`POST /opportunities/{id}/close`** — idempotent where possible; align with existing `execution` / exchange error mapping.

**Auth** — Same session cookie as today; all routes under authenticated BFF.

## Web app

### Header

- **Bot mode** control: `manual` / `auto` / `paused` (segmented control or select), **persisted** via API.
- Optional compact **indicator**: runner off when paused/disabled (muted vs live).

### Routing

- **New primary route** **`/opportunities`** (display title: **Opportunities**).
- **Remove** **`/positions`** and **`/positions/:id`** as the **exchange-only** views; update **Overview** link and **e2e** specs accordingly.

### Opportunities page

- Table or cards: strategy name, legs + bid/ask, greeks, max P/L, recommendation + explanation, status, actions.
- **Paused**: full-width **Alert**/banner **Paused**; **no rows**; supporting text: resume scanning when switching mode away from paused.
- **Disabled account** (if user can still open page): clear **not authorized to trade** / contact admin messaging.

### Settings

- Add **risk**: **max loss as % of equity** (default **10**), validated range (e.g. 1–100 or 1–50 — pick in implementation).
- Persist in **operator settings** (new column(s) or encrypted blob extension + migration).

## Positions parity

Today **`/positions`** shows **raw exchange** open/closed rows. After removal:

- **Bot-tracked** opportunities cover **strategies opened via this UI** (and auto-open in `auto` mode).
- **Orphan exchange positions** (manual venue trades, partial failures, external activity) are **not** shown on Opportunities v1 unless the runner **reconciles** them into rows — **explicitly out of scope** unless a small **“Unmanaged positions”** subsection is added later.
- **Close** flows for **Active** opportunities should reuse or extend **`internal/dashboard`** close/preview patterns where applicable.

## Risk and strategy integration

- **Equity** source: same exchange account summary path as overview (per operator).
- **Max loss** per candidate compared to **equity × (max_loss_pct / 100)** before **Open** or auto-entry.
- The **Opportunities selector** produces **`risk.CandidateRisk`** (and correlation IDs) for each row and calls **`risk.Engine.Check`** where pre-trade gates apply; **`cost.ScoreCandidate`** runs on **computed** expected edge and leg books — not placeholder strings.

## Testing

- **Go** — Unit tests for runner eligibility, mode transitions, and state machine; handler tests for new endpoints (mock exchange).
- **Web** — Component tests optional; **Playwright** updates for new route titles, header mode control, paused banner; remove obsolete **Positions** assertions.

## Follow-on

After this spec is approved in the repo, use **writing-plans** for phased implementation. **Order the plan** so **selector + strategy maturity** land **before** or **in tight parallel with** UI polish: e.g. parameterized strategy builders + selector Milestone 1 (stub runner loop) → BFF snapshot API → Opportunities page → RunnerManager + persistence + SSE/poll → execution hooks → auto mode → retire old **`/positions`** routes → e2e refresh.
