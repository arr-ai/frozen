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

.PHONY: test
test:
	for length in "-short" ""; do \
		for tags in "" "frozen_ptr_safe" "frozen_intf"; do \
			printf "go test \e[32m$$length \e[35m$${tags:+-tags=$$tags}\e[0m $(TESTFLAGS) ./...\n"; \
			go test $$length $${tags:+-tags=$$tags} $(TESTFLAGS) ./...; \
		done; \
	done

.PHONY: lint
lint: gen
	golangci-lint run
