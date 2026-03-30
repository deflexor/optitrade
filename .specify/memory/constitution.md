<!--
  Sync Impact Report
  Version change: 1.0.0 → 1.1.0 (MINOR: expanded principles, brownfield-derived quality/testing/UX/perf)
  Principles:
    - I Operator safety → retained (wording tightened)
    - II Boundaries → retained, merged BFF/API/static and anti-duplication
    - III Testing → split and expanded into IV Testing + parts moved to III Code quality
    - III (new) Code quality & correctness → monetary types, schema validation, logging discipline
    - IV Observability → folded into VI Performance/reliability/observability
    - V Minimal changes → retained as VII Focused delivery
    - (new) V User experience consistency (dashboard stack + API/client conventions)
    - (new) VI Performance, reliability, and observability
  Sections: Technology & Ops expanded (authoritative go.mod, lint); Workflow expanded; Governance
    materially expanded (precedence, decision rules, review expectations)
  Templates:
    - .specify/templates/plan-template.md — ✅ Constitution Check gates updated
    - .specify/templates/spec-template.md — ✅ reviewed, no mandatory section change
    - .specify/templates/tasks-template.md — ✅ task hints for quality/UX/perf
    - .specify/templates/commands/*.md — n/a
  Follow-up TODOs: none
-->

# Optitrade Constitution

## Core Principles

### I. Operator and capital safety (NON-NEGOTIABLE)

Features that touch live trading, order routing, position sizing, policies, or exchange credentials
MUST preserve or improve operator safety. Follow `docs/trader-safety-cheatsheet.md` and
`docs/runbook-incident.md` where applicable; document new risk edges before shipping. MUST NOT
introduce silent increases in capital exposure or bypass existing risk gates (policy, vetoes,
protective mode) without explicit spec and plan justification.

**Rationale:** Deribit options workflows have immediate financial impact; the tree encodes this in
policy-driven risk engines and audited decisions.

### II. System and API boundaries

The Go module (`src/`) owns CLI, domain logic (risk, execution, state, Deribit clients), SQLite
persistence, JSON Schema–validated policy loading, and the dashboard BFF (`internal/dashboard`).
The operator UI is the React + Vite app under `web/`, built into `internal/dashboard/dist/` for
`go:embed`. MUST keep HTTP contracts explicit: breaking `/api/v1` JSON shapes or semantics require
versioned paths, migration notes, or coordinated UI updates. MUST NOT duplicate trading or policy
rules in the SPA; the server remains the source of truth except for purely presentational logic.

**Rationale:** Matches the current split (BFF + SPA proxy) and avoids divergent behavior.

### III. Code quality and correctness

Go code MUST follow patterns visible in existing packages: table-driven and parallel-safe tests
where appropriate; `t.Helper()` for shared fixtures; `context` plumbed through I/O boundaries;
structured errors that carry enough context for operators without leaking secrets. Policy and
limits MUST be loaded through validated configuration (`config` + JSON Schema); monetary and risk
limits MUST NOT be represented as `float64` in policy or gate logic where the codebase already uses
decimal strings or `big.Rat`-style exact arithmetic—new features MUST stay consistent with those
representations. Logging MUST use `log/slog` with redaction for sensitive keys (`internal/audit`
`RedactingReplaceAttr` / `NewJSONHandler` patterns); MUST NOT log raw tokens, secrets, or
passwords.

**Rationale:** Aligns with `config.LoadBytes`, risk snapshot math, RPC error handling, and audit
logging already in-tree.

### IV. Testing standards

Changes MUST pass repository gates: `make test` (Go `go test ./...` plus `research/` pytest),
`make test-web` when `web/` changes, and `make lint` (`go vet`) for Go changes. New Go behavior in
packages that already have tests MUST add or extend tests in the same style (unit tests with
in-memory SQLite, stubbed exchange interfaces as in `internal/execution`, etc.). Live exchange
tests MUST remain behind `//go:build integration`, optional env vars, and `t.Skip` without
credentials (see `internal/observe/observe_integration_test.go`). New or substantially implemented
dashboard HTTP handlers MUST include handler-level tests (status codes, JSON shape, security
headers as applicable) before being treated as production-ready—the current BFF is still thin but
growing. Research changes under `research/` SHOULD keep `tests/` meaningful rather than empty
smoke-only fixtures unless explicitly out of scope.

**Rationale:** Makefile and package layout define the team baseline; brownfield gaps (e.g. no
dashboard tests yet) MUST shrink as endpoints gain behavior.

### V. User experience consistency

Dashboard UI work MUST stay consistent with the existing stack: Tailwind CSS, slate-oriented
palette, shared layout shell (`Shell` + `Outlet` in `web/src/App.tsx`), and typography/spacing
patterns already in use. API consumption SHOULD go through the shared Axios instance
(`web/src/api/client.ts`): JSON accept/content types, configurable `VITE_API_BASE`, 30s timeout,
`withCredentials` aligned with cookie/session plans. Server JSON errors SHOULD follow a stable,
machine-friendly shape (e.g. `error` and `message` fields as in `internal/dashboard` handlers).
Operator-visible copy MUST be clear for stressful use (precise units, avoid ambiguous trading
jargon without context). Vite dev proxy targets and Makefile `DASHBOARD_LISTEN` MUST stay
documented and in sync when ports change.

**Rationale:** The SPA is early-stage; conventions prevent a fragmented operator experience.

### VI. Performance, reliability, and observability

Hot paths (RPC, WebSocket feeds, order placement) MUST respect `context` cancellation and bounded
resource use (e.g. HTTP response bodies size-limited as in `internal/deribit/rpc`, client
timeouts aligned with upstream expectations). Long-running processes SHOULD shut down gracefully
(`http.Server` shutdown patterns in `cmd/optitrade`). User-relevant services SHOULD expose health
checks (`/healthz`). Latency-sensitive work SHOULD avoid unnecessary allocation and serial bottlenecks
in identified hot loops; when trading off clarity versus performance, prefer clarity unless profiling
or SLA motivates otherwise—document such trade-offs in the plan. New failure modes SHOULD be
observable (structured logs, metrics hooks, or runbook updates) consistent with operator docs and
health endpoints.

**Rationale:** Matches existing client timeouts, body limits, dashboard health endpoint, and
incident docs.

### VII. Focused delivery

Implement the smallest diff that satisfies the spec. SHOULD match local naming, package layout, and
Makefile conventions. MUST NOT introduce unrelated refactors, drive-by dependency churn, or
documentation sprawl not tied to the active feature or a constitution amendment.

**Rationale:** Brownfield safety depends on reviewable, incremental change.

## Technology and operational constraints

Authoritative **Go** version is the `go` directive in `src/go.mod` (toolchain may track latest stable
used by maintainers); human-facing docs SHOULD reference `go.mod` when they disagree with older
README lines. **Node.js 20+** and **npm** power `web/`; **Make** encodes canonical developer
commands. Dashboard static assets are generated under `src/internal/dashboard/dist/`, not edited by
hand. **Python** in `research/` uses **uv** + pytest as wired in `make test`.

## Development workflow and quality gates

Feature work using SpecKit MUST validate against this constitution in `/speckit.specify`,
`/speckit.plan`, `/speckit.tasks`, and `/speckit.analyze`. Implementation plans MUST complete the
Constitution Check gate before implementation; MUST-level violations MUST be documented under plan
**Complexity Tracking** with justification and mitigation.

Pull requests SHOULD name which principles (by Roman numeral) are materially impacted. Reviewers
MUST block merges that violate any MUST in sections I–IV; SHOULD request fixes for SHOULD items in
V–VI unless waived with rationale in the plan or PR.

## Governance

This constitution supersedes conflicting ad-hoc practices for SpecKit and engineering judgment on
this repository.

**Versioning (this document):** Amendments MUST update this file, set **Last Amended** to the
change date (ISO `YYYY-MM-DD`), bump **Version** semantically: MAJOR—removed or incompatible
redefinition of a MUST; MINOR—new principle/section or material new guidance; PATCH—clarifications
only.

**Precedence when principles tension:** **I (Safety)** overrides speed and convenience. **III
(Correctness / secrets)** overrides ergonomics that would log or mishandle sensitive data. Among
non-safety trade-offs, **testability and boundary clarity** override short-term implementation ease
unless the plan records a time-boxed exception with owner and review date.

**Applying principles to technical decisions:** New dependencies, languages, or deployment shapes
MUST be justified against boundaries (II), quality (III), and operational impact (VI). Persistent
deviations from a SHOULD MUST become either an accepted pattern (update constitution) or
time-limited debt in the plan. Ambiguity about whether work falls under a MUST SHOULD be resolved by
a constitution PATCH or team agreement, not silent reinterpretation.

**Ratification and review:** **Compliance** SHOULD be revisited when onboarding contributors, after
incidents, or when core domains (execution, dashboard, exchange integrations) receive large changes.

**Version**: 1.1.0 | **Ratified**: 2026-03-30 | **Last Amended**: 2026-03-30
