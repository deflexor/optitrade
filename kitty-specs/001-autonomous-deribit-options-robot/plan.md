# Implementation Plan: Autonomous Deribit Options Robot

*Path: [.kittify/missions/software-dev/templates/plan-template.md](.kittify/missions/software-dev/templates/plan-template.md)*

**Branch**: `master` (planning repo; feature slug `001-autonomous-deribit-options-robot`) | **Date**: 2026-03-28 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification at `kitty-specs/001-autonomous-deribit-options-robot/spec.md`, aligned with root `plan.md` technical brief.

## Summary

Deliver an autonomous Deribit BTC/ETH **options** system that: (1) ingests exchange market and account state, (2) classifies **low / normal / high** vol regime and selects a **defined-risk** playbook per regime, (3) generates liquid candidates, scores **net edge after fees, spread, and slippage**, (4) enforces **hard Greek and loss limits** before any order, (5) executes with **post-only preference** and structured exits, and (6) enters **protective flatten or freeze** on feed, auth, or book-quality anomalies within the success-criteria time budget. Use a **two-layer** architecture: Python research/backtest vs **Go** execution, **SQLite** for durable positions/orders/audit, structured logging without secrets.

## Technical Context

**Language/Version**: Go 1.22+ (execution service); Python 3.12+ (research, backtests, regime playbook prototyping)  
**Primary Dependencies**: Go: `net/http` or `resty` for HTTPS JSON-RPC; WebSocket client (`gorilla/websocket` or `nhooyr/websocket`); `zap` or `zerolog` for logs; `modernc.org/sqlite` or `sqlite` driver with parameterized queries. Python: `pandas`, `numpy`, `pytest` for research path.  
**Storage**: SQLite (positions, orders, fills, daily PnL aggregates, audit decision records)  
**Testing**: Go `go test` (table-driven for risk gates, cost model, regime labels with fake clock); Python `pytest` for research/backtest; integration tests against Deribit **testnet** or recorded **golden** JSON-RPC fixtures (no secrets in repo).  
**Target Platform**: Linux (containers or bare metal); operator CLI or systemd-run binary for MVP.  
**Project Type**: Single repo, **two deployable concern**s: `research/` (offline) and `execution/` (online).  
**Performance Goals**: Protective mode within **60 seconds** of anomaly detection (per spec SC-003); decision tick and risk evaluation MUST NOT grow without bound (bounded worker pools, capped order-book depth ingestion per project constitution). Hot path: O(1) or O(legs) per candidate with documented max legs.  
**Constraints**: No naked short options; no martingale; secrets via env or secret manager, never logged; IV-based quotes must be corroborated with order book under fast-move detection (spec edge case).  
**Scale/Scope**: Single operator account per deployment MVP; BTC and ETH options only; liquid strikes per configurable depth/spread thresholds.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle (from `.kittify/memory/constitution.md`) | How this plan complies |
|----------------------------------------------------|-------------------------|
| Testing MUST cover critical paths (risk, auth, data integrity) | Risk gate matrix, cost model, regime classifier, and anomaly FSM have **automated** tests; integration suite for order lifecycle on testnet or fixtures. |
| MUST NOT log secrets | Logging design uses redaction for API keys, client secrets, full tokens; audit logs store **decision metadata** only. |
| Parameterized DB access | All SQLite access via prepared statements / ORM; no string-built SQL from config or exchange payloads. |
| Performance: no unbounded hot-path work | Worker pools with fixed concurrency; bounded book depth; explicit caps on candidate enumeration documented in `research.md`. |
| Public APIs documented | Exported Go packages and any CLI flags documented; config validated against `contracts/config-policy.schema.json`. |

**Post-Phase 1 re-check**: PASS. No new MUST violations; operator UI is out of scope for MVP (CLI + logs) so UX consistency applies when/if a dashboard is added.

## Project Structure

### Documentation (this feature)

```
kitty-specs/001-autonomous-deribit-options-robot/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
└── tasks.md             # Phase 2 only: created by /spec-kitty.tasks
```

### Source Code (repository root)

```
execution/                    # Go: live connectivity, risk, execution
├── go.mod
├── cmd/optitrade/            # main binary
└── internal/
    ├── deribit/              # JSON-RPC + WS client
    ├── market/               # book, ticker, vol index ingestion
    ├── regime/               # classification (may start rule-based)
    ├── strategy/             # playbooks, candidate gen
    ├── risk/                 # gates, portfolio snapshot
    ├── exec/                 # order placement, labels, reduce-only
    ├── state/                # SQLite repos
    └── audit/                # structured decision logs

research/                     # Python: backtests, parameter studies
├── pyproject.toml
└── src/optitrade_research/

config/
└── examples/                 # sample policy YAML validated by schema

tests/                        # optional top-level shared fixtures
└── fixtures/deribit/         # recorded RPC (sanitized)
```

**Structure Decision**: **Option 1 variant**: single repo with `execution/` (Go) and `research/` (Python). Research informs parameters and playbooks; execution is the only component that touches live keys for MVP.

## Complexity Tracking

*No constitution violations requiring justification. Complexity (two languages) is inherited from product brief: Python for research velocity, Go for production execution.*

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| N/A | | |

## Phase Outputs (this command)

| Phase | Artifact | Path |
|-------|-----------|------|
| 0 | Research decisions | `kitty-specs/001-autonomous-deribit-options-robot/research.md` |
| 1 | Data model | `kitty-specs/001-autonomous-deribit-options-robot/data-model.md` |
| 1 | Contracts | `kitty-specs/001-autonomous-deribit-options-robot/contracts/` |
| 1 | Quickstart | `kitty-specs/001-autonomous-deribit-options-robot/quickstart.md` |

**Stop**: Task generation is **not** performed here. Run `/spec-kitty.tasks` next.
