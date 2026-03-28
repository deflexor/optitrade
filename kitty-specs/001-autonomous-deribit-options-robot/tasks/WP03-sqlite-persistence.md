---
work_package_id: WP03
title: SQLite persistence
lane: planned
dependencies: [WP01]
subtasks:
- T010
- T011
- T012
- T013
- T014
phase: Phase 1 - Foundation
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
- FR-010
---

# Work Package Prompt: WP03 SQLite persistence

## Review Feedback

*[Empty until review.]*

---

## Implementation command

```bash
spec-kitty implement WP03 --base WP01
```

## Objectives and success criteria

- Migrations apply cleanly on empty DB.
- Repositories use parameterized queries only.
- Tests cover insert/select for orders and audit_decision.

## Context and constraints

- Data model: `kitty-specs/001-autonomous-deribit-options-robot/data-model.md`
- Constitution: no string-concatenated SQL with user-derived fragments.

## Subtasks and detailed guidance

### T010 Migration framework

- **Purpose**: Versioned schema evolution.
- **Steps**: Pick goose / golang-migrate / hand-rolled version table; store SQL files under `execution/internal/state/migrations/`.
- **Files**: `001_initial.sql` (or numbered).

### T011 Tables

- **Purpose**: Match entities instrument, order_record, fill_record, position_snapshot, risk_policy, audit_decision, regime_state, trade_candidate.
- **Steps**: Use INTEGER PKs or TEXT UUIDs per data-model; add indexes on `order_record(instrument_name)`, `audit_decision(ts)`.

### T012 Repository interfaces

- **Purpose**: Decouple domain from SQL.
- **Steps**: Interfaces in `internal/state`; implementations in `internal/state/sqlite`.

### T013 SQLite pragmas

- **Purpose**: WAL and busy_timeout for single-writer bot.
- **Steps**: `_journal_mode=WAL`, `busy_timeout=5000` ms documented.

### T014 Tests

- **Purpose**: Regression on migrations.
- **Steps**: Use `t.TempDir()` for DB file; run migrations; CRUD smoke.

## Risks and mitigations

- Migration failure mid-deploy: document downgrade policy in WP13 runbook.

## Review guidance

- SQL injection review: only static queries + bound params.

## Activity Log

- 2026-03-28T00:49:20Z - system - lane=planned - Prompt created.
