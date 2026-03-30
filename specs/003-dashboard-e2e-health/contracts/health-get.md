# Contract: `GET /api/v1/health`

Public (no session). Implemented in `src/internal/dashboard/health.go`.

## 200 response (success body, illustrative)

```json
{
  "health": {
    "uptime_seconds": 3600,
    "memory_heap_alloc_bytes": 12345678,
    "memory_sys_bytes": 45678912,
    "collected_at": "2026-03-31T12:00:00Z"
  },
  "trading": {
    "mode": "test",
    "exchange_reachable": false,
    "detail": "exchange_not_configured"
  }
}
```

## Display contract (this feature)

- SPA SHOULD render `uptime_seconds` and `memory_heap_alloc_bytes` in **human-meaningful** units (spec **FR-003**).
- Raw numeric fields remain in JSON for tests and diagnostics.
