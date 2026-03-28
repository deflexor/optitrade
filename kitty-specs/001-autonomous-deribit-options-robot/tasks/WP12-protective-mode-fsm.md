---
work_package_id: WP12
title: Protective mode and session FSM
lane: "done"
dependencies: [WP05, WP10, WP11]
base_branch: 001-autonomous-deribit-options-robot-WP12-merge-base
base_commit: b6cc43ea157d127313fae169b0714e7ab680ba06
created_at: '2026-03-28T09:46:39.367103+00:00'
subtasks:
- T055
- T056
- T057
- T058
- T059
phase: Phase 4 - Safety
assignee: ''
agent: "cursor"
shell_pid: "54525"
review_status: "approved"
reviewed_by: "Dmitriy Knyazev"
history:
- timestamp: '2026-03-28T00:49:20Z'
  lane: planned
  agent: system
  shell_pid: ''
  action: Prompt generated via /spec-kitty.tasks
requirement_refs:
- FR-009
---

# Work Package Prompt: WP12 Protective mode and session FSM

## Review Feedback

*[Empty until review.]*

---

## Implementation command

```bash
spec-kitty implement WP12 --base WP11
```

## Objectives and success criteria

- When triggered, no new risk-increasing orders are submitted; reduce-only exits allowed per policy.
- Measurable delay from injected staleness to blocked submits (target spec SC-003 60s worst case includes detection).

## Context and constraints

- FR-009; inputs from market quality_flags, RPC auth errors, WS disconnect duration.

## Subtasks and detailed guidance

### T055 FSM

- **Purpose**: States in data-model `session` section; transitions table in code.

### T056 Triggers

- **Purpose**: Subscribe to health events from WS client and market staleness; config thresholds from policy `protective_mode`.

### T057 Execution guard

- **Purpose**: Execution layer checks FSM before submit; returns error if blocked.

### T058 Tests

- **Purpose**: Fake market feed stops updating; advance clock or inject flag; assert state `protective_flatten`.

### T059 NFR note

- **Purpose**: Document timing budget in code comment; if full 60s end-to-end not met MVP, mark gap in Activity Log for follow-up.

## Risks and mitigations

- Deadlock between flatten and reconciliation: timeout and escalate to `frozen`.

## Review guidance

- Manual drill checklist in WP13 should reference FSM states.

## Activity Log

- 2026-03-28T00:49:20Z - system - lane=planned - Prompt created.
- 2026-03-28T09:46:39Z – cursor – shell_pid=51298 – lane=doing – Assigned agent via workflow command
- 2026-03-28T09:49:20Z – cursor – shell_pid=51298 – lane=for_review – Ready for review: session FSM (running/paused/protective_flatten/frozen), market+WS+RPC triggers, Placer AllowSubmit guard, policy ws_down_grace_ms/max_flatten_wait_ms, WS OnHealth; SC-003 full E2E timing noted in session package comment
- 2026-03-28T09:56:53Z – cursor – shell_pid=54525 – lane=doing – Started review via workflow command
- 2026-03-28T09:57:11Z – cursor – shell_pid=54525 – lane=done – Review passed: session FSM matches transition table, Placer AllowSubmit guard, WS OnHealth for disconnect/reconnect, policy protective timers, tests green; SC-003 MVP gap documented in package doc. WP13 (for_review) should rebase onto WP12 after merge if stack shifts.
