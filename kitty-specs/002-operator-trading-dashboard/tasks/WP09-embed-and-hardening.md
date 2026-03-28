---
work_package_id: WP09
title: SPA embed build pipeline and hardening
lane: "doing"
dependencies: [WP08]
base_branch: 002-operator-trading-dashboard-WP08
base_commit: 95e4029a6bb5797c22d9328b1ae300b26fb30ccd
created_at: '2026-03-28T12:50:25.068285+00:00'
subtasks:
- T043
- T044
- T045
- T046
- T047
phase: Phase 5 - Release
assignee: ''
agent: "cursor"
shell_pid: "85625"
review_status: ''
reviewed_by: ''
history:
- timestamp: '2026-03-28T11:05:00Z'
  lane: planned
  agent: system
  shell_pid: ''
  action: Prompt generated via /spec-kitty.tasks
requirement_refs:
- FR-017
---

# Work Package Prompt: WP09 -- SPA embed build pipeline and hardening

## Objectives & Success Criteria

- Production binary serves `GET /` and assets from embedded `web/dist`; API remains under `/api/v1/`.
- Makefile (or CI) builds web then Go with embed.
- Auth pages complete **User Story 1** scenarios 4-7 in spec (register, login, allowlist gate, wrong password).
- Encoding validation passes; CSRF strategy documented.

## Context & Constraints

- **Implement command**: `spec-kitty implement WP09 --base WP08`
- Note: if WP07 not merged in same branch, rebase or implement WP09 after WP04+WP07 merged - release candidate needs full flows.

## Subtasks & Detailed Guidance

### T043 -- go:embed

- Build pipeline copies or outputs to `web/dist`.
- Static handler: try file; fallback `index.html` for SPA routes except `/api`.
- Security headers baseline (optional `middleware`).

### T044 -- Makefile / CI

- Target `make web` and `make build` ordering documented in root `Makefile`.

### T045 -- Auth pages UX

- `/login`, `/register` routes; protected layout wraps dashboard routes redirecting to login if 401 from bootstrap fetch.
- Display server error messages exactly for not-ready path.

### T046 -- CSRF / cookies

- Document: SameSite=Lax + POST-only mutations + no third-party origins in v1.
- If needed, add CSRF token header for POSTs in same release.

### T047 -- validate-encoding

- Run `spec-kitty validate-encoding --feature 002-operator-trading-dashboard --fix`; no smart quotes in new markdown.

## Test Strategy

- Manual smoke script in `quickstart.md`.
- Optional Playwright later (out of scope unless added).

## Risks

- **Cache busting**: Vite hashed assets work with embed FS paths.

## Review Feedback

*[Empty.]*

## Activity Log

- 2026-03-28T11:05:00Z -- system -- lane=planned -- Prompt created via /spec-kitty.tasks
- 2026-03-28T12:50:25Z -- cursor -- shell_pid=85625 -- lane=doing -- Assigned agent via workflow command
