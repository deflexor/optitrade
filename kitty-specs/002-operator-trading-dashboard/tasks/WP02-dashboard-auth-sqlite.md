---
work_package_id: WP02
title: Dashboard SQLite auth and handlers
lane: planned
dependencies: [WP01]
subtasks:
- T007
- T008
- T009
- T010
- T011
- T012
- T013
phase: Phase 1 - Auth
assignee: ''
agent: ''
shell_pid: ''
review_status: ''
reviewed_by: ''
history:
- timestamp: '2026-03-28T11:05:00Z'
  lane: planned
  agent: system
  shell_pid: ''
  action: Prompt generated via /spec-kitty.tasks
requirement_refs:
- FR-014
- FR-015
- FR-016
- FR-018
---

# Work Package Prompt: WP02 -- Dashboard SQLite auth and handlers

## Objectives & Success Criteria

- Users can register (username/password only); passwords stored salted-hashed.
- Login: allowlisted users with correct password get session cookie; non-allowlisted get JSON `{ "error": "Sorry, feature not ready" }` with HTTP 403; wrong password for allowlisted user gets 401 generic body **without** that exact string.
- Logout clears session server-side and cookie.
- Automated tests cover the above matrix.

## Context & Constraints

- OpenAPI: `kitty-specs/002-operator-trading-dashboard/contracts/dashboard-api.openapi.yaml`
- Data model: `kitty-specs/002-operator-trading-dashboard/data-model.md`
- Clarifications in `spec.md`: exact string `Sorry, feature not ready`
- **Implement command**: `spec-kitty implement WP02 --base WP01`

## Subtasks & Detailed Guidance

### Subtask T007 -- Migrations

- **Purpose**: Durable users and sessions.
- **Steps**:
  1. Add migration files creating `dashboard_user` and `dashboard_session` with columns from data-model.
  2. Index `dashboard_user(username)` unique; index `dashboard_session(token_hash)`.
  3. Hook migrations into `internal/state` migration runner used by bot DB file (same SQLite path or documented dashboard DB if split; prefer single file per plan).
- **Files**: `src/internal/state/migrations/...` or dashboard-specific migrate pkg
- **Parallel?**: No (schema first).

### Subtask T008 -- Password hashing

- **Purpose**: FR-018.
- **Steps**:
  1. Choose Argon2id (`golang.org/x/crypto/argon2` with `passlib`-style wrapper) or bcrypt.
  2. Functions `HashPassword(plain string) (string, error)`, `VerifyPassword(hash, plain) bool` using constant-time compare.
- **Files**: `src/internal/dashboard/auth/hash.go`
- **Parallel?**: Yes after T007.

### Subtask T009 -- Register repository

- **Purpose**: FR-014.
- **Steps**:
  1. Normalize username: trim, lower-case for lookup; reject empty, length bounds (e.g. 3-32).
  2. Insert user with hash; map `UNIQUE` to HTTP 409.
- **Files**: `src/internal/dashboard/auth/user_repo.go`
- **Parallel?**: Yes.

### Subtask T010 -- Allowlist

- **Purpose**: FR-015.
- **Steps**:
  1. Parse `OPTITRADE_DASHBOARD_ALLOWLIST` at startup into `map[string]struct{}`.
  2. Expose `IsAllowlisted(username string) bool`.
- **Files**: `src/internal/dashboard/auth/allowlist.go`
- **Parallel?**: Yes.

### Subtask T011 -- Sessions

- **Purpose**: httpOnly cookie auth.
- **Steps**:
  1. On login success: 32+ byte random token from `crypto/rand`; SHA-256 store in DB; Set-Cookie `optitrade_session` with Path `/`, HttpOnly, SameSite=Lax, Secure when TLS (config flag).
  2. Middleware loads session by hash, checks `expires_at`, refreshes sliding window optional (document).
  3. Logout deletes session row and Max-Age=0 cookie.
- **Files**: `src/internal/dashboard/auth/session.go`
- **Parallel?**: After T007.

### Subtask T012 -- HTTP handlers

- **Purpose**: Match OpenAPI paths and status codes.
- **Steps**:
  1. Register under `/api/v1/auth/` on dashboard mux.
  2. JSON errors: content-type application/json; never echo password.
  3. Login flow: if user exists and NOT allowlisted -> 403 + exact not-ready message (check order: spec implies after "password verify" - implement: if !allowlisted { return 403 not-ready } without revealing password validity; if allowlisted && bad password -> 401 generic `"error": "invalid credentials"`).
  4. Align with SC-006: wrong password allowlisted must NOT return not-ready message.
- **Files**: `src/internal/dashboard/auth_handlers.go`
- **Parallel?**: After T008-T011.

### Subtask T013 -- httptest matrix

- **Purpose**: Constitution critical path tests.
- **Steps**:
  1. Table tests: register twice, login allowlisted good, login non-allowlisted -> 403 exact body, login allowlisted bad pass -> 401, protected route without cookie -> 401.
  2. Optional: hook for rate limiting structure (skip implementation if not in scope; leave `// TODO WP rate limit`).
- **Files**: `src/internal/dashboard/auth_handlers_test.go`
- **Parallel?**: After T012.

## Test Strategy

- `go test ./src/internal/dashboard/... -count=1`

## Risks & Mitigations

- **User enumeration**: Spec requires distinct non-allowlisted message; accept tradeoff per product.
- **SQL injection**: only prepared statements.

## Review Guidance

- Grep for `Sorry, feature not ready` - must match exactly once in handler path.
- No password logging in any package.

## Activity Log

- 2026-03-28T11:05:00Z -- system -- lane=planned -- Prompt created via /spec-kitty.tasks

## Markdown Formatting

Use ```go for Go snippets.

## Review Feedback

*[Empty until review.]*
