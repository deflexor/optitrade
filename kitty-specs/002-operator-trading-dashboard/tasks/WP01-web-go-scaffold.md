---
work_package_id: WP01
title: Web and Go dashboard scaffold
lane: "done"
dependencies: []
base_branch: master
base_commit: 8f31411cf266b93931932ed6a3aa2ea2be48141d
created_at: '2026-03-28T11:46:07.800115+00:00'
subtasks:
- T001
- T002
- T003
- T004
- T005
- T006
phase: Phase 1 - Foundation
assignee: ''
agent: "cursor"
shell_pid: "68718"
review_status: "approved"
reviewed_by: "Dmitriy Knyazev"
history:
- timestamp: '2026-03-28T11:05:00Z'
  lane: planned
  agent: system
  shell_pid: ''
  action: Prompt generated via /spec-kitty.tasks
requirement_refs:
- FR-017
---

# Work Package Prompt: WP01 -- Web and Go dashboard scaffold

## Objectives & Success Criteria

- `web/` runs `npm run dev` and `npm run build` with React, TS, Tailwind, Zustand, Axios, Router.
- Go package `src/internal/dashboard` compiles; `cmd/optitrade` accepts dashboard listen config and starts HTTP server stub (health route OK, API returns 404/501 placeholders).
- Vite proxies `/api` to local BFF port per `quickstart.md` plan.

## Context & Constraints

- Spec: `kitty-specs/002-operator-trading-dashboard/spec.md`
- Plan: `kitty-specs/002-operator-trading-dashboard/plan.md`
- Constitution: `.kittify/memory/constitution.md`
- **Implement command**: `spec-kitty implement WP01`

## Subtasks & Detailed Guidance

### Subtask T001 -- Vite React TS Tailwind shell

- **Purpose**: Operator SPA foundation per plan Assumptions.
- **Steps**:
  1. `npm create vite@latest web -- --template react-ts` (or equivalent) in repo root.
  2. Add Tailwind per official Vite guide; `content` globs `index.html`, `src/**/*.{ts,tsx}`.
  3. Replace starter UI with minimal router outlet and "Optitrade Dashboard" header.
- **Files**: `web/package.json`, `web/vite.config.ts`, `web/tailwind.config.js`, `web/postcss.config.js`, `web/src/main.tsx`, `web/src/App.tsx`
- **Parallel?**: Yes vs Go (T004).

### Subtask T002 -- Axios client and env

- **Purpose**: Cookie sessions require `withCredentials`.
- **Steps**:
  1. Add `axios`, create `src/api/client.ts` exporting configured instance: `baseURL: import.meta.env.VITE_API_BASE ?? '/api/v1'`, `withCredentials: true`, 30s timeout, JSON.
  2. Add `zustand` for future stores; stub `src/stores/authStore.ts` with `isAuthed: false` placeholder.
  3. Add `react-router-dom`; wrap app in `BrowserRouter`.
- **Files**: `web/src/api/client.ts`, `web/src/stores/authStore.ts`, `web/.env.example` with `VITE_API_BASE=`
- **Parallel?**: Yes.

### Subtask T003 -- Vite dev proxy

- **Purpose**: Dev ergonomics: single browser origin.
- **Steps**:
  1. In `vite.config.ts`, `server.proxy['/api'] = { target: 'http://127.0.0.1:8080', changeOrigin: true }` (adjust port to match Go flag default).
  2. Document port in comment.
- **Files**: `web/vite.config.ts`
- **Parallel?**: Yes.

### Subtask T004 -- Go internal/dashboard skeleton

- **Purpose**: BFF mount point.
- **Steps**:
  1. Create `src/internal/dashboard/server.go`: type `Server struct { mux *http.ServeMux }`, `NewServer(...)`, `Handler() http.Handler`.
  2. Register `GET /healthz` -> 200 `ok` for probes.
  3. Namespace API under `/api/v1/` (submux or strip prefix).
- **Files**: `src/internal/dashboard/server.go`, optional `src/internal/dashboard/doc.go`
- **Parallel?**: Yes vs web.

### Subtask T005 -- cmd/optitrade wiring

- **Purpose**: Start dashboard listener when configured.
- **Steps**:
  1. Add flag `--dashboard-listen` or env `OPTITRADE_DASHBOARD_LISTEN` (match plan).
  2. If non-empty, `go http.ListenAndServe` with `dashboard.Server` in goroutine; log listen addr at Info (no secrets).
  3. Respect graceful shutdown if main already has context (wire exit signal).
- **Files**: `src/cmd/optitrade/main.go`
- **Parallel?**: After T004.

### Subtask T006 -- embed stub

- **Purpose**: Prepare WP09 production embed.
- **Steps**:
  1. Add `src/internal/dashboard/static.go` with `//go:embed dist/*` pointing to `web/dist` OR temporary empty embed directory `web/dist/.gitkeep` and embed `all:dist` with build tag `embedui` documented.
  2. If dist missing in dev, static handler falls through to 404 JSON.
- **Files**: `web/dist/.gitkeep`, `src/internal/dashboard/static.go` (as appropriate)
- **Parallel?**: Yes.

## Test Strategy

- Manual: `go run ./src/cmd/optitrade --dashboard-listen=:8080` and `curl -s localhost:8080/healthz`.
- Manual: `cd web && npm run dev`, open browser root.

## Risks & Mitigations

- **Go version**: match root `go.mod`.
- **WSL**: document file notify quirks if Vite HMR fails.

## Review Guidance

- Verify no secrets in template files.
- Confirm UTF-8 only in new markdown.

## Activity Log

- 2026-03-28T11:05:00Z -- system -- lane=planned -- Prompt created via /spec-kitty.tasks
- 2026-03-28T11:46:07Z -- cursor -- shell_pid=66783 -- lane=doing -- Assigned agent via workflow command
- 2026-03-28T11:49:30Z -- cursor -- shell_pid=66783 -- lane=for_review -- Ready for review: Vite+Tailwind SPA, axios+zustand+router, /api dev proxy, dashboard BFF healthz+API 404 JSON + embed stub, optitrade dashboard command
- 2026-03-28T11:49:51Z -- cursor -- shell_pid=68718 -- lane=doing -- Started review via workflow command
- 2026-03-28T11:50:10Z -- cursor -- shell_pid=68718 -- lane=done -- Review passed: web stack + proxy + dashboard Server/healthz/API stub + optitrade dashboard cmd + graceful shutdown; go test and npm build verified. Note: WP02+ should rebase on master after WP01 merges.

## Markdown Formatting

Wrap HTML/XML tags in backticks. Use language identifiers in code blocks.

## Review Feedback

*[Empty until review.]*

## Optional Phase Subdirectories

Not used; flat `tasks/` only.
