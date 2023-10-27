.PHONY: all
all: lint test

GO ?= go

LENGTHS = -short ""

.PHONY: test
test:
	@set -e; \
	for length in $(LENGTHS); do \
		printf "\e[1m$(GO) test \e[32m$$length $(TESTFLAGS) ./...\e[0m\n"; \
		$(GO) test $$length $(TESTFLAGS) ./...; \
	done

.PHONY: lint
lint:
	golangci-lint run --max-same-issues 10

.PHONY: bench
bench:
	$(GO) test -run=^$$ -bench=. -cpuprofile=cpu.prof -memprofile=mem.prof -benchmem | perl -pe 's/[ \t]+/\t/g if m{allocs/op$$}'
