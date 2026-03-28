# Work Packages: Autonomous Deribit Options Robot

**Inputs**: `kitty-specs/001-autonomous-deribit-options-robot/` (spec, plan, data-model, research, contracts, quickstart, docs)  
**Prerequisites**: plan.md, spec.md  
**Tests**: Constitution and spec require automated tests on critical paths (risk, auth, connectivity); integration against testnet or fixtures where noted.

**Organization**: Subtasks `T001`..`T068` roll into work packages `WP01`..`WP13`. Prompts live in `kitty-specs/001-autonomous-deribit-options-robot/tasks/` (flat directory).

**Path conventions** (from plan): `execution/` (Go), `research/` (Python), `config/examples/`, repo `docs/`.

---


## Dependency and execution summary

- **Sequence**: WP01 -> WP02, WP03 (parallel) -> WP04 -> WP05 -> WP06 -> WP07 -> **WP11** (audit core) -> WP08 -> WP09 -> WP10 -> WP12 -> WP13.
- **Parallel after WP01**: WP02 and WP03 may proceed in parallel. **WP11** is sequenced **after WP07** so `spec-kitty implement WP08 --base WP11` always rests on a branch that already includes candidates + regime + persistence.
- **MVP slice**: WP01-WP07 + **WP11** + WP08-WP09 yields **decision + veto logging** on fixtures or testnet read-only data **without** submitting orders. Full trade loop requires WP10-WP12. Resolves G1: FR-010 applies to pre-trade vetoes before execution.

---

## Work Package WP01: Monorepo scaffold and tooling (Priority: P0)

**Goal**: Create `execution/` Go module, `research/` Python package layout, shared `Makefile` or `task` runner, and editor/lint baseline per constitution.  
**Independent Test**: `go build ./...` and `pytest` (empty or smoke) succeed from clean checkout.  
**Prompt**: [tasks/WP01-monorepo-scaffold.md](tasks/WP01-monorepo-scaffold.md)  
**Estimated prompt size**: ~320 lines

**Requirements Refs**: FR-001

### Included Subtasks
- [x] T001 Create directory layout: `execution/cmd/optitrade`, `execution/internal/...`, `research/`, `config/examples/` per plan.md
- [x] T002 Initialize `execution/go.mod` at Go 1.26+; add `toolchain` directive if needed
- [x] T003 [P] Initialize `research/pyproject.toml` with dev deps (`pytest`, `ruff` optional)
- [x] T004 Add root `Makefile` or `justfile` targets: `build`, `test`, `lint` (go vet, staticcheck optional)
- [x] T005 [P] Document local dev in `quickstart.md` cross-links; add `.env.example` without real secrets

### Implementation Notes
Pin dependency versions; no secrets in repo. Follow `.kittify/AGENTS.md` UTF-8 rules for all new markdown.

### Dependencies
- None

### Risks and mitigations
- WSL path quirks: document in quickstart.

---

## Work Package WP02: Config load and policy validation (Priority: P0)

**Goal**: Load operator policy JSON/YAML, validate against `contracts/config-policy.schema.json`, and expose typed config to execution.  
**Independent Test**: Invalid config fails fast with actionable errors; valid example loads.  
**Prompt**: [tasks/WP02-config-policy-loading.md](tasks/WP02-config-policy-loading.md)  
**Estimated prompt size**: ~300 lines

**Requirements Refs**: FR-007, FR-011

### Included Subtasks
- [x] T006 Implement schema embed or load from `kitty-specs/001-autonomous-deribit-options-robot/contracts/config-policy.schema.json` at build or runtime (document choice)
- [x] T007 Implement `internal/config` loader: env overrides for `DERIBIT_*`, `OPTITRADE_CONFIG_PATH`
- [x] T008 Add `config/examples/policy.example.json` aligned with schema (testnet defaults, conservative limits)
- [x] T009 Unit tests: reject unknown fields if schema disallows; reject missing required limits

