# Research: Operator Trading Dashboard

**Feature**: `002-operator-trading-dashboard`  
**Date**: 2026-03-28

## Decision: Co-located Go BFF inside `optitrade`

**Rationale**: One deployment unit, direct read access to in-memory caches and SQLite without a second network hop. Aligns with single-operator scope and keeps secrets in one process boundary.

**Alternatives considered**: Separate Node or Python API (rejected: more moving parts, duplicate Deribit state). GraphQL gateway (rejected: YAGNI for v1).

## Decision: httpOnly cookie sessions stored server-side (SQLite)

**Rationale**: Browsers send cookies automatically; easy CSRF hardening with SameSite=Lax and POST-only mutations. Server-side session rows support revocation and match spec assumption that sessions may reset on process restart.

**Alternatives considered**: JWT in localStorage (rejected: XSS exposure worse than cookie+httpOnly for operator tool). Pure in-memory sessions (acceptable for dev; production uses SQLite for survivability across graceful restarts if desired).

## Decision: Allowlist at login from configuration

**Rationale**: Spec requires configurable list (e.g. `opti`). Implementation: env `OPTITRADE_DASHBOARD_ALLOWLIST=opti,backup` or config key; registration still writes any username, login checks allowlist after password verify path branches for messaging.

**Alternatives considered**: DB-only allowlist flag (possible future; config-first is faster for v1 ops).

## Decision: UI polling 2 to 3 seconds for summary strip

**Rationale**: Keeps observed age under 5 s staleness threshold when backend healthy; avoids WebSocket complexity for MVP.

**Alternatives considered**: 1 s polling (more load); SSE (good phase 2).

## Decision: Argon2id or bcrypt for password hashing

**Rationale**: Constitution and industry norm; Go std ecosystem well supported.

**Alternatives considered**: scrypt (fine; pick one and document cost parameters in implementation).

## Decision: Regime + Market mood from single bot field

**Rationale**: Spec FR-005 requires identical values on two labels; API returns `classification` once and UI duplicates labels, or API returns `{ "regime": "high", "market_mood": "high" }` from one source. Prefer one canonical string in JSON to prevent drift.

**Alternatives considered**: Two independent backend fields (rejected: violates SC-009).
