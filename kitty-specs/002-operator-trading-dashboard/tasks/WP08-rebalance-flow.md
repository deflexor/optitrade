---
work_package_id: WP08
title: Rebalance preview and execution
lane: "doing"
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
review_status: ''
reviewed_by: ''
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

*[Empty.]*

## Activity Log

- 2026-03-28T11:05:00Z -- system -- lane=planned -- Prompt created via /spec-kitty.tasks
- 2026-03-28T12:40:54Z – cursor – shell_pid=83071 – lane=doing – Assigned agent via workflow command
- 2026-03-28T12:44:27Z – cursor – shell_pid=83071 – lane=for_review – Ready for review: rebalance GET preview, confirmed POST 202, modal UI with poll
- 2026-03-28T12:44:41Z – cursor – shell_pid=84200 – lane=doing – Started review via workflow command