### Implementation Notes
Constitution: validate all untrusted file input at boundary.

### Dependencies
- Depends on WP01

### Risks and mitigations
- Schema drift: single source under `contracts/`.

---

## Work Package WP03: SQLite persistence (Priority: P0)

**Goal**: Implement schema and repositories per `data-model.md` using parameterized SQL.  
**Independent Test**: Migrations apply; CRUD smoke for `order_record`, `audit_decision`, `risk_policy`.  
**Prompt**: [tasks/WP03-sqlite-persistence.md](tasks/WP03-sqlite-persistence.md)  
**Estimated prompt size**: ~350 lines

**Requirements Refs**: FR-004, FR-010

### Included Subtasks
- [x] T010 Add migration framework (e.g. `goose` or embedded SQL files) under `execution/internal/state/migrations/`
- [x] T011 Create tables: `instrument`, `order_record`, `fill_record`, `position_snapshot`, `risk_policy`, `audit_decision`, `regime_state`, `trade_candidate` as in data-model
- [x] T012 Implement repository interfaces: orders, fills, audit, regime snapshots
- [x] T013 [P] WAL mode and busy_timeout configuration for SQLite
- [x] T014 Repository tests with in-memory or temp file DB

### Implementation Notes
No string concatenation for dynamic SQL from user content.

### Dependencies
- Depends on WP01

### Risks and mitigations
Migration ordering; backup story for ops (document in runbook in WP13).

---

## Work Package WP04: Deribit JSON-RPC and WebSocket client (Priority: P0)

**Goal**: Authenticated HTTP JSON-RPC client and WebSocket subscription manager with reconnect policy.  
**Independent Test**: Auth against testnet succeeds; subscriptions resume after forced disconnect in test harness.  
**Prompt**: [tasks/WP04-deribit-client.md](tasks/WP04-deribit-client.md)  
**Estimated prompt size**: ~380 lines

**Requirements Refs**: FR-001, FR-004

### Included Subtasks
- [x] T015 Implement `internal/deribit/rpc` request/response types, id mapping, error decode
- [x] T016 Implement OAuth/token refresh or key auth per current Deribit API (document in code comments)
- [x] T017 Implement `internal/deribit/ws` with ping/pong, reconnect with backoff, resubscribe
- [x] T018 [P] Map `private/get_positions`, `private/get_open_orders`, `private/get_account_summaries`
- [x] T019 [P] Map `public/get_instruments`, `public/get_order_book`, `public/ticker`
- [x] T020 Integration test: testnet login read-only (skipped if `-short` or missing env)

### Implementation Notes
Never log secrets; redact auth headers in debug.

### Dependencies
- Depends on WP01, WP02

### Risks and mitigations
Rate limits: centralize RPC scheduler if needed.

---

## Work Package WP05: Market data pipeline (Priority: P1)

**Goal**: Discover BTC/ETH options, maintain book/ticker/vol index cache with staleness flags.  
**Independent Test**: Instrument list filtered to options; book updates advance monotonic timestamps in tests.  
**Prompt**: [tasks/WP05-market-data-ingestion.md](tasks/WP05-market-data-ingestion.md)  
**Estimated prompt size**: ~340 lines

**Requirements Refs**: FR-002, FR-003

### Included Subtasks
- [x] T021 Implement instrument discovery and filter (BTC/ETH options only)
- [x] T022 Implement depth-limited order book cache per instrument
- [x] T023 Wire `public/get_volatility_index_data` (or equivalent) for regime inputs
- [x] T024 Expose `MarketSnapshot` view with `quality_flags` (stale, gap) for downstream
- [x] T025 Unit tests with golden JSON fixtures under `tests/fixtures/deribit/`

### Dependencies
- Depends on WP04

### Risks and mitigations
Bounded memory: cap number of subscribed instruments.

---

