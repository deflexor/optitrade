---
work_package_id: WP08
title: Rebalance preview and execution
lane: "for_review"
dependencies: [WP06]
base_branch: 002-operator-trading-dashboard-WP06
base_commit: 866a248244e064fcd8ffe6dcbbefbd37d8dc3f61
created_at: '2026-03-28T12:40:54.617192+00:00'
subtasks:
- T039
- T040
- T041
- T042
phase: Phase 4 - Rebalance
assignee: ''
agent: "cursor"
shell_pid: "84200"
review_status: "has_feedback"
reviewed_by: "Dmitriy Knyazev"
review_feedback_file: "/tmp/spec-kitty-review-feedback-WP08.md"
history:
- timestamp: '2026-03-28T11:05:00Z'
  lane: planned
  agent: system
  shell_pid: ''
  action: Prompt generated via /spec-kitty.tasks
requirement_refs:
- FR-011
- FR-012
- FR-013
---

# Work Package Prompt: WP08 -- Rebalance preview and execution

## Objectives & Success Criteria

- Preview returns ranked `suggestions` with title, detail, expected_effect (FR-011).
- Execute path requires confirmation (FR-012); 202 response acceptable for async work.

## Context & Constraints

- **Implement command**: `spec-kitty implement WP08 --base WP06`
- If engine has no rebalance planner, return `{ "suggestions": [] }` and UI explains "no suggestions" (structural compliance).

## Subtasks & Detailed Guidance

### T039 -- rebalance-preview

- Hook into strategy/risk module placeholder returning 0-N suggestions.

### T040 -- rebalance POST

- Validate position still open; delegate to execution; surface errors.

### T041 -- Modal UI

- Numbered list; expand row for detail; confirm button.

### T042 -- Refresh

- After 202, poll position detail until state change or timeout with message.

## Test Strategy

- Handler tests with canned suggestions JSON.

## Risks

- **Misleading outcomes**: copy uses "possible" wording per spec Out of scope on guarantees.

## Review Feedback

**Reviewed by**: Dmitriy Knyazev
**Status**: ❌ Changes Requested
**Date**: 2026-03-28
**Feedback file**: `/tmp/spec-kitty-review-feedback-WP08.md`

**Issue 1 (contract)**: `GET /api/v1/positions/{positionId}/rebalance-preview` does not match the dashboard API contract or task checklist. In `kitty-specs/002-operator-trading-dashboard/contracts/dashboard-api.openapi.yaml`, `/positions/{positionId}/rebalance-preview` is defined as **POST** (returns 200 + `RebalancePreview`). `kitty-specs/002-operator-trading-dashboard/tasks.md` T039 also lists **POST** for rebalance-preview.

**Fix**: Register `POST /positions/{positionId}/rebalance-preview` in `server.go`, accept POST in `handleRebalancePreview` (body may be empty JSON `{}`), update `rebalance_handlers_test.go` to call POST, and change `PositionDetailPage.tsx` to use `api.post(..., {})` instead of `api.get` for preview.

**Dependency check**: WP08 depends on **WP06**; topology shows WP06 **done** and this branch stacks on `002-operator-trading-dashboard-WP06` -- OK.

**Dependent check**: **WP09** depends on WP08 and is **planned** (`branch: null`). After this WP is fixed and re-reviewed, anyone starting WP09 should branch from the updated WP08 tip.

**Rebase warning**: If WP09 worktrees exist before the POST fix lands, rebase them after WP08 updates, for example: `cd /home/dfr/optitrade/.worktrees/002-operator-trading-dashboard-WP09 && git fetch && git rebase 002-operator-trading-dashboard-WP08`

**Note (non-blocking)**: OpenAPI for `POST .../rebalance` does not yet document a `confirm` body; the implementation correctly requires confirmation for FR-012. Consider a follow-up to extend the OpenAPI schema to match (optional cleanup).


## Activity Log

- 2026-03-28T11:05:00Z -- system -- lane=planned -- Prompt created via /spec-kitty.tasks
- 2026-03-28T12:40:54Z – cursor – shell_pid=83071 – lane=doing – Assigned agent via workflow command
- 2026-03-28T12:44:27Z – cursor – shell_pid=83071 – lane=for_review – Ready for review: rebalance GET preview, confirmed POST 202, modal UI with poll
- 2026-03-28T12:44:41Z – cursor – shell_pid=84200 – lane=doing – Started review via workflow command
- 2026-03-28T12:44:54Z – cursor – shell_pid=84200 – lane=planned – Moved to planned
- 2026-03-28T12:48:14Z – cursor – shell_pid=84200 – lane=for_review – POST rebalance-preview + Allow header; ready for re-review
