---
work_package_id: WP03
title: Auth middleware and summary snapshot API
lane: planned
dependencies: [WP02]
subtasks:
- T014
- T015
- T016
- T017
- T018
- T019
- T020
phase: Phase 2 - Snapshot
assignee: ''
agent: ''
shell_pid: ''
review_status: ''
reviewed_by: ''
history:
- timestamp: '2026-03-28T11:05:00Z'
  lane: planned
  agent: system
  shell_pid: ''
  action: Prompt generated via /spec-kitty.tasks
requirement_refs:
- FR-001
- FR-002
- FR-003
- FR-005
- FR-017
- FR-019
---

# Work Package Prompt: WP03 -- Auth middleware and summary snapshot API

## Objectives & Success Criteria

- All dashboard APIs except `/api/v1/auth/*` require valid session (FR-017).
- `GET /api/v1/snapshot` returns OpenAPI `Snapshot` shape: equity, `deribit_environment` test|live, `deribit_connected`, uptime, RSS, `classification`, `regime_label`, `market_mood_label` (all three strings aligned), `snapshot_utc`, `stale` bool (FR-019).
- Unit tests prove `stale=true` when synthetic clock exceeds 5s since snapshot assembly start or cached data timestamp.

## Context & Constraints

- Implement command: `spec-kitty implement WP03 --base WP02`
- Reuse bot internals: `internal/regime`, Deribit client env, mem stats from `runtime`.

## Subtasks & Detailed Guidance

### T014 -- Auth middleware

- Extract session cookie, lookup DB, set `context.Context` value `dashboard.UserID`.
- On failure write 401 JSON `{"error":"unauthorized"}` consistent with OpenAPI.
- Apply to sub-router mounting `/api/v1/snapshot`, `/positions`, etc.

### T015 -- Process metrics

- `runtime.ReadMemStats` for RSS approximation (document field as bytes).
- Uptime: `time.Since(startTime)` from process `var startTime` in `main` or dashboard server start.

### T016 -- Equity

- Call existing account summary if available; else return `"0"` with `stale=true` and log WARNING once. Mark `TODO` linking to bot integration task.

### T017 -- Regime triple

- Single canonical string `c` from regime evaluator; set `classification=c`, `regime_label=c`, `market_mood_label=c` (FR-005, SC-009).

### T018 -- Stale rule

- `snapshot_utc` = `time.Now()` when bundle assembled (server UTC).
- If any upstream segment older than 5s or disconnected, set `stale=true`.
- Inject `type Clock interface { Now() time.Time }` for tests.

### T019 -- Route

- Register `GET /api/v1/snapshot` on protected mux.

### T020 -- Tests

- Fake clock: data fresh at T0, advance 6s, expect `stale=true`.
- Test regime strings equal always.

## Test Strategy

- `go test ./src/internal/dashboard/...`

## Risks

- **Clock skew**: document server-only `snapshot_utc`; optional `server_time` header later.

## Review Feedback

*[Empty.]*

## Activity Log

- 2026-03-28T11:05:00Z -- system -- lane=planned -- Prompt created via /spec-kitty.tasks

## Markdown Formatting

Wrap XML tags in backticks.
