---
work_package_id: WP04
title: Deribit JSON-RPC and WebSocket client
lane: "done"
dependencies: [WP01, WP02]
base_branch: 001-autonomous-deribit-options-robot-WP04-merge-base
base_commit: 6d1722aa28d3a54d10886473ee8e16951b5a93d3
created_at: '2026-03-28T01:34:09.547621+00:00'
subtasks:
- T015
- T016
- T017
- T018
- T019
- T020
phase: Phase 2 - Connectivity
assignee: ''
agent: "cursor"
shell_pid: "26989"
review_status: "approved"
reviewed_by: "Dmitriy Knyazev"
history:
- timestamp: '2026-03-28T00:49:20Z'
  lane: planned
  agent: system
  shell_pid: ''
  action: Prompt generated via /spec-kitty.tasks
requirement_refs:
- FR-001
- FR-004
---

# Work Package Prompt: WP04 Deribit JSON-RPC and WebSocket client

## Review Feedback

*[Empty until review.]*

---

## Implementation command

```bash
spec-kitty implement WP04 --base WP02
```

(Use base that includes WP01+WP02; if WP03 merged first, branch from latest foundation.)

## Objectives and success criteria

- Authenticated read-only calls work on testnet.
- WebSocket reconnect restores subscriptions.

## Context and constraints

- Deribit JSON-RPC 2.0 over HTTPS; WebSocket for streaming (see `research.md` section 3).
- Never log client secret or access token values.

## Subtasks and detailed guidance

### T015 JSON-RPC core

- **Purpose**: Typed requests with `id`, method, params; parse error object.
- **Files**: `execution/internal/deribit/rpc/client.go`

### T016 Authentication

- **Purpose**: Implement current Deribit auth (client credentials / refresh); store token in memory; mutex for refresh.
- **Steps**: Read official auth flow; redact errors before log.
- **Files**: `execution/internal/deribit/auth.go`

### T017 WebSocket client

- **Purpose**: Subscribe channels; heartbeat; exponential backoff reconnect; resubscribe list.
- **Files**: `execution/internal/deribit/ws/client.go`

### T018 Private REST maps

- **Purpose**: `get_positions`, `get_open_orders`, `get_account_summaries` structs matching JSON fields with nullable types.
- **Parallel**: Yes vs T019.

### T019 Public REST maps

- **Purpose**: `get_instruments`, `get_order_book`, `ticker` (as used by plan).

### T020 Integration test

- **Purpose**: Guard live test with env vars; skip in CI without credentials.
- **Files**: `execution/internal/deribit/integration_test.go` build tag `integration` optional.

## Risks and mitigations

- RPC id race: single flight or mutex for inflight map.

## Review guidance

- Grep logs for `secret`, `Bearer` literals.

## Activity Log

- 2026-03-28T00:49:20Z - system - lane=planned - Prompt created.
- 2026-03-28T01:34:09Z – cursor – shell_pid=26989 – lane=doing – Assigned agent via workflow command
- 2026-03-28T10:01:31Z – cursor – shell_pid=26989 – lane=done – Review passed: JSON-RPC client (rpc/), auth w/ mutex+refresh, REST facade (private+public maps), WS reconnect+resubscribe, integration test skips w/o creds; no secret logging in error paths beyond redacted bodies. Verified on master: make test green. WP04 branch had no unique commits vs base.
