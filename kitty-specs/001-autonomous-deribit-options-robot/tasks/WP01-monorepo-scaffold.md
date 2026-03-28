---
work_package_id: WP01
title: Monorepo scaffold and tooling
lane: "doing"
dependencies: []
base_branch: master
base_commit: 2bb50aaac87c962dcb122758c3cb3b2de7393b6b
created_at: '2026-03-28T01:00:26.355374+00:00'
subtasks:
- T001
- T002
- T003
- T004
- T005
phase: Phase 1 - Foundation
assignee: ''
agent: "cursor"
shell_pid: "15322"
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
---

# Work Package Prompt: WP01 Monorepo scaffold and tooling

## Review Feedback

*[Empty until review.]*

---

## Implementation command

```bash
spec-kitty implement WP01
```

## Objectives and success criteria

- Repository matches `kitty-specs/001-autonomous-deribit-options-robot/plan.md` project structure.
- `go build ./...` from `execution/` passes (may be minimal `main` stub).
- `pytest` from `research/` passes (may be empty smoke).
- No secrets committed; `.env.example` documents variable names only.

## Context and constraints

- Constitution: ASCII-safe markdown in docs; pin deps; no agent credential dirs in git.
- Paths: `/home/dfr/optitrade/execution/`, `/home/dfr/optitrade/research/`, `/home/dfr/optitrade/config/examples/`.

## Subtasks and detailed guidance

### T001 Directory layout

- **Purpose**: Establish physical structure before other WPs land code.
- **Steps**: Create `execution/cmd/optitrade/main.go` stub printing version; `execution/internal/` placeholders; `research/src/` or flat package; `config/examples/` empty placeholder `.gitkeep` if needed.
- **Files**: `execution/cmd/optitrade/main.go`, `research/pyproject.toml` parent dirs.
- **Validation**: Tree matches plan.md "Source Code" section.

### T002 Go module

- **Purpose**: Lock Go version for reproducible builds.
- **Steps**: `go mod init` under `execution/` with module path chosen by team (e.g. `github.com/you/optitrade/execution`); set `go 1.22`.
- **Files**: `execution/go.mod`, optional `execution/go.sum` after first tidy.

### T003 Python project

- **Purpose**: Research lane ready for backtests later.
- **Steps**: Minimal `pyproject.toml` with `pytest`; optional `src/optitrade_research/__init__.py`.
- **Parallel**: Yes vs T002.

### T004 Makefile

- **Purpose**: Single entry for CI and humans.
- **Steps**: Targets: `build` (cd execution && go build), `test` (go test + pytest), `lint` (go vet).
- **Files**: `/home/dfr/optitrade/Makefile`.

### T005 Env example and quickstart touchpoint

- **Purpose**: Document `DERIBIT_CLIENT_ID`, `DERIBIT_CLIENT_SECRET`, `OPTITRADE_CONFIG_PATH` without values.
- **Files**: `/home/dfr/optitrade/.env.example` (repo root); one-line pointer in `kitty-specs/.../quickstart.md` if paths changed.

## Risks and mitigations

- Wrong working directory in Makefile: use `$(CURDIR)` and explicit `cd execution`.

## Review guidance

- Verify no real credentials; `git grep -i secret` clean of literals.

## Activity Log

- 2026-03-28T00:49:20Z - system - lane=planned - Prompt created.
- 2026-03-28T01:00:26Z – cursor – shell_pid=15322 – lane=doing – Assigned agent via workflow command
