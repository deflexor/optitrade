ROOT := $(CURDIR)

.PHONY: build test lint

build:
	cd "$(ROOT)/execution" && go build ./...

test:
	cd "$(ROOT)/execution" && go test ./...
	cd "$(ROOT)/research" && uv run --with pytest pytest -q tests/

# Live testnet: requires DERIBIT_CLIENT_ID / DERIBIT_CLIENT_SECRET (see docs/quickstart.md).
test-integration:
	cd "$(ROOT)/execution" && go test -tags=integration ./internal/observe/...

lint:
	cd "$(ROOT)/execution" && go vet ./...
