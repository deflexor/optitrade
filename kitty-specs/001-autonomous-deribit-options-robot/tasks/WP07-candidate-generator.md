---
work_package_id: WP07
title: Playbooks liquidity and candidates
lane: "for_review"
dependencies: [WP05, WP06, WP02]
base_branch: 001-autonomous-deribit-options-robot-WP07-merge-base
base_commit: 923bd42415085bdc640a6524d2c4fd07e969307b
created_at: '2026-03-28T09:19:59.704213+00:00'
subtasks:
- T030
- T031
- T032
- T033
- T034
- T035
phase: Phase 2 - Strategy
assignee: ''
agent: "cursor"
shell_pid: "39382"
review_status: ''
reviewed_by: ''
history:
- timestamp: '2026-03-28T00:49:20Z'
  lane: planned
  agent: system
  shell_pid: ''
  action: Prompt generated via /spec-kitty.tasks
requirement_refs:
- FR-002
- FR-005
- FR-011
---

# Work Package Prompt: WP07 Playbooks liquidity and candidates

## Review Feedback

*[Empty until review.]*

---

## Implementation command

```bash
spec-kitty implement WP07 --base WP06
```

## Objectives and success criteria

- Only defined-risk structures allowed; naked short legs impossible by construction.
- Illiquid strikes never emitted.

## Context and constraints

- FR-005, FR-011; playbook enums in policy schema.

## Subtasks and detailed guidance

### T030 Liquidity gate

- **Purpose**: For each candidate strike/expiry compute spread bps and size at touch; compare to policy.

### T031 Playbook resolver

- **Purpose**: Map current regime to `allowed_structures` list; error if empty.

### T032 Templates

- **Purpose**: Build leg specs for vertical credit/debit and iron condor skeleton for BTC/ETH; width parameters from config or constants file.

### T033 Invariants

- **Purpose**: Assert net short options not exposed without long hedge per structure template; panic in dev if violated.

### T034 Unit tests

- **Purpose**: Wide spread -> no candidate; mock book in memory.

### T035 Documentation

- **Purpose**: Comment or short markdown in repo listing Deribit instrument name pattern per leg (update `research.md` or package doc.go).

## Risks and mitigations

- Wrong instrument name -> order reject at exchange; add dry-run validation call if API supports.

## Review guidance

- Strategy code readable enough for manual audit of defined-risk.

## Activity Log

- 2026-03-28T00:49:20Z - system - lane=planned - Prompt created.
- 2026-03-28T09:19:59Z – cursor – shell_pid=39382 – lane=doing – Assigned agent via workflow command
- 2026-03-28T09:20:49Z – cursor – shell_pid=39382 – lane=for_review – Ready for review: strategy package with liquidity gate, playbook resolver, vertical/IC templates, defined-risk validation, tests, docs/research.md
