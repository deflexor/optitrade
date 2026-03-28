---
work_package_id: WP10
title: Order execution and reconciliation
lane: "doing"
dependencies: [WP04, WP09, WP11]
base_branch: 001-autonomous-deribit-options-robot-WP10-merge-base
base_commit: 071378c129c0f4421a08ac04ca2599140ffe7c0e
created_at: '2026-03-28T09:40:55.860168+00:00'
subtasks:
- T046
- T047
- T048
- T049
- T050
phase: Phase 4 - Execution
assignee: ''
agent: "cursor-composer"
shell_pid: "49744"
review_status: ''
reviewed_by: ''
history:
- timestamp: '2026-03-28T00:49:20Z'
  lane: planned
  agent: system
  shell_pid: ''
  action: Prompt generated via /spec-kitty.tasks
requirement_refs:
- FR-004
- FR-008
---

# Work Package Prompt: WP10 Order execution and reconciliation

## Review Feedback

*[Empty until review.]*

---

## Implementation command

```bash
spec-kitty implement WP10 --base WP09
```

(Base already includes WP11 from the WP08/WP09 chain; explicit WP11 dependency for audit on submits and fills.)

## Objectives and success criteria

- Order path respects post-only default and reduce-only on exits.
- Local order state converges to exchange after reconciliation cycle.

## Context and constraints

- FR-008; combo mapping from WP07 templates.

## Subtasks and detailed guidance

### T046 Placement

- **Purpose**: Build `private/buy` or equivalent with `post_only`, `label`, `advanced` fields per Deribit options API.

### T047 Multi-leg

- **Purpose**: If exchange supports combo order name, map legs; else document sequential placement with rollback policy MVP.

### T048 Cancel helpers

- **Purpose**: `cancel_all_by_instrument`, `cancel_by_label` wrappers.

### T049 Fill handler

- **Purpose**: On user_trade event or polling fills, update DB and in-memory exposure.

### T050 Reconciliation

- **Purpose**: Periodic diff: missing local -> insert; orphan local -> mark canceled.

## Risks and mitigations

- Duplicate submit: use client id idempotency pattern.

## Review guidance

- Dry-run mode flag prevents RPC submit during tests.

## Activity Log

- 2026-03-28T00:49:20Z - system - lane=planned - Prompt created.
- 2026-03-28T09:40:55Z – cursor-composer – shell_pid=49744 – lane=doing – Assigned agent via workflow command
