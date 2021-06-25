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
    node.go \
    nodeArgs.go \
    packer.go \
    packer_iter.go \
    tree.go \
    unBranch.go \
    unDefroster.go \
    unEmptyNode.go \
    unLeaf.go \
    unNode.go \
    unTree.go

.PHONY: gen
gen: gen-kv

.PHONY: gen-kv
gen-kv: gen-kv-iterator gen-kv-tree

gen-kv-%:
	./gen-kv.sh internal/$* internal/$*/$($*.package) $($*.files)

.PHONY: test
test:
	go test $(TESTFLAGS) ./...

.PHONY: lint
lint: gen
	golangci-lint run
