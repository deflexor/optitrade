# Quickstart: Dashboard feature development

**Branch**: `002-dashboard-operator-trading`

## Prerequisites

- Go **1.26+** (`src/go.mod`), Node **20+**, `make`.
- Deribit credentials for **testnet** when exercising live handlers (optional for auth-only milestones).

## Run BFF + SPA (two terminals)

From repo root:

```bash
make run-dashboard
```

```bash
make web-dev
```

Open the URL Vite prints (default `http://127.0.0.1:5173`). API calls go to `/api/v1` via the Vite proxy to the Go listener (see `README.md`).

Health:

```bash
curl -sS http://127.0.0.1:8080/healthz
```

## Configuration (auth)

**Default (no file):** if `OPTITRADE_DASHBOARD_AUTH_PATH` is unset, the binary embeds a dev allowlist — sign in as **`opti` / `opti`**. Use a real auth file for non-local deployments.

1. (Optional) Create a dashboard auth JSON file (see `research.md` R-01):

   - `version`, `users`: array of `{ "username", "password_hash" }` (bcrypt).
   - Generate `password_hash` with bcrypt tooling.

2. Point the process at it when not using the default:

   ```bash
   export OPTITRADE_DASHBOARD_AUTH_PATH=/path/to/dashboard.auth.json
   ```

3. Restart `dashboard` after allowlist edits, or send **`SIGHUP`** when using a file; existing sessions for removed users must fail on the next request.

## Cookies / CORS

- Dev: same-site between Vite and BFF may require proxy-only API calls (no CORS) — **preferred**.
- If the SPA is served from another origin in dev, BFF must issue `Secure` cookies only when HTTPS and set explicit `CORS` + credentials — treat as **non-default**; document if enabled.

## Tests

```bash
make test
make test-web
make lint
```

Add handler tests under `src/internal/dashboard/` as endpoints gain behavior.

## API contract

See [`contracts/openapi.yaml`](./contracts/openapi.yaml) for paths and shapes.
