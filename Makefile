.PHONY: all
all: lint test

.PHONY: ci
ci: test

GO ?= go

LENGTHS = -short ""

.PHONY: test
test:
	@set -e; \
	for length in $(LENGTHS); do \
		printf "\e[1m$(GO) test \e[32m$$length $(TESTFLAGS) ./...\e[0m\n"; \
		$(GO) test $$length $(TESTFLAGS) ./...; \
	done

GOCACHE = $(shell go env GOCACHE)
GOMODCACHE = $(shell go env GOMODCACHE)

DOCKERRUN = docker run --rm \
	-w /app \
	-v $(PWD):/app \
	-v $(GOCACHE):/root/.cache/go-build \
	-v $(GOMODCACHE):/go/pkg/mod

.PHONY: lint
lint:
	$(DOCKERRUN) golangci/golangci-lint:v1.60.1-alpine \
		golangci-lint run

.PHONY: bench
bench:
	$(GO) test -run=^$$ -bench=. -cpuprofile=cpu.prof -memprofile=mem.prof -benchmem | perl -pe 's/[ \t]+/\t/g if m{allocs/op$$}'