## Work Package WP06: Regime classifier (Priority: P1)

**Goal**: Rule-based `low` / `normal` / `high` labels with versioned inputs digest.  
**Independent Test**: Given fixture paths, label matches expected table cases.  
**Prompt**: [tasks/WP06-regime-classifier.md](tasks/WP06-regime-classifier.md)  
**Estimated prompt size**: ~280 lines

**Requirements Refs**: FR-005

### Included Subtasks
- [x] T026 Implement `rules_v1` classifier using config thresholds from policy
- [x] T027 Persist `regime_state` row on each evaluation tick (or significant change only; document)
- [x] T028 Deterministic tests: synthetic vol index series -> label
- [x] T029 Wire classifier to market snapshot interface from WP05

### Dependencies
- Depends on WP05, WP03

### Risks and mitigations
Label flapping: add hysteresis window in config if needed.

---

## Work Package WP07: Playbooks, liquidity, and candidates (Priority: P1)

**Goal**: Enforce playbook structures per regime; generate liquid multi-leg candidates only.  
**Independent Test**: No candidate emitted for wide-spread illiquid strikes; structures match allowed templates.  
**Prompt**: [tasks/WP07-candidate-generator.md](tasks/WP07-candidate-generator.md)  
**Estimated prompt size**: ~400 lines

**Requirements Refs**: FR-002, FR-005, FR-011

### Included Subtasks
- [x] T030 Implement liquidity gate: min top size, max spread bps from policy
- [x] T031 Implement playbook resolver: `low`/`normal`/`high` -> allowed structure types
- [x] T032 Implement candidate templates for vertical spreads and iron condor skeletons (BTC/ETH)
- [x] T033 Reject naked short legs always; assert defined-risk invariants
- [x] T034 Unit tests for generator edge cases (empty book, boundary strikes)
- [x] T035 [P] Document Deribit instrument naming for legs in `research.md` appendix or code doc

### Dependencies
- Depends on WP05, WP06, WP02

### Risks and mitigations
Combo naming: align with exchange before placing orders (execution WP10).

---

## Work Package WP11: Audit trail and structured logging (Priority: P1)

**Goal**: Persist `audit_decision` rows and structured logs with correlation IDs; provide `DecisionLogger` so **cost vetoes (WP08)** and **risk vetoes (WP09)** satisfy FR-010 **before** any order submission.  
**Independent Test**: Synthetic cost or risk veto produces one `audit_decision` row with regime, reason fields, and correlation ID.  
**Prompt**: [tasks/WP11-audit-and-observability.md](tasks/WP11-audit-and-observability.md)  
**Estimated prompt size**: ~320 lines

**Requirements Refs**: FR-010

### Included Subtasks
- [x] T051 Implement `audit.DecisionLogger` interface and SQLite persistence for `audit_decision` (used by cost, risk, execution)
- [x] T052 Serialize gate outcomes, cost model version, veto reasons, and `correlation_id` into `audit_decision`
- [x] T053 Configure zap/zerolog JSON with redaction filters for sensitive keys
- [x] T054 Optional: emit `event-envelope.schema.json` compatible JSON lines for external tools

### Implementation Notes
Land **after WP07**, **before WP08**: cost and risk packages accept `DecisionLogger` via constructor or functional options; avoid import cycles (`audit` must not import `strategy`).

### Dependencies
- Depends on WP07

### Risks and mitigations
Logger on hot path: allow sync insert for MVP; batch later if needed.

---

## Work Package WP08: Cost model and edge scoring (Priority: P1)

**Goal**: Net edge after fees, half-spread, slippage, regime-aware adverse selection bump.  
**Independent Test**: Table-driven: known fees and quotes -> veto or approve.  
**Prompt**: [tasks/WP08-cost-model.md](tasks/WP08-cost-model.md)  
**Estimated prompt size**: ~300 lines

**Requirements Refs**: FR-006, FR-011

