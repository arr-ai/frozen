.PHONY: all
all: lint test

LENGTHS = -short ""

.PHONY: test
test:
	@set -e; \
	for length in $(LENGTHS); do \
		printf "\e[1mgo test \e[32m$$length $(TESTFLAGS) ./...\e[0m\n"; \
		go test $$length $(TESTFLAGS) ./...; \
	done

.PHONY: lint
lint:
	golangci-lint run

.PHONY: bench
bench:
	go test -run=^$$ -bench=. -cpuprofile=cpu.prof -memprofile=mem.prof -benchmem | perl -pe 's/[ \t]+/\t/g if m{allocs/op$$}'
