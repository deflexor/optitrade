# Quickstart: Operator Trading Dashboard

**Feature**: `002-operator-trading-dashboard`

## Prerequisites

- Go 1.26+ (match repo `src/go.mod`)
- Node.js 20+ and npm or pnpm (for `web/`)
- Running or mock `optitrade` process with dashboard HTTP enabled (flags TBD in implementation)

## Local development (two terminals)

1. **API and static (once implemented)**  
   From repo root, start the bot binary with dashboard listen address, for example:

   ```bash
   export OPTITRADE_STATE_DB=/tmp/optitrade-state.db
   export OPTITRADE_DASHBOARD_ALLOWLIST=your_username
   cd src && go run ./cmd/optitrade dashboard -listen=127.0.0.1:8080
   ```

   `OPTITRADE_STATE_DB` must point at the shared SQLite file (migrations add `dashboard_*` tables). Use `-state-db` to override the env var.

2. **SPA dev server**  
   From `web/`:

   ```bash
   cd web
   npm install
   npm run dev
   ```

   Configure Vite `server.proxy` so `/api` forwards to `http://127.0.0.1:8080` (or the chosen port). Axios MUST use `withCredentials: true` so cookies flow on login.

## First-time operator setup

1. Register a username/password via the UI (or `POST /api/v1/auth/register`).
2. Add that username to the server allowlist (env or config), e.g. `OPTITRADE_DASHBOARD_ALLOWLIST=opti`.
3. Sign in; you should receive a session cookie and see `/api/v1/snapshot` return 200.

## Sanity checks

- Non-allowlisted user login returns JSON with `Sorry, feature not ready` and no session.
- Stop the snapshot source or delay responses: UI shows stale state when `snapshot_utc` age exceeds 5 seconds per spec.
- OpenAPI: `kitty-specs/002-operator-trading-dashboard/contracts/dashboard-api.openapi.yaml`
