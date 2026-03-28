---
work_package_id: WP05
title: Market data pipeline
lane: "for_review"
dependencies: [WP04]
base_branch: 001-autonomous-deribit-options-robot-WP04
base_commit: c8fe1a26fefeabbf64c18f3a6718109a55efaf5b
created_at: '2026-03-28T08:57:30.174730+00:00'
subtasks:
- T021
- T022
- T023
- T024
- T025
phase: Phase 2 - Market
assignee: ''
agent: "cursor"
shell_pid: "33684"
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
- FR-003
---

# Work Package Prompt: WP05 Market data pipeline

## Review Feedback

*[Empty until review.]*

---

## Implementation command

```bash
spec-kitty implement WP05 --base WP04
```

## Objectives and success criteria

- BTC/ETH options universe discoverable with filters.
- Each tracked instrument exposes latest book and staleness flags.

## Context and constraints

- FR-002, FR-003; bounded subscriptions (plan performance).

## Subtasks and detailed guidance

### T021 Instrument discovery

- **Purpose**: Call `get_instruments`, filter `kind`, `base_currency`, active flag.
- **Files**: `execution/internal/market/instruments.go`

### T022 Order book cache

- **Purpose**: Maintain best bid/ask and depth N levels; thread-safe.

### T023 Volatility index

- **Purpose**: Poll or subscribe vol index series needed for regime (method per latest API).

### T024 MarketSnapshot type

- **Purpose**: Struct with ts, quality_flags slice or bitmask: `StaleBook`, `WideSpread`, `Gap`.
- **Files**: `execution/internal/market/snapshot.go`

### T025 Fixtures tests

- **Purpose**: Deserialize canned JSON from `tests/fixtures/deribit/` (create minimal files).
- **Steps**: No network in unit tests.

## Risks and mitigations

- Memory: cap instruments set from config whitelist if needed.

## Review guidance

- Verify options-only filter cannot accidentally subscribe futures unless intended.

## Activity Log

- 2026-03-28T00:49:20Z - system - lane=planned - Prompt created.
- 2026-03-28T08:57:30Z – cursor – shell_pid=33684 – lane=doing – Assigned agent via workflow command
- 2026-03-28T08:59:06Z – cursor – shell_pid=33684 – lane=for_review – Ready for review: market package (instruments, book cache, vol index REST, snapshot flags), Deribit get_volatility_index_data, fixtures tests without network
