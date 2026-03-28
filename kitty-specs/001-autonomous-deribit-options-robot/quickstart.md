# Quickstart: Autonomous Deribit Options Robot (development)

**Feature**: `001-autonomous-deribit-options-robot`  
**Date**: 2026-03-28

## Prerequisites

- Go 1.22+ (execution)
- Python 3.12+ (research)
- SQLite 3
- Deribit **testnet** account and API key with trade permission (use testnet for all first runs)

## Environment

Never commit secrets. Use:

- `DERIBIT_CLIENT_ID`
- `DERIBIT_CLIENT_SECRET`

(Optional) `OPTITRADE_CONFIG_PATH` pointing to a JSON file validated by `contracts/config-policy.schema.json`.

## Build execution (when code exists)

```bash
cd execution
go build -o ../bin/optitrade ./cmd/optitrade
```

## Run (placeholder)

Until implementation lands:

1. Copy `config/examples/policy.example.json` (to be added in implementation) and fill limits.
2. Run `./bin/optitrade -config "$OPTITRADE_CONFIG_PATH"` against **testnet**.

## Research workflow

```bash
cd research
uv sync || pip install -e .
pytest
```

Use offline data or sanitized fixtures under `tests/fixtures/deribit/` only.

## Verify constitution-aligned checks

- Run `go test ./...` in `execution/` before any PR.
- Confirm logs redact secrets (manual spot check).
- Run schema validation on config in CI when wired.

## Support

- Spec: `kitty-specs/001-autonomous-deribit-options-robot/spec.md`
- Plan: `kitty-specs/001-autonomous-deribit-options-robot/plan.md`
- Trader safety and account setup: `kitty-specs/001-autonomous-deribit-options-robot/trader-safety-cheatsheet.md`
