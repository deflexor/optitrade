# Optitrade quickstart (Deribit testnet)

**Feature**: `001-autonomous-deribit-options-robot`  
**Last updated**: 2026-03-28

This guide targets operators who want a **read-only testnet** path first. Defaults assume **Deribit testnet** RPC (`https://test.deribit.com/api/v2`).

## 1. Build the binary

From the repository root (this worktree):

```bash
cd src
go build -o optitrade ./cmd/optitrade
```

The CLI binary name is **`optitrade`**.

## 2. Configure credentials

Copy `.env.example` to `.env` and set:

- `DERIBIT_CLIENT_ID` / `DERIBIT_CLIENT_SECRET` — testnet API key pair (no withdrawal scope).
- Optional: `DERIBIT_BASE_URL` — override JSON-RPC base (default is testnet).

Load them into your shell (example):

```bash
set -a && source ../.env && set +a
```

## 3. Read-only observation (T060 / E2E)

Poll **private positions** and **public order book** for one instrument for a bounded duration:

```bash
./optitrade observe -duration 45s -interval 3s -instrument BTC-PERPETUAL
```

Exits **0** when `-duration` elapses. Requires valid testnet credentials.

**CI / integration**: with the same env vars set:

```bash
cd src && go test -tags=integration ./internal/observe/...
```

## 4. Optional: tiny testnet order smoke (T061)

**Default is off.** Sending any order requires **both**:

1. `OPTITRADE_ALLOW_TESTNET_ORDERS=1`
2. A policy file whose `"environment"` field is exactly `"testnet"` (see `config/examples/policy.example.json`).

```bash
export OPTITRADE_POLICY_PATH="$(pwd)/../config/examples/policy.example.json"
export OPTITRADE_ALLOW_TESTNET_ORDERS=1
./optitrade smoke-order -instrument BTC-PERPETUAL -amount 10 -policy "$OPTITRADE_POLICY_PATH"
```

This places a **deep out-of-the-market post-only** bid, then **cancels all orders** on that instrument. It is still real testnet exposure—use minimal size and inspect the book first.

## 5. Policy and schema

- Example policy: `config/examples/policy.example.json`
- JSON Schema (embedded in the binary for validation): `src/internal/config/policy.schema.json`

## 6. Spec success criteria tests (G2 / SC-001–SC-005)

Automated checks live next to the relevant packages (run `make test` from repo root):

| Spec   | Package        | Test name prefix / note        |
|--------|----------------|--------------------------------|
| SC-001 | `internal/strategy` | `TestSC001_*`              |
| SC-002 | `internal/risk`     | `TestSC002_*`, `TestEngineDailyLossVeto` |
| SC-003 | `internal/session`  | `TestSC003_*`              |
| SC-004 | `internal/execution` | `TestSC004_*`           |
| SC-005 | `internal/audit`    | `TestSC005_*`              |

## 7. Further reading

- Operator safety: `docs/trader-safety-cheatsheet.md`
- Incidents: `docs/runbook-incident.md`
- Technical brief: root `plan.md`
