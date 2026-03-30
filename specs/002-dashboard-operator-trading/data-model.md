# Data model: Operator dashboard

**Spec**: [`spec.md`](./spec.md)  
**Branch**: `002-dashboard-operator-trading`

Logical entities (BFF-centric). Field names are indicative; exact JSON in `contracts/openapi.yaml`.

## Config: `DashboardAuthFile` (JSON, schema-validated)

| Field | Type | Notes |
|-------|------|--------|
| `version` | string | Schema version. |
| `users[]` | array | Non-empty for enabled auth. |
| `users[].username` | string | Unique, case-sensitive rule TBD (recommend ASCII lower canonical). |
| `users[].password_hash` | string | Bcrypt / argon2id encoded hash string. |

**Validation**: Unique usernames; no plaintext passwords; file readable only by service OS user.

**Reload**: On SIGHUP or restart; session validation must re-check username still in loaded allowlist.

## Session (persistent store)

| Field | Type | Notes |
|-------|------|--------|
| `id` | int64 | Primary key (optional if using random session id only). |
| `username` | string | Matches config allowlist entry. |
| `token_hash` | string | Hash of cookie value (never store raw token). |
| `created_at` | unix ms | Audit. |
| `user_agent` | string optional | Debug only; do not log secrets. |

**Removed / unused**: `expires_at` for **authorization** (may remain NULL or sentinel for schema migration compatibility).

**Transitions**: Created on successful login; deleted on logout; **invalid** if username ∉ current config.

## System health snapshot (DTO)

| Field | Type |
|-------|------|
| `uptime_seconds` | int64 or string |
| `memory_heap_alloc_bytes` | int64 |
| `memory_sys_bytes` | int64 (optional) |
| `collected_at` | RFC3339 |

## Trading connection profile (DTO)

| Field | Type |
|-------|------|
| `mode` | enum: `test` \| `live` |
| `exchange_reachable` | bool |
| `detail` | string optional |

## Account snapshot (DTO)

| Field | Type |
|-------|------|
| `currency` | string e.g. `BTC`, `USDC` |
| `equity` / `balance` | decimal **string** |
| `available_funds` | decimal string optional |

## P/L series (DTO)

| Field | Type |
|-------|------|
| `points[]` | `{ "t": RFC3339, "pnl_quote": string }` |
| `window` | `{ "from": date, "to": date }` |

## Market mood (DTO)

| Field | Type |
|-------|------|
| `label` | string |
| `score` | number optional |
| `explanation` | string optional |
| `available` | bool |

## Strategy metadata (DTO)

| Field | Type |
|-------|------|
| `expected_pnl` | object with horizon + decimal strings |
| `win_rate` | string or null |
| `win_rate_defined` | bool |

## Position (list + detail)

| Field | Type |
|-------|------|
| `id` | string (stable BFF/exchange id) |
| `instrument_summary` | string |
| `open` | bool |
| `legs[]` | leg objects |
| `metrics` | object (liquidity, spreads, etc.) |
| `greeks` | nullable object |
| `usd_pnl` / `quote_pnl` | decimal string |

## Closed position row (DTO)

| Field | Type |
|-------|------|
| `closed_at` | RFC3339 |
| `realized_pnl_usd` | string |
| `realized_pnl_pct` | string nullable |
| `percent_basis` | string enum key |
| `percent_basis_label` | string |

## Close estimate (preview)

| Field | Type |
|-------|------|
| `estimated_exit_pnl` | string |
| `wait_vs_close_guidance` | string |
| `assumptions[]` | string optional |
| `preview_token` | string (if using two-step confirm) |

## Rebalance proposal (preview)

| Field | Type |
|-------|------|
| `suggested_adjustments[]` | structured legs/orders |
| `projected_outcome` | human + numeric summary |
| `preview_token` | string optional |

## Relationships

- **Session** → **username** → must exist in **DashboardAuthFile.users**.
- **Positions / balance** → sourced from **Deribit** (and internal ledger if added later), not from auth file.
