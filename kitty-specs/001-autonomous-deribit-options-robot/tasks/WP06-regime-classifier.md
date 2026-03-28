---
work_package_id: WP06
title: Regime classifier
lane: "doing"
dependencies: [WP05, WP03]
base_branch: 001-autonomous-deribit-options-robot-WP06-merge-base
base_commit: 099e09b02498cd70c3852ac5a320e544dfb08bb2
created_at: '2026-03-28T09:08:40.845179+00:00'
subtasks:
- T026
- T027
- T028
- T029
phase: Phase 2 - Strategy
assignee: ''
agent: "cursor-composer"
shell_pid: "37351"
review_status: ''
reviewed_by: ''
history:
- timestamp: '2026-03-28T00:49:20Z'
  lane: planned
  agent: system
  shell_pid: ''
  action: Prompt generated via /spec-kitty.tasks
requirement_refs:
- FR-005
---

# Work Package Prompt: WP06 Regime classifier

## Review Feedback

*[Empty until review.]*

---

## Implementation command

```bash
spec-kitty implement WP06 --base WP05
```

(Chain includes WP03 for persistence; if implementing linearly after WP05, use `--base` that has WP03+WP05 merged.)

## Objectives and success criteria

- Deterministic `rules_v1` labels for synthetic inputs.
- Regime changes durable for audit (spec SC-005 alignment).

## Context and constraints

- `research.md` section 4: rule-based MVP with classifier_version.

## Subtasks and detailed guidance

### T026 rules_v1

- **Purpose**: Read `regime` block from policy; compare vol index to low/high thresholds; output `low|normal|high`.
- **Optional**: Hysteresis: require N minutes in band before flip.

### T027 Persist regime_state

- **Purpose**: Insert row on change or fixed interval; include `inputs_digest` hash of inputs.

### T028 Unit tests

- **Purpose**: Table tests for boundary thresholds.

### T029 Wire to market

- **Purpose**: Regime evaluator accepts `MarketSnapshot` + vol features interface.

## Risks and mitigations

- Missing data: default to `normal` or `frozen` behavior per policy flag (document).

## Review guidance

- Every label must carry `classifier_version` for audit.

## Activity Log

- 2026-03-28T00:49:20Z - system - lane=planned - Prompt created.
- 2026-03-28T09:08:40Z – cursor-composer – shell_pid=37351 – lane=doing – Assigned agent via workflow command