### Included Subtasks
- [x] T036 Implement fee and half-spread cost components from policy and live quotes
- [x] T037 Implement slippage and adverse-selection adders per regime
- [x] T038 Implement `ScoreCandidate` returning veto reason codes for audit
- [x] T039 Compare IV-based quotes vs book sanity check hook (defer full logic if IV orders off)

### Dependencies
- Depends on WP07, WP11

### Risks and mitigations
Over-optimistic slippage: default conservative bps.

---

## Work Package WP09: Risk engine (Priority: P1)

**Goal**: Pre-trade gates for delta, vega, premium at risk, daily loss, order count caps, time-in-trade.  
**Independent Test**: Matrix test: each gate trips independently; combined snapshot from exchange + DB.  
**Prompt**: [tasks/WP09-risk-engine.md](tasks/WP09-risk-engine.md)  
**Estimated prompt size**: ~380 lines

**Requirements Refs**: FR-004, FR-007, FR-011

### Included Subtasks
- [ ] T040 Build `PortfolioRiskSnapshot` from exchange positions + optional local reconciliation
- [ ] T041 Implement gates: max portfolio delta/vega, max open premium, max orders per instrument
- [ ] T042 Implement daily loss tracker using fills and mark prices (define session boundary)
- [ ] T043 Implement max loss per trade on hypothetical candidate before submit
- [ ] T044 Implement max time-in-trade tracking for open strategies
- [ ] T045 Property or table tests: veto when any limit exceeded

### Dependencies
- Depends on WP03, WP04, WP08, WP11

### Risks and mitigations
Exchange vs internal PnL mismatch: log diff; constitution suggests future margin compare.

---

## Work Package WP10: Order execution and reconciliation (Priority: P1)

**Goal**: Post-first execution, combo placement where supported, reduce-only exits, cancel-by-label.  
**Independent Test**: Simulated or testnet dry-run records intended orders; reconciliation detects drift.  
**Prompt**: [tasks/WP10-order-execution.md](tasks/WP10-order-execution.md)  
**Estimated prompt size**: ~420 lines

**Requirements Refs**: FR-004, FR-008

### Included Subtasks
- [ ] T046 Implement order placement path: post-only default, label assignment, idempotency key strategy
- [ ] T047 Implement combo/multi-leg placement mapping from candidate legs
- [ ] T048 Implement cancel-by-label and cancel-all-for-instrument wrappers
- [ ] T049 Implement fill handler updating local `order_record` and exposure caches
- [ ] T050 Reconciliation loop: periodic diff open orders vs exchange

### Dependencies
- Depends on WP04, WP09, WP11

### Risks and mitigations
Partial fills: document handling; MVP may flatten-on-partial policy flag.

---

## Work Package WP12: Protective mode and session FSM (Priority: P1)

**Goal**: FR-009 behaviors: feed stale, auth failure, book gap -> no new risk-increasing orders; flatten policy.  
**Independent Test**: Simulated disconnect triggers FSM within internal deadlines; asserts no new opens.  
**Prompt**: [tasks/WP12-protective-mode-fsm.md](tasks/WP12-protective-mode-fsm.md)  
**Estimated prompt size**: ~360 lines

**Requirements Refs**: FR-009

### Included Subtasks
- [ ] T055 Implement session FSM: `running`, `paused`, `protective_flatten`, `frozen`
- [ ] T056 Wire triggers from WS health, RPC auth errors, book gap detector from WP05
- [ ] T057 Block new risk-increasing submissions while in protective paths; allow reduce-only per policy
- [ ] T058 Integration test: feed staleness threshold breaches -> state transition
- [ ] T059 Document 60s detection target from spec SC-003 as NFR test where measurable

### Dependencies
- Depends on WP05, WP10, WP11

### Risks and mitigations
Race with execution: single-writer for session state or mutex documented.

---

## Work Package WP13: Integration, examples, and operator runbook (Priority: P2)

