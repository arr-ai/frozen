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

.PHONY: vet
vet:
	@printf "\e[1m$(GO) test \e[32m-tags frozen_vet -short $(TESTFLAGS) ./...\e[0m\n"
	@$(GO) test -tags frozen_vet -short $(TESTFLAGS) ./...

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

BENCHSTAT   ?= benchstat
BENCH_DIR    = benchmarks
BENCH_COUNT ?= 6

.PHONY: bench-ops
bench-ops:
	$(GO) test -run=^$$ -bench=BenchmarkSetOps -benchmem -count=$(BENCH_COUNT) .

.PHONY: bench-save
bench-save:
	@mkdir -p $(BENCH_DIR)
	$(GO) test -run=^$$ -bench=BenchmarkSetOps -benchmem -count=$(BENCH_COUNT) . \
		| tee $(BENCH_DIR)/$$(date +%Y-%m-%d)-$$(git rev-parse --short HEAD).txt

.PHONY: bench-compare
bench-compare:
	@if [ -n "$(OLD)" ] && [ -n "$(NEW)" ]; then \
		$(BENCHSTAT) $(OLD) $(NEW); \
	else \
		files=$$(ls -t $(BENCH_DIR)/*.txt 2>/dev/null | head -2); \
		count=$$(echo "$$files" | wc -l | tr -d ' '); \
		if [ "$$count" -lt 2 ]; then \
			echo "Need at least 2 result files in $(BENCH_DIR)/. Run 'make bench-save' twice."; \
			exit 1; \
		fi; \
		new=$$(echo "$$files" | head -1); \
		old=$$(echo "$$files" | sed -n '2p'); \
		echo "Comparing $$(basename $$old) → $$(basename $$new):"; \
		$(BENCHSTAT) $$old $$new; \
	fi
