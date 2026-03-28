# Optitrade Constitution

> Established via `/spec-kitty.constitution` with explicit principle themes  
> Created: 2026-03-28  
> Version: 1.0.0

## Purpose

This constitution defines project-wide principles for **code quality**, **safety**, **testing**, **user experience consistency**, **clarity** (comments and docstrings), and **performance**. Every feature specification, plan, task, and pull request MUST align with these principles unless an approved exception is documented per [Governance](#governance).

## Technical standards

### Languages and frameworks

- **MUST** follow the languages, runtimes, and frameworks the repository declares as authoritative (root `README`, package manifests, or mission docs). When ambiguity exists, resolve it in the feature spec before implementation.
- **SHOULD** prefer supported LTS or stable releases and pin dependency versions to keep builds reproducible.

### Testing requirements

- **MUST** include automated tests for new behavior and for regressions fixed by a change. Tests MUST fail when the behavior they guard is broken.
- **MUST** cover critical paths (auth, billing, data integrity, safety-relevant logic, and irrevocable actions) with tests that assert explicit outcomes, not only “happy path” smoke checks.
- **SHOULD** maintain a fast default test suite suitable for pre-push or CI; slower tests **SHOULD** be labeled or separated so developers can run them deliberately.
- **SHOULD** favor deterministic tests. Nondeterminism (time, network, randomness) **MUST** be controlled via fakes, clocks, or seeds documented in the test.

### Performance and scale

- **MUST** state measurable performance expectations in the feature spec or plan when user-visible latency, throughput, or resource limits are part of the requirement (for example: p95 response time, batch size, memory ceiling).
- **MUST NOT** introduce unbounded work on hot paths (unbounded queries, unbounded in-memory growth, blocking I/O on the UI thread) without explicit justification and mitigation in the design.
- **SHOULD** profile or measure before optimizing; **SHOULD** document any known trade-offs when choosing a faster-but-less-clear implementation.

### Deployment and constraints

- **MUST** respect environment and deployment constraints documented for the project (containers, platforms, secrets handling). Until documented, **SHOULD** assume production builds are reproducible from source and do not rely on developer-only paths.

## Code quality and safety

### Code quality

- **MUST** keep changes scoped and reviewable: unrelated refactors **SHOULD** land as separate commits or PRs.
- **MUST** preserve or improve type safety and static checks where the project uses them; new code **SHOULD NOT** widen `any`-style escapes without justification.
- **SHOULD** follow the project’s formatter and linter configuration; missing configuration **SHOULD** be added rather than ignored long-term.

### Safety and trust

- **MUST** validate and encode untrusted input at trust boundaries (HTTP APIs, CLIs reading files, user-provided config, web forms).
- **MUST NOT** log secrets, tokens, passwords, or full payment identifiers. Redact or omit sensitive fields in logs and diagnostics.
- **MUST** use parameterized queries or vetted ORM/query builders for database access; **MUST NOT** concatenate raw user input into queries or commands.
- **SHOULD** apply least privilege for credentials and service accounts introduced by a feature.

## User experience consistency

- **MUST** match existing patterns for navigation, terminology, errors, loading states, and accessibility within the same product surface unless the spec explicitly introduces a new pattern with migration rationale.
- **MUST** provide user-visible feedback for asynchronous or long-running operations (loading, success, retry, failure).
- **SHOULD** meet baseline accessibility expectations for the platform (focus order, labels for controls, sufficient contrast) consistent with prior screens in the app.
- **SHOULD** keep copy actionable: errors **SHOULD** say what failed and what the user can do next when safe to do so.

## Clarity: comments and docstrings

- **MUST** document public APIs (functions, types, endpoints, CLI commands) with docstrings or equivalent contract documentation describing purpose, inputs, outputs, errors, and invariants material to callers.
- **MUST** comment non-obvious logic: algorithms, concurrency, security-sensitive branches, and compatibility hacks **SHOULD** state *why* the code exists, not restate *what* the syntax does.
- **SHOULD** keep comments current: **MUST** update or remove comments when behavior changes.
- **SHOULD** avoid noise: do not add comments that duplicate names or obvious control flow.

## Quality gates (merge readiness)

Before merging, changes **MUST**:

- Pass the project’s automated test suite and required checks.
- Satisfy this constitution’s Testing, Safety, and Clarity rules for the scope of the change.
- Include UX-consistent behavior for user-facing changes, or document intentional deviations in the PR.

**SHOULD** additionally pass linters, formatters, and type checks where configured.

## Governance

### Amendment process

Any team member **MAY** propose amendments via pull request. Material changes to MUST-level rules **SHOULD** allow time for review by stakeholders who own quality, security, or UX.

### Compliance validation

Code reviewers **MUST** treat this constitution as binding for features in scope. Conflicts **SHOULD** be resolved by adjusting the spec, plan, or implementation—**MUST NOT** silently ignore a MUST.

### Exception handling

Exceptions **MUST** be rare, justified in writing (PR description or linked note), and include who approved them and any sunset or follow-up task. Recurring exceptions **SHOULD** trigger a constitution update instead of permanent waivers.
