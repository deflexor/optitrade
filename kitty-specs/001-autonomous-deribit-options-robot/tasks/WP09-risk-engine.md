---
work_package_id: WP09
title: Risk engine
lane: planned
dependencies:
- WP03
- WP04
- WP08
subtasks:
- T040
- T041
- T042
- T043
- T044
- T045
phase: Phase 3 - Risk
assignee: ''
agent: ''
shell_pid: ''
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
- FR-007
- FR-011
---

# Work Package Prompt: WP09 Risk engine

## Review Feedback

*[Empty until review.]*

---

## Implementation command

```bash
spec-kitty implement WP09 --base WP08
```

(Base branch must include **WP11** so risk vetoes persist via `DecisionLogger`.)

## Objectives and success criteria

- Each gate independently testable; combined pre-trade check blocks bad candidates.
- **FR-010**: Every gate failure writes an `audit_decision` (or equivalent) via WP11 logger with gate map and correlation ID.

## Context and constraints

- FR-004, FR-007, FR-011; Greeks from exchange position summaries if available, else document approximation.

## Subtasks and detailed guidance

### T040 Portfolio snapshot builder

- **Purpose**: Merge RPC positions with local optional adjustments; compute delta/vega aggregates (use exchange if provided).

### T041 Limit gates

- **Purpose**: max delta, vega, premium at risk, open orders per instrument.

### T042 Daily loss

- **Purpose**: Track realized/unrealized per policy session boundary (UTC midnight or configurable).

### T043 Per-trade max loss

- **Purpose**: Estimate worst loss for candidate structure + fees; compare to limit.

### T044 Time in trade

- **Purpose**: Track open strategy opened_at; veto new adds if exceeds max (define per strategy id).

### T045 Tests

- **Purpose**: Fabricate portfolio state and candidate; assert correct veto.

## Risks and mitigations

- Greek cache stale: refresh before each trade decision or subscribe user events.

## Review guidance

- Cross-check limit fields with `config-policy.schema.json`.

## Activity Log

- 2026-03-28T00:49:20Z - system - lane=planned - Prompt created.
