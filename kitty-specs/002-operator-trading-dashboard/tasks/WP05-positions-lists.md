---
work_package_id: WP05
title: Open and closed positions list API and UI
lane: "for_review"
dependencies: [WP03]
base_branch: 002-operator-trading-dashboard-WP03
base_commit: 8e3f17ed1c7ece4fc5ca9857f62fcd4f3a6e4710
created_at: '2026-03-28T12:05:27.446381+00:00'
subtasks:
- T025
- T026
- T027
- T028
- T029
phase: Phase 3 - Lists
assignee: ''
agent: "cursor"
shell_pid: "77003"
review_status: ''
reviewed_by: ''
history:
- timestamp: '2026-03-28T11:05:00Z'
  lane: planned
  agent: system
  shell_pid: ''
  action: Prompt generated via /spec-kitty.tasks
requirement_refs:
- FR-006
- FR-007
- FR-008
---

# Work Package Prompt: WP05 -- Open and closed positions list API and UI

## Objectives & Success Criteria

- Open positions list with strategy name, expected P/L at open, win rate (FR-006).
- Ten newest closed positions with USD and percent P/L (FR-007).
- Open rows expose control to start close flow (FR-008); wire button to route or modal hook for WP07.

## Context & Constraints

- **Implement command**: `spec-kitty implement WP05 --base WP03`
- Map from `internal/state` position snapshots or execution map; align IDs with Deribit.

## Subtasks & Detailed Guidance

### T025 -- GET /positions/open

- Transform internal model to OpenAPI `PositionSummaryOpen`.
- Stable sort: by `position_id` or opened time desc (document).

### T026 -- GET /positions/closed

- Query limit max 10 enforced server-side; `ORDER BY closed_at DESC`.

### T027 -- Strategy stats

- If win rate unavailable, return `null` and UI shows em dash; document follow-up. Expected P/L must match bot definition when wired.

### T028 -- React tables

- Accessible table headers; mobile-friendly stacked cards optional.
- Close button per FR-008.

### T029 -- Empty states

- Zero open / zero closed copy per spec edge cases.

## Test Strategy

- `go test` with fake repo implementing small interface.

## Risks

- **Reconciliation lag**: show `stale` from snapshot if positions older than threshold.

## Review Feedback

*[Empty.]*

## Activity Log

- 2026-03-28T11:05:00Z -- system -- lane=planned -- Prompt created via /spec-kitty.tasks
- 2026-03-28T12:05:27Z – cursor – shell_pid=77003 – lane=doing – Assigned agent via workflow command
- 2026-03-28T12:07:58Z – cursor – shell_pid=77003 – lane=for_review – Ready for review: GET /positions/open and /closed, operator_position migration + sqlite repo, React tables with empty states and close-preview route hook
