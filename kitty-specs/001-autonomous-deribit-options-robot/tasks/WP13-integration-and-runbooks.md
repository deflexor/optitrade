---
work_package_id: WP13
title: Integration examples and operator runbooks
lane: "doing"
dependencies: [WP12]
base_branch: 001-autonomous-deribit-options-robot-WP12
base_commit: cd6f070d4e58bcbd24852e7a184d7d3a4cf13779
created_at: '2026-03-28T09:53:53.573188+00:00'
subtasks:
- T060
- T061
- T062
- T063
- T064
- T065
- T066
- T067
- T068
phase: Phase 5 - Ship
assignee: ''
agent: ''
shell_pid: "53122"
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
- **G2**: spec.md **SC-001 through SC-005** each have a mapped subtask (T064-T068) with automated test, scripted check, or explicitly documented manual acceptance where automation is infeasible.

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

### T064 SC-001 defined-risk certification

- **Purpose**: Per spec SC-001: with seeded market data or mocks, assert 100% of structures that would complete (simulated fills) are allowed defined-risk templates for the active playbook; no naked short legs.

### T065 SC-002 daily loss cap

- **Purpose**: Per SC-002: stress scenario where cumulative loss hits daily cap; assert no additional risk-increasing orders until reset or session rule; align with WP09 session boundary.

### T066 SC-003 protective timing

- **Purpose**: Per SC-003: simulate feed loss or auth failure; protective mode blocks new risk within 60s detection budget; cross-reference WP12 T058-T059.

### T067 SC-004 reconciliation / orphans

- **Purpose**: Per SC-004: scripted trading window against mock or testnet; reconciliation leaves zero unexplained orphan legs per runbook procedure.

### T068 SC-005 audit completeness sampling

- **Purpose**: Per SC-005: on fixture corpus of decisions, measure fraction of rows with regime + cost model version + risk gate outcome; assert >=90% on sample; 100% on injected warning-breach cases.

## Risks and mitigations

- Mainnet footgun: example policies must stay `testnet`.

## Review guidance

- Walkthrough by reviewer who did not author the bot.

## Activity Log

- 2026-03-28T00:49:20Z - system - lane=planned - Prompt created.
