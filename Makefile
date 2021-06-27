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
	hasher.go \
	leaf.go \
	masker.go \
	node_intf.go \
	node_ptr.go \
	node_ptr_safe.go \
	node_ptr_unsafe.go \
	nodeArgs.go \
	packer.go \
	packer_iter.go \
	tree.go \

.PHONY: gen
gen: gen-kv

.PHONY: gen-kv
gen-kv: gen-kv-iterator gen-kv-tree

gen-kv-%:
	./gen-kv.sh internal/$* internal/$*/$($*.package) $($*.files)

LENGTHS = -short ""
TAGSES = "" frozen_ptr_safe frozen_intf

.PHONY: test
test:
	@set -e; \
	for length in $(LENGTHS); do \
		for tags in $(TAGSES); do \
			printf "\e[1mgo test \e[32m$$length \e[35m$${tags:+-tags=$$tags}\e[0;1m $(TESTFLAGS) ./...\e[0m\n"; \
			go test $$length $${tags:+-tags=$$tags} $(TESTFLAGS) ./... \
				| (fgrep -v '[no test files]' || true); \
		done \
	done

.PHONY: lint
lint: gen
	golangci-lint run
