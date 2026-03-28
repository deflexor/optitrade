---
work_package_id: WP06
title: Position detail legs and Greeks
lane: planned
dependencies: [WP05]
subtasks:
- T030
- T031
- T032
- T033
- T034
phase: Phase 3 - Detail
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
- FR-009
---

# Work Package Prompt: WP06 -- Position detail legs and Greeks

## Objectives & Success Criteria

- `GET /api/v1/positions/{positionId}` returns `PositionDetail` with legs, liquidity notes, metrics map, greeks when known (FR-009).
- React page shows leg table and summary metrics; clicking list row navigates here.

## Context & Constraints

- **Implement command**: `spec-kitty implement WP06 --base WP05`
- Edge case: incomplete multi-leg -> still list partial legs with state flags.

## Subtasks & Detailed Guidance

### T030 -- Handler and routing

- Path param validation; 404 unknown id.
- Join legs from execution state or exchange positions API cache.

### T031 -- Leg DTO

- Map instrument, side, size as strings; liquidity from spread or depth heuristic (document).

### T032 -- Greeks

- Per-leg optional; portfolio greeks in `metrics` if computed in risk engine.

### T033 -- React `/positions/:id`

- Layout: header summary, legs table, metrics cards, actions row for Close / Rebalance buttons (disabled until WP07/08).

### T034 -- Navigation

- List row click -> detail; browser back returns to list.

## Test Strategy

- Handler test with golden JSON fixture under `src/internal/dashboard/testdata/`.

## Risks

- **Large payload**: acceptable for defined-risk small leg count.

## Review Feedback

*[Empty.]*

## Activity Log

- 2026-03-28T11:05:00Z -- system -- lane=planned -- Prompt created via /spec-kitty.tasks
