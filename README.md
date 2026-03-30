# Optitrade

Go services and tooling for Deribit-focused options workflows, plus an **operator dashboard** (React + Vite SPA with a Go BFF).

## Requirements

- **Go** 1.22+ (see `src/go.mod`)
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
```

**Terminal B** -- Vite dev server (HMR):

```bash
make web-dev
```

Open the URL Vite prints (by default **http://127.0.0.1:5173**). Browser calls to `/api/v1/...` go to the Go server on port **8080**.

Sanity checks:

```bash
curl -sS http://127.0.0.1:8080/healthz    # expect: ok
```

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
# or: ./optitrade --dashboard-listen=:8080
```

### Makefile quick reference

| Command | Purpose |
|--------|---------|
| `make help` | List Makefile targets |
| `make run-dashboard` | Start Go dashboard BFF (default `:8080`) |
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
- `docs/trader-safety-cheatsheet.md` -- operator safety
- `docs/runbook-incident.md` -- incidents
- `plan.md` -- technical brief at repo root (when maintained)
