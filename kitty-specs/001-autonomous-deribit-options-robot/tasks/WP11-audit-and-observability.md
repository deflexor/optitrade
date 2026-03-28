---
work_package_id: WP11
title: Audit trail and structured logging
lane: planned
dependencies: [WP03, WP02, WP10]
subtasks:
- T051
- T052
- T053
- T054
phase: Phase 4 - Observability
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
- FR-010
---

# Work Package Prompt: WP11 Audit trail and structured logging

## Review Feedback

*[Empty until review.]*

---

## Implementation command

```bash
spec-kitty implement WP11 --base WP10
```

## Objectives and success criteria

- Every veto and submit path can emit structured audit row + log line with shared `correlation_id`.
- Secrets never appear in logs (constitution).

## Context and constraints

- FR-010; optional alignment with `contracts/event-envelope.schema.json`.

## Subtasks and detailed guidance

### T051 DecisionLogger interface

- **Purpose**: `LogDecision(ctx, DecisionRecord)` writes SQLite + logger; async safe or sync with error return.

### T052 Audit payload

- **Purpose**: Fields: regime_label, cost_model_version, gate_results map, candidate_id, reason enum.

### T053 Redacting logger

- **Purpose**: Wrap zap/zerolog hook to scrub keys matching `token|secret|password|client_secret`.

### T054 Optional JSONL envelope

- **Purpose**: If enabled, marshal envelope compatible with event schema for SIEM.

## Risks and mitigations

- Double-write failure: log error if DB write fails but continue protective behavior.

## Review guidance

- Spot-check log output with sample auth error (redacted).

## Activity Log

- 2026-03-28T00:49:20Z - system - lane=planned - Prompt created.
