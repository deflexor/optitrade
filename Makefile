ROOT := $(CURDIR)
SRC_DIR := $(ROOT)/src
WEB_DIR := $(ROOT)/web
DASHBOARD_EMBED_DIST := $(SRC_DIR)/internal/dashboard/dist
OPTITRADE_BIN := $(ROOT)/optitrade
# Must match web/vite.config.ts server.proxy['/api'].target (host:port)
DASHBOARD_LISTEN ?= 127.0.0.1:8080
# Optional: allowlist JSON and session DB for operator login (see README). Empty = rely on OPTITRADE_* env or CLI defaults.
DASHBOARD_AUTH_PATH ?=
DASHBOARD_SESSION_PATH ?=

.PHONY: help build test test-integration lint test-e2e \
	ensure-dashboard-embed-dir \
	web-install web-build web-dev test-web \
	run-dashboard dev-info dashboard-sync-assets build-optitrade build-dashboard

help:
	@echo "Optitrade Makefile"
	@echo ""
	@echo "Dashboard (frontend + Go BFF):"
	@echo "  make run-dashboard          Go BFF + API on DASHBOARD_LISTEN (default $(DASHBOARD_LISTEN))"
	@echo "    Optional: DASHBOARD_AUTH_PATH=... DASHBOARD_SESSION_PATH=... (or OPTITRADE_DASHBOARD_AUTH_PATH / OPTITRADE_DASHBOARD_SESSION_PATH)"
	@echo "  make web-dev                Vite dev server (run web-install once first)"
	@echo "  make dev-info               Print two-terminal dev instructions"
	@echo "  web-install / web-build     npm ci / production build in web/"
	@echo "  make dashboard-sync-assets  web-build + copy dist -> $(DASHBOARD_EMBED_DIST)"
	@echo "  make build-optitrade        go build ./cmd/optitrade -> ./optitrade"
	@echo "  make build-dashboard        dashboard-sync-assets + build-optitrade"
	@echo ""
	@echo "Tests and quality:"
	@echo "  make test / test-integration / lint / test-web / test-e2e"
	@echo "    test-e2e runs Playwright in web/ (requires web-install / npm ci in web/)"
	@echo ""
	@echo "Library build:"
	@echo "  make build                  go build ./... under src/"
	@echo "  ensure-dashboard-embed-dir  ensure src/internal/dashboard/dist has a stub file for go:embed"

# go:embed all:dist requires at least one file; dist/ is gitignored (see .gitignore).
ensure-dashboard-embed-dir:
	@mkdir -p "$(DASHBOARD_EMBED_DIST)"
	@if [ -z "$$(find "$(DASHBOARD_EMBED_DIST)" -mindepth 1 -type f 2>/dev/null | head -n1)" ]; then \
		printf '%s\n' 'placeholder for go:embed; replace with output of make dashboard-sync-assets' > "$(DASHBOARD_EMBED_DIST)/.goembed-placeholder"; \
	fi

build: ensure-dashboard-embed-dir
	cd "$(SRC_DIR)" && go build ./...

build-optitrade: ensure-dashboard-embed-dir
	cd "$(SRC_DIR)" && go build -o "$(OPTITRADE_BIN)" ./cmd/optitrade

test: ensure-dashboard-embed-dir
	cd "$(SRC_DIR)" && go test ./...
	cd "$(ROOT)/research" && uv run --with pytest pytest -q tests/

# Live testnet: requires DERIBIT_CLIENT_ID / DERIBIT_CLIENT_SECRET (see docs/quickstart.md).
test-integration: ensure-dashboard-embed-dir
	cd "$(SRC_DIR)" && go test -tags=integration ./internal/observe/...

lint: ensure-dashboard-embed-dir
	cd "$(SRC_DIR)" && go vet ./...

web-install:
	cd "$(WEB_DIR)" && npm ci

web-build: web-install
	cd "$(WEB_DIR)" && npm run build

test-web: web-install
	cd "$(WEB_DIR)" && npm run build

# Playwright: starts BFF + Vite via web/playwright.config.ts (run from clean shell; see quickstart).
test-e2e: web-install
	cd "$(WEB_DIR)" && npm run test:e2e

web-dev:
	cd "$(WEB_DIR)" && npm run dev

run-dashboard: ensure-dashboard-embed-dir
	cd "$(SRC_DIR)" && go run ./cmd/optitrade dashboard -listen=$(DASHBOARD_LISTEN) \
		$(if $(DASHBOARD_AUTH_PATH),-auth=$(DASHBOARD_AUTH_PATH)) \
		$(if $(DASHBOARD_SESSION_PATH),-session-db=$(DASHBOARD_SESSION_PATH))

dev-info:
	@echo "Development dashboard (two terminals, same repo root):"
	@echo "  1) make run-dashboard  (optional auth file; default dev login opti/opti)"
	@echo "  2) make web-dev"
	@echo "Then open the Vite URL (default http://127.0.0.1:5173). API proxy targets $(DASHBOARD_LISTEN)."

# Copies Vite output into the Go embed tree for: go run ./cmd/optitrade dashboard ...
dashboard-sync-assets: web-build
	rm -rf "$(DASHBOARD_EMBED_DIST)"
	mkdir -p "$(DASHBOARD_EMBED_DIST)"
	cp -a "$(WEB_DIR)/dist/." "$(DASHBOARD_EMBED_DIST)/"

build-dashboard: dashboard-sync-assets build-optitrade
