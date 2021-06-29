.PHONY: all
all: lint test

iterator.package = kvi
iterator.files = \
	empty.go \
	iterator.go \
	slice.go

tree.package = kvt
tree.files = \
	branch.go \
	builder.go \
	empty.go \
	hasher.go \
	leaf.go \
	node.go \
	nodeArgs.go \
	packer.go \
	packer_iter.go \
	tree.go \
	twig.go \
	vet.go

.PHONY: gen
gen: gen-kv

.PHONY: gen-kv
gen-kv: gen-kv-iterator gen-kv-tree

gen-kv-%:
	./gen-kv.sh internal/$* internal/$*/$($*.package) $($*.files)

LENGTHS = -short ""

.PHONY: test
test:
	@set -e; \
	for length in $(LENGTHS); do \
		printf "\e[1mgo test \e[32m$$length $(TESTFLAGS) ./...\e[0m\n"; \
		go test $$length $(TESTFLAGS) ./... \
			| (fgrep -v '[no test files]' || true); \
	done

.PHONY: lint
lint: gen
	golangci-lint run

.PHONY: bench
bench:
	go test -run=^$$ -bench=. -cpuprofile=cpu.prof -memprofile=mem.prof -benchmem | perl -pe 's/[ \t]+/\t/g if m{allocs/op$$}'
