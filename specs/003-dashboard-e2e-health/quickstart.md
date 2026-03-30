# Quickstart: verify dashboard robustness work

## Development (two processes)

From repo root (`/home/dfr/optitrade`):

1. `make run-dashboard` — BFF on `DASHBOARD_LISTEN` (default `127.0.0.1:8080`).
2. `make web-dev` — Vite on `127.0.0.1:5173` (proxy `/api` to BFF per `web/vite.config.ts`).

Default dev credentials (when allowlist not overridden): see project README / auth file conventions.

## End-to-end tests (Playwright)

Chromium needs OS libraries (for example on Debian/Ubuntu/WSL: install packages from `npx playwright install-deps chromium`, which typically requires sudo). If the browser fails to start with errors about missing `.so` files, run that command once on the machine.

## Full automated gate (spec **FR-007** / **SC-006**)

After all tasks are done, run **once** from repo root:

```bash
make test && make lint && make test-web && make test-e2e
```

Playwright starts its own BFF + Vite via `web/playwright.config.ts`; local dev servers need not be running.

## Optional integration tests

```bash
make test-integration
```

Requires Deribit credentials per project docs — not part of the default CI gate unless configured.
