ROOT := $(CURDIR)

.PHONY: build test lint

build:
	cd "$(ROOT)/execution" && go build ./...

test:
	cd "$(ROOT)/execution" && go test ./...
	cd "$(ROOT)/research" && uv run --with pytest pytest -q tests/

lint:
	cd "$(ROOT)/execution" && go vet ./...
