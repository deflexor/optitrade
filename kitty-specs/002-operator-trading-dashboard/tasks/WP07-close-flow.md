---
work_package_id: WP07
title: Close preview and close execution
lane: planned
dependencies: [WP06]
subtasks:
- T035
- T036
- T037
- T038
phase: Phase 4 - Close
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
- FR-010
- FR-012
- FR-013
---

# Work Package Prompt: WP07 -- Close preview and close execution

## Objectives & Success Criteria

- Close preview shows `estimated_pnl_usd`, `quote_as_of`, `recommendation` (close_now | wait | neutral), `rationale` (FR-010).
- Confirm only after user sees fresh preview (SC-005); block if quote age overlaps dashboard staleness rules for trading actions.
- Close POST returns 202/200 per design; errors explicit (FR-013); UI reconciles list.

## Context & Constraints

- **Implement command**: `spec-kitty implement WP07 --base WP06`
- Wire to `internal/execution` when available; else stub with clear `501` until bot ready.

## Subtasks & Detailed Guidance

### T035 -- close-preview endpoint

- Compute mark-to-close from risk/market cache; set `quote_as_of` from newest quote timestamp.
- Map engine decision to recommendation enum.

### T036 -- close endpoint

- Require JSON `{ "confirm": true }`; verify session; idempotency key optional via header.
- On conflict (position already flat) return 409 with message.

### T037 -- React modal

- Flow: user clicks Close -> fetch preview -> display -> Confirm calls close.
- Disable confirm if `Date.now()-quote_as_of > threshold` (5s or separate quote TTL documented).

### T038 -- Post-close UX

- On success: invalidate Zustand caches, refetch lists; on error show banner with exchange text.

## Test Strategy

- `httptest` preview happy path; table tests for 409.

## Risks

- **Partial close**: document behavior when multi-leg not fillable.

## Review Feedback

*[Empty.]*

## Activity Log

- 2026-03-28T11:05:00Z -- system -- lane=planned -- Prompt created via /spec-kitty.tasks
