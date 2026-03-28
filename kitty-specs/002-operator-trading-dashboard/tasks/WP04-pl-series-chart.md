---
work_package_id: WP04
title: P/L series API and chart UI
lane: "done"
dependencies: [WP03]
base_branch: 002-operator-trading-dashboard-WP03
base_commit: 8e3f17ed1c7ece4fc5ca9857f62fcd4f3a6e4710
created_at: '2026-03-28T12:00:21.276342+00:00'
subtasks:
- T021
- T022
- T023
- T024
phase: Phase 2 - P/L
assignee: ''
agent: "cursor"
shell_pid: "76273"
review_status: "approved"
reviewed_by: "Dmitriy Knyazev"
history:
- timestamp: '2026-03-28T11:05:00Z'
  lane: planned
  agent: system
  shell_pid: ''
  action: Prompt generated via /spec-kitty.tasks
requirement_refs:
- FR-004
---

# Work Package Prompt: WP04 -- P/L series API and chart UI

## Objectives & Success Criteria

- Default API range **30d** (FR-004, SC-007); response documents actual span if shorter history.
- React chart reflects default on first load; operator sees labeled axes and active range.

## Context & Constraints

- **Implement command**: `spec-kitty implement WP04 --base WP03`
- OpenAPI: `pl-series` path and `PLSeriesResponse` schema.

## Subtasks & Detailed Guidance

### T021 -- Storage or derived series

- Add table `dashboard_pnl_point (captured_at INTEGER, pnl_usd TEXT)` or reuse bot PnL if exists.
- For MVP, seed from periodic writer in bot or manual migration with zero rows acceptable if UI handles empty.

### T022 -- Handler

- `GET /api/v1/pl-series?range=30d|7d|24h` validated; default 30d when query missing.
- Return `{ "range": "30d", "range_effective_from": "...", "points": [...] }` extend OpenAPI if needed via README note.

### T023 -- Chart component

- Fetch with axios on interval (2-3s) or on mount + refresh button; avoid chart library heavier than needed (e.g. `recharts` or raw SVG polyline).
- Zustand slice `usePlSeriesStore`.

### T024 -- Short history UX

- If `points.length` implies <30d, show banner "Showing N days of available history" per spec.

## Test Strategy

- Go: handler tests with sqlite fixture points.
- Vitest: component renders empty state.

## Risks

- **Empty series**: show flat line or message, not error.

## Review Feedback

*[Empty.]*

## Activity Log

- 2026-03-28T11:05:00Z -- system -- lane=planned -- Prompt created via /spec-kitty.tasks
- 2026-03-28T12:00:21Z – cursor – shell_pid=74094 – lane=doing – Assigned agent via workflow command
- 2026-03-28T12:03:32Z – cursor – shell_pid=74094 – lane=for_review – Ready for review: P/L migration, GET /api/v1/pl-series, Zustand + SVG chart, Vitest empty state
- 2026-03-28T12:04:23Z – cursor – shell_pid=76273 – lane=doing – Started review via workflow command
- 2026-03-28T12:04:41Z – cursor – shell_pid=76273 – lane=done – Review passed: migration 0004, parameterized pl-series handler + tests, Zustand/SVG chart + Vitest empty state, FR-004 defaults and short-history UX; WP03 stacked base noted for merge
