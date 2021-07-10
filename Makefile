.PHONY: all
all: lint test

iterator.package = kvi
iterator.root = internal/iterator
iterator.files = \
	empty.go \
	iterator.go \
	slice.go

tree.package = kvt
tree.root = internal/tree
tree.files = \
	branch.go \
	builder.go \
	hasher.go \
	leaf.go \
	node.go \
	nodeArgs.go \
	packer.go \
	packer_iter.go \
	tree.go \
	twig.go \
	vet.go

siterator.package = skvi
siterator.root = $(iterator.root)
siterator.files = $(iterator.files)

stree.package = skvt
stree.root = $(tree.root)
stree.files = $(tree.files)

.PHONY: gen
gen: gen-kv

.PHONY: gen-kv
gen-kv: \
	gen-kv-iterator gen-kv-tree \
	gen-kv-siterator gen-kv-stree

gen-kv-%:
	./gen-kv.sh $($*.root) $($*.root)/$($*.package) $($*.files)

LENGTHS = -short ""

.PHONY: test
test:
	@set -e; \
	for length in $(LENGTHS); do \
		printf "\e[1mgo test \e[32m$$length $(TESTFLAGS) ./...\e[0m\n"; \
		go test $$length $(TESTFLAGS) ./...; \
	done

.PHONY: lint
lint: gen
	golangci-lint run

.PHONY: bench
bench:
	go test -run=^$$ -bench=. -cpuprofile=cpu.prof -memprofile=mem.prof -benchmem | perl -pe 's/[ \t]+/\t/g if m{allocs/op$$}'
