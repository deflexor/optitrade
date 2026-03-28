# Data Model: Autonomous Deribit Options Robot

**Feature**: `001-autonomous-deribit-options-robot`  
**Date**: 2026-03-28  
**Storage**: SQLite (execution service)

## Conventions

- Timestamps: UTC, RFC3339 in API and logs; INTEGER Unix ms acceptable internally if consistent.
- Money: store as strings of decimal OR integer **atomic units** per Deribit docs; never floats for balances.
- IDs: exchange `order_id`, `trade_id`, `position_id` as text; internal UUID for audit correlation.

## Entities

### instrument

Exchange-tradable option contract.

| Field | Type | Notes |
|-------|------|--------|
| instrument_name | TEXT PK | Deribit name |
| kind | TEXT | `option`, `future`, etc. |
| base | TEXT | `BTC`, `ETH` |
| expiry | TEXT | ISO date or Deribit expiry |
| strike | TEXT | Decimal as text |
| option_type | TEXT | `call`, `put` |
| tick_size | TEXT | |
| min_trade_amount | TEXT | |
| is_active | INTEGER | 0/1 |

### market_snapshot (optional hot cache, or derived only)

Point-in-time top of book + vol features.

| Field | Type | Notes |
|-------|------|--------|
| id | INTEGER PK | |
| instrument_name | TEXT FK | |
| captured_at | INTEGER | ms |
| bid_px | TEXT | |
| ask_px | TEXT | |
| bid_qty | TEXT | |
| ask_qty | TEXT | |
| mark_iv | TEXT | nullable |
| vol_index_features | TEXT | JSON blob |
| quality_flags | TEXT | JSON (stale_book, gap, etc.) |

### regime_state

Latest classification for audit trail.

| Field | Type | Notes |
|-------|------|--------|
| id | INTEGER PK | |
| effective_at | INTEGER | ms |
| label | TEXT | `low`, `normal`, `high` |
| classifier_version | TEXT | |
| inputs_digest | TEXT | hash of inputs for replay |

### trade_candidate

Proposed structure before risk gates.

| Field | Type | Notes |
|-------|------|--------|
| id | TEXT PK | UUID |
| created_at | INTEGER | |
| regime_label | TEXT | |
| playbook_id | TEXT | |
| structure_type | TEXT | e.g. `vertical`, `iron_condor` |
| legs_json | TEXT | JSON array of leg specs |
| expected_edge | TEXT | after costs |
| cost_breakdown_json | TEXT | fees, spread, slip |

### order_record

| Field | Type | Notes |
|-------|------|--------|
| internal_order_id | TEXT PK | client id |
| exchange_order_id | TEXT | nullable until ack |
| instrument_name | TEXT | |
| label | TEXT | strategy label |
| side | TEXT | |
| order_type | TEXT | limit, market, combo... |
| price | TEXT | nullable |
| amount | TEXT | |
| post_only | INTEGER | 0/1 |
| reduce_only | INTEGER | 0/1 |
| state | TEXT | new, open, filled, canceled, rejected |
| created_at | INTEGER | |
| updated_at | INTEGER | |
| candidate_id | TEXT FK nullable | |

### fill_record

| Field | Type | Notes |
|-------|------|--------|
| id | TEXT PK | |
| order_id | TEXT FK | internal |
| trade_id | TEXT | exchange |
| instrument_name | TEXT | |
| qty | TEXT | |
| price | TEXT | |
| fee | TEXT | |
| filled_at | INTEGER | |

### position_snapshot

Periodic or event-driven portfolio Greeks and limits usage.

| Field | Type | Notes |
|-------|------|--------|
| id | INTEGER PK | |
| captured_at | INTEGER | |
| net_delta | TEXT | |
| net_vega | TEXT | |
| net_gamma | TEXT | nullable |
| premium_at_risk | TEXT | |
| unrealized_pnl | TEXT | |
| raw_json | TEXT | optional full exchange payload |

### risk_policy

Versioned config row or single active row + history table (implementation choice).

| Field | Type | Notes |
|-------|------|--------|
| id | INTEGER PK | |
| version | TEXT | |
| policy_json | TEXT | validated by `contracts/config-policy.schema.json` |
| active | INTEGER | 0/1 |

### audit_decision

One row per pre-trade or anomaly decision.

| Field | Type | Notes |
|-------|------|--------|
| id | TEXT PK | |
| ts | INTEGER | |
| decision_type | TEXT | `veto_risk`, `veto_cost`, `submit`, `protective_mode`, etc. |
| candidate_id | TEXT nullable | |
| regime_label | TEXT | |
| cost_model_version | TEXT | |
| risk_gate_results | TEXT | JSON map gate->pass/fail |
| reason | TEXT | human-readable |
| correlation_id | TEXT | ties logs |

## State machine (session)

**States**: `running`, `paused`, `protective_flatten`, `frozen`  

**Transitions** (illustrative):

- `running` -> `protective_flatten` on feed loss, auth failure, book gap (FR-009)
- `protective_flatten` -> `frozen` if flatten completes or operator forces
- `*` -> `paused` on operator command

Persist `session_state` in SQLite or reload from exchange on startup; document single-writer discipline in execution module.

## Relationships

- `trade_candidate` 1:N `order_record`
- `order_record` 1:N `fill_record`
- `audit_decision` N:1 `trade_candidate` (optional)