**Goal**: End-to-end testnet checklist, `policy.example.json`, update `docs/trader-safety-cheatsheet.md` links, merge validation, **and explicit acceptance coverage for spec success criteria SC-001-SC-005 (G2)**.  
**Independent Test**: Operator can follow `quickstart.md` on testnet without code changes; SC matrix traceable from `tasks.md` to tests or documented manual steps.  
**Prompt**: [tasks/WP13-integration-and-runbooks.md](tasks/WP13-integration-and-runbooks.md)  
**Estimated prompt size**: ~420 lines

**Requirements Refs**: FR-001, FR-002, FR-003, FR-004, FR-005, FR-006, FR-007, FR-008, FR-009, FR-010, FR-011

### Success criteria acceptance (spec.md)

| ID | Subtask |
|----|---------|
| SC-001 | T064 |
| SC-002 | T065 |
| SC-003 | T066 |
| SC-004 | T067 |
| SC-005 | T068 |

### Included Subtasks
- [ ] T060 E2E harness: read-only session on testnet (positions, books) behind build tag or env
- [ ] T061 Optional gated live order smoke on testnet with minimal size (explicit env `OPTITRADE_ALLOW_TESTNET_ORDERS`)
- [ ] T062 Update `quickstart.md` and `docs/trader-safety-cheatsheet.md` with final binary names and flags
- [ ] T063 Add `docs/runbook-incident.md` skeleton: kill, flatten, key rotation
- [ ] T064 **SC-001**: Automated or scripted certification with seeded data: 100% of simulated fills are defined-risk templates allowed by active playbook (assert no naked short); document command under test/CI target
- [ ] T065 **SC-002**: Stress test: after daily loss cap, assert no further risk-increasing intents (no new opening orders in harness); document session boundary alignment with WP09
- [ ] T066 **SC-003**: Feed-loss or auth-failure simulation: assert protective/block path within 60s budget and zero new risk-increasing submits afterward; link to WP12 tests or extend
- [ ] T067 **SC-004**: Reconciliation acceptance: after scripted window, assert no unexplained orphan legs vs exchange mock/fixture; reference bounded procedure in runbook
- [ ] T068 **SC-005**: Audit sampling test: sample of decisions includes regime label, cost model version, and risk gate outcome; measure 90% threshold on fixture corpus (100% for mock "warning breach" cases)

### Dependencies
- Depends on WP12

### Risks and mitigations
Accidental mainnet: default config `environment: testnet` in example.

---

## Subtask index

**Implementation note**: Complete **WP11 (T051-T054)** after WP07 and **before** WP08 so task IDs are not monotonic in wall-clock order.

| ID | Summary | WP |
|----|---------|-----|
| T001-T005 | Scaffold | WP01 |
| T006-T009 | Config | WP02 |
| T010-T014 | SQLite persistence | WP03 |
| T015-T020 | Deribit client | WP04 |
| T021-T025 | Market data | WP05 |
| T026-T029 | Regime | WP06 |
| T030-T035 | Candidates | WP07 |
| T051-T054 | Audit (before cost/risk wiring) | WP11 |
| T036-T039 | Cost model | WP08 |
| T040-T045 | Risk engine | WP09 |
| T046-T050 | Execution | WP10 |
| T055-T059 | Protective FSM | WP12 |
| T060-T068 | Integration, runbooks, SC acceptance | WP13 |

---

## FR coverage checklist

- FR-001: WP04  
- FR-002: WP05, WP07  
- FR-003: WP05, WP06  
- FR-004: WP03, WP04, WP09, WP10  
- FR-005: WP06, WP07  
- FR-006: WP08  
- FR-007: WP02, WP09  
- FR-008: WP10  
- FR-009: WP12  
- FR-010: WP03, WP11 (WP08/WP09 call logger from WP11)  
- FR-011: WP02, WP07, WP08, WP09  
