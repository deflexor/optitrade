# Data Model: Operator Trading Dashboard

**Feature**: `002-operator-trading-dashboard`  
**Date**: 2026-03-28  
**Storage**: SQLite (extends existing optitrade DB file; new tables prefixed `dashboard_`)

## Conventions

- Timestamps: UTC RFC3339 in JSON; INTEGER ms allowed internally if consistent with `001` patterns.
- Money: decimal strings; never float (matches `001` conventions).
- Session token: opaque random string; store only **hash** of token in DB.

## New tables

### dashboard_user

Registered operator identity (v1: no email).

| Field | Type | Notes |
|-------|------|--------|
| id | INTEGER PK | |
| username | TEXT UNIQUE | normalized lowercase for lookup |
| password_hash | TEXT | Argon2id or bcrypt encoding |
| created_at | INTEGER | ms |

### dashboard_session

| Field | Type | Notes |
|-------|------|--------|
| id | INTEGER PK | |
| user_id | INTEGER FK | -> dashboard_user |
| token_hash | TEXT | hash of opaque bearer/cookie value |
| created_at | INTEGER | ms |
| expires_at | INTEGER | ms |
| user_agent | TEXT | optional, truncated |

**Rules**: On logout or password change, delete or invalidate rows. On process restart, spec allows invalidation if documented (optional: wipe table on boot flag).

## Aggregates and payloads (logical, not all persisted)

### Snapshot bundle (API DTO)

Aligned with FR-019 and Dashboard snapshot entity.

| Field | Notes |
|-------|--------|
| snapshot_utc | RFC3339 from server when bundle assembled |
| process_uptime_sec | |
| process_rss_bytes | |
| deribit_connected | bool |
| deribit_environment | `test` \| `live` |
| equity_usd | decimal string |
| classification | single canonical label e.g. `low` `normal` `high` |
| regime_label | duplicate of classification for UI FR-005 |
| market_mood_label | duplicate of classification |
| stale | bool: true if age of source data exceeds 5 s rule |

### P/L series point

| Field | Notes |
|-------|--------|
| t | RFC3339 |
| pnl_usd | cumulative or interval; document in OpenAPI |

### Position summary (open or closed)

Maps to spec Key Entities; fields populated from bot state + strategy metadata.

| Field | Notes |
|-------|--------|
| position_id | stable string |
| strategy_name | |
| expected_pnl_usd_at_open | decimal string |
| win_rate_pct | 0-100, from backend stats |
| realized_pnl_usd | closed only |
| realized_pnl_pct | closed only |

### Position leg

| Field | Notes |
|-------|--------|
| instrument_name | |
| side | |
| size | |
| liquidity_notes | summary string or score |
| delta, gamma, vega, theta | optional decimals per leg |

### Close preview / rebalance suggestion

DTOs returned by preview endpoints; may be computed on demand, not stored long-term.

## Relationship to feature 001

- Reuses **regime_state** or live classifier output for `classification`.
- Reuses position and account data from exchange reconciliation layers as they exist in `src/` (exact joins defined during implementation).
- No duplicate source of truth for open positions: dashboard reads through bot services.
