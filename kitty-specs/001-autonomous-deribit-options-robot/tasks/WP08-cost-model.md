---
work_package_id: WP08
title: Cost model and edge scoring
lane: planned
dependencies: [WP07]
subtasks:
- T036
- T037
- T038
- T039
phase: Phase 3 - Decisions
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
- FR-006
- FR-011
---

# Work Package Prompt: WP08 Cost model and edge scoring

## Review Feedback

*[Empty until review.]*

---

## Implementation command

```bash
spec-kitty implement WP08 --base WP07
```

## Objectives and success criteria

- Veto reasons enumerated for audit when edge after costs non-positive.
- IV vs book sanity hook stub or full implementation per `research.md` section 7.

## Context and constraints

- FR-006; conservative defaults.

## Subtasks and detailed guidance

### T036 Fees and spread

- **Purpose**: Load fee bps from policy; compute half-spread from bid/ask; convert to common unit with expected edge.

### T037 Slippage and adverse selection

- **Purpose**: Add regime-dependent bps; document formula in code comment.

### T038 ScoreCandidate API

- **Purpose**: Returns `(ok bool, vetoReason string, breakdown struct)` for logging.

### T039 IV-book sanity

- **Purpose**: If using IV quotes, compare against book mid movement; on conflict return veto `iv_stale`.

## Risks and mitigations

- Unit mismatch (BTC vs USD): align with Deribit instrument `quote_currency`.

## Review guidance

- Table tests with golden CSV or struct literals.

## Activity Log

- 2026-03-28T00:49:20Z - system - lane=planned - Prompt created.
