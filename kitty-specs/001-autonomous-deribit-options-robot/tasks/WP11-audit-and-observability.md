---
work_package_id: WP11
title: Audit trail and structured logging
lane: "done"
dependencies: [WP07]
base_branch: 001-autonomous-deribit-options-robot-WP07
base_commit: 4fb4e6fa29238b31d36e9addd20fba6bba339f11
created_at: '2026-03-28T09:29:53.657594+00:00'
subtasks:
- T051
- T052
- T053
- T054
phase: Phase 4 - Observability
assignee: ''
agent: "cursor"
shell_pid: "52515"
review_status: "approved"
reviewed_by: "Dmitriy Knyazev"
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
spec-kitty implement WP11 --base WP07
```

(WP07 brings in the full chain through market/regime/candidates; audit stays free of import cycles from `strategy` internals.)

## Objectives and success criteria

- **FR-010 before execution**: Cost vetoes (WP08) and risk vetoes (WP09) emit structured audit rows + log lines with shared `correlation_id`. Submit/fill paths (WP10) reuse the same logger.
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
- 2026-03-28T09:29:53Z – cursor – shell_pid=44076 – lane=doing – Assigned agent via workflow command
- 2026-03-28T09:31:13Z – cursor – shell_pid=44076 – lane=for_review – Ready for review: audit DecisionLogger, redacting slog handler, JSONL envelopes
- 2026-03-28T09:32:11Z – cursor – shell_pid=45132 – lane=doing – Started implementation via workflow command
- 2026-03-28T09:32:25Z – cursor – shell_pid=45132 – lane=for_review – Ready for review: DecisionLogger (SQLite+slog), DecisionRecord + reasons, redacting JSON slog handler, optional JSONL event envelopes (FR-010)
- 2026-03-28T09:50:08Z – cursor – shell_pid=52515 – lane=doing – Started review via workflow command
- 2026-03-28T09:50:37Z – cursor – shell_pid=52515 – lane=done – Review passed: DecisionLogger+SQLite+redacting slog JSON handler+optional JSONL envelopes aligned with event schema; DB fail still logs; tests green. Note: map order_submit to order_submitted in envelopeEventTypeFor or set EnvelopeEventType when wiring WP10/WP12.
