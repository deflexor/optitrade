---
work_package_id: WP13
title: Integration examples and operator runbooks
lane: planned
dependencies: [WP12]
subtasks:
- T060
- T061
- T062
- T063
phase: Phase 5 - Ship
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
- FR-001
- FR-002
- FR-003
- FR-004
- FR-005
- FR-006
- FR-007
- FR-008
- FR-009
- FR-010
- FR-011
---

# Work Package Prompt: WP13 Integration examples and operator runbooks

## Review Feedback

*[Empty until review.]*

---

## Implementation command

```bash
spec-kitty implement WP13 --base WP12
```

## Objectives and success criteria

- New operator can run testnet read-only path using only repo docs.
- Incident kill/flatten steps documented.

## Context and constraints

- `docs/trader-safety-cheatsheet.md`, `quickstart.md`, spec SC-* where applicable.

## Subtasks and detailed guidance

### T060 E2E read-only

- **Purpose**: CLI mode or test that logs positions and books for N seconds on testnet; exits 0.

### T061 Gated live orders

- **Purpose**: Optional env `OPTITRADE_ALLOW_TESTNET_ORDERS=1` tiny size smoke; default off.

### T062 Docs sync

- **Purpose**: Update `quickstart.md` with actual binary name and flags; ensure cheatsheet references match.

### T063 Runbook

- **Purpose**: Create `docs/runbook-incident.md`: kill switch order (systemctl stop, cancel-all), key rotation, who to ping.

## Risks and mitigations

- Mainnet footgun: example policies must stay `testnet`.

## Review guidance

- Walkthrough by reviewer who did not author the bot.

## Activity Log

- 2026-03-28T00:49:20Z - system - lane=planned - Prompt created.
