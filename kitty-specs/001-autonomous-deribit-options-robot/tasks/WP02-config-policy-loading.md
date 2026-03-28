---
work_package_id: WP02
title: Config load and policy validation
lane: "done"
dependencies: [WP01]
base_branch: 001-autonomous-deribit-options-robot-WP01
base_commit: 105fe0cb3edb429128ea3febdffc6218bd9a388c
created_at: '2026-03-28T01:12:23.887120+00:00'
subtasks:
- T006
- T007
- T008
- T009
phase: Phase 1 - Foundation
assignee: ''
agent: "cursor"
shell_pid: "21820"
review_status: "approved"
reviewed_by: "Dmitriy Knyazev"
history:
- timestamp: '2026-03-28T00:49:20Z'
  lane: planned
  agent: system
  shell_pid: ''
  action: Prompt generated via /spec-kitty.tasks
requirement_refs:
- FR-007
- FR-011
---

# Work Package Prompt: WP02 Config load and policy validation

## Review Feedback

*[Empty until review.]*

---

## Implementation command

```bash
spec-kitty implement WP02 --base WP01
```

## Objectives and success criteria

- Runtime loads policy file and rejects valid-schema violations with clear errors.
- Example policy file exists for testnet with conservative limits.

## Context and constraints

- Schema: `kitty-specs/001-autonomous-deribit-options-robot/contracts/config-policy.schema.json`
- Constitution: treat file contents as untrusted; validate before use.

## Subtasks and detailed guidance

### T006 Schema integration

- **Purpose**: Single validation path for policy JSON.
- **Steps**: Choose embedded schema (go:embed) or load from path relative to repo; document production behavior (embed recommended).
- **Files**: `execution/internal/config/schema.go` (suggested).

### T007 Config loader

- **Purpose**: Typed access to limits, playbooks, cost model, protective thresholds.
- **Steps**: Parse JSON; map to structs with explicit types for decimals (string or shopspring decimal); env override only for paths and secrets, not for numeric limits unless you document.
- **Files**: `execution/internal/config/load.go`

### T008 Example policy

- **Purpose**: Operators copy and edit.
- **Steps**: Fill `config/examples/policy.example.json` with `environment: testnet`, tight limits, `rules_v1` regime placeholders.
- **Files**: `config/examples/policy.example.json`

### T009 Unit tests

- **Purpose**: Gate bad configs.
- **Steps**: Cases: missing `max_daily_loss`, wrong enum for playbook structure, extra field if `additionalProperties: false`.
- **Files**: `execution/internal/config/load_test.go`

## Risks and mitigations

- JSON decimal precision: avoid `float64` for money fields.

## Review guidance

- Confirm schema version field enforced.

## Activity Log

- 2026-03-28T00:49:20Z - system - lane=planned - Prompt created.
- 2026-03-28T01:12:23Z – cursor – shell_pid=19720 – lane=doing – Assigned agent via workflow command
- 2026-03-28T01:13:17Z – cursor – shell_pid=19720 – lane=for_review – Ready for review: embedded JSON Schema policy validation, LoadFile/LoadBytes, testnet example policy, unit tests for schema violations
- 2026-03-28T01:14:32Z – cursor – shell_pid=21820 – lane=doing – Started review via workflow command
- 2026-03-28T01:14:47Z – cursor – shell_pid=21820 – lane=done – Review passed: schema matches contract, validate-before-unmarshal, string decimals, tests cover required cases and version pattern
