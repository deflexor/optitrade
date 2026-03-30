# Data model: presentation & API payloads (003-dashboard-e2e-health)

**Scope**: No new databases. Documents **existing** JSON shapes relevant to formatting, errors, and tests.

## `GET /api/v1/health` (200)

| Field | Type | Notes |
|-------|------|--------|
| `health.uptime_seconds` | number (int) | Monotonic since process start; UI formats to human duration. |
| `health.memory_heap_alloc_bytes` | number (int) | `runtime.MemStats.Alloc`; UI formats to MB/GiB-style string + unit. |
| `health.memory_sys_bytes` | number (int) | Optional future UI; not required for this feature. |
| `health.collected_at` | string | RFC3339 UTC. |
| `trading.mode` | string | `test` \| `live` (from env/base URL heuristic). |
| `trading.exchange_reachable` | boolean | RPC check when `xchg != nil`. |
| `trading.detail` | string | e.g. `exchange_not_configured`, `exchange_unreachable`. |

## `GET /api/v1/overview` (200) — `market_mood`

| Field | Type | Notes |
|-------|------|--------|
| `market_mood.available` | boolean | When false, UI shows explanation copy (must be operator-safe). |
| `market_mood.label` | string | Shown when `available`. |
| `market_mood.score` | number \| null | Optional. |
| `market_mood.explanation` | string | Shown when `!available`; **must not** contain dev-only wiring notes after this feature. |

## Positions list payloads (200) — summary

Open: `{ items, next_cursor, total_count }`  
Closed: `{ items, truncated }`  
Row fields per existing handlers; pagination **limit 25** preserved.

## API error envelope (4xx/5xx)

Stable object for BFF errors:

```json
{
  "error": "<machine_code>",
  "message": "<operator-safe short description>"
}
```

Relevant codes for this feature:

| HTTP | `error` | Typical meaning |
|------|---------|-----------------|
| 503 | `exchange_unavailable` | No exchange client configured. |
| 502 | `exchange_error` | Client present but RPC failed. |

## UI state (conceptual)

**PositionsPage**: Should distinguish `loading` / `ready` / `error` per **Open** and **Closed** (or unified error that clears loading spinners). Avoid `open === null && err !== null` still showing “Loading…”.

**HealthPanel**: `loading` \| `error` \| `ready(formattedDisplay)`.
