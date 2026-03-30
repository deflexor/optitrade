# Optitrade

Go services and tooling for Deribit-focused options workflows, plus an **operator dashboard** (React + Vite SPA with a Go BFF).

## Requirements

- **Go** 1.26+ (see `src/go.mod`)
- **Node.js** 20+ and **npm** (for `web/`)

## Operator dashboard (frontend + API)

The dashboard is two processes in development: the **Vite dev server** serves the UI and proxies `/api` to the Go server.

### 1. Install web dependencies

```bash
make web-install
```

(Use `cd web && npm install` if you prefer not to use a clean `npm ci`.)

### 2. Run Go BFF and SPA (two terminals)

**Terminal A** -- HTTP API + health check (default listen address matches the Vite proxy):

```bash
make run-dashboard
```

Override the listen address if needed:

```bash
make run-dashboard DASHBOARD_LISTEN=127.0.0.1:9090
```

If you change the listen port, update `web/vite.config.ts` proxy `target` to match (or set `OPTITRADE_DASHBOARD_LISTEN`). The embed directory `src/internal/dashboard/dist/` is gitignored; `make run-dashboard` creates a stub there when needed. Plain `go run` only works after `make ensure-dashboard-embed-dir`:

```bash
cd src
go run ./cmd/optitrade dashboard -listen=127.0.0.1:8080
# with auth and custom session DB:
# go run ./cmd/optitrade dashboard -listen=127.0.0.1:8080 \
#   -auth=./dashboard.auth.json -session-db=./dashboard-sessions.sqlite
```

(If `-auth` is omitted and `OPTITRADE_DASHBOARD_AUTH_PATH` is unset, the built-in **opti** / **opti** allowlist is used.)

**Terminal B** -- Vite dev server (HMR):

```bash
make web-dev
```

Open the URL Vite prints (by default **http://127.0.0.1:5173**). Browser calls to `/api/v1/...` go to the Go server on port **8080**.

Sanity checks:

```bash
curl -sS http://127.0.0.1:8080/healthz    # expect: ok
```

### 2b. Operator auth (allowlist + sessions)

Sign-in uses a **JSON allowlist** (username + bcrypt `password_hash`) and a **SQLite file** for server-side sessions (not idle-expired; revoked if you remove a user from the file and reload).

**If you do not set `OPTITRADE_DASHBOARD_AUTH_PATH`**, the server loads a **built-in development allowlist**: username **`opti`**, password **`opti`**. For anything beyond local dev, point at your own auth file (and use strong passwords).

1. (Optional) Create an auth file (schema: `specs/002-dashboard-operator-trading/quickstart.md`).
2. Point the process at it when overriding the default, e.g.:

```bash
export OPTITRADE_DASHBOARD_AUTH_PATH=/path/to/dashboard.auth.json
# Optional: session DB (default if unset: ./optitrade-dashboard.sqlite in the process working directory)
export OPTITRADE_DASHBOARD_SESSION_PATH=/path/to/dashboard-sessions.sqlite
```

Or pass flags (same effect as the env vars):

```bash
make run-dashboard \
  DASHBOARD_AUTH_PATH=./dashboard.auth.json \
  DASHBOARD_SESSION_PATH=./dashboard-sessions.sqlite
```

The Go CLI also accepts **`dashboard -auth=...`** and **`dashboard -session-db=...`**. Changing the allowlist on disk is picked up on **`SIGHUP`** without restarting.

### 3. Build the SPA into the Go embed directory

The directory `src/internal/dashboard/dist/` is **gitignored**. It is filled either by a tiny **stub file** from `make ensure-dashboard-embed-dir` (so `go build` works) or by a real production bundle from the step below.

For a **single binary** that serves the built UI from `go:embed`:

```bash
make dashboard-sync-assets
```

Then build the CLI:

```bash
make build-optitrade
```

Or do both in one step:

```bash
make build-dashboard
```

Run the fat binary (serves static files if assets were synced; otherwise API routes like `/healthz` still work):

```bash
./optitrade dashboard -listen=:8080
# optional: -auth=./dashboard.auth.json -session-db=./sessions.sqlite
# or: ./optitrade --dashboard-listen=:8080
```

### Makefile quick reference

| Command | Purpose |
|--------|---------|
| `make help` | List Makefile targets |
| `make run-dashboard` | Start Go dashboard BFF (default `:8080`; optional `DASHBOARD_AUTH_PATH`, `DASHBOARD_SESSION_PATH`) |
| `make web-dev` | Vite dev server + `/api` proxy |
| `make web-build` | Production build of `web/` |
| `make dashboard-sync-assets` | `web-build` + copy into `src/internal/dashboard/dist/` |
| `make build-dashboard` | Sync assets + build `./optitrade` |
| `make test-web` | Typecheck and bundle SPA (`npm run build`) |
| `make dev-info` | Print the two-terminal dev instructions |

## Testing

From the repository root:

```bash
make test          # Go unit tests + research pytest
make test-web      # Frontend production build (compile check)
```

Dashboard Go packages are included in `go test ./...`. Integration tests for Deribit need credentials; see **Further reading** below.

## CLI beyond the dashboard

Read-only testnet observation, smoke orders, policy paths, and SC test mapping are documented in:

- `docs/quickstart.md`

## Further reading

- `docs/quickstart.md` -- Deribit env, `observe`, `smoke-order`, integration tests
- `specs/002-dashboard-operator-trading/quickstart.md` -- dashboard auth file, curl, two-terminal dev
- `docs/trader-safety-cheatsheet.md` -- operator safety
- `docs/runbook-incident.md` -- incidents (includes dashboard BFF notes)
- `specs/002-dashboard-operator-trading/plan.md` -- dashboard feature plan (branch `002-dashboard-operator-trading`)
