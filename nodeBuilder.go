package frozen

// nodeBuilder provides a more efficient way to build nodes incrementally.
type nodeBuilder struct {
	root nodeRoot
}

func newNodeBuilder(capacity int) *nodeBuilder {
	return &nodeBuilder{}
}

func (b *nodeBuilder) Count() int {
	return b.root.count
}

func (b *nodeBuilder) Add(args *combineArgs, v interface{}) {
	matched := 0
	nodeAdd(b.root.x(), args, v, 0, newHasher(v, 0), &matched, &b.root.n)
	b.root.count += 1 - matched
}

func (b *nodeBuilder) Remove(args *eqArgs, v interface{}) {
	removed := 0
	nodeRemove(b.root.x(), args, v, 0, newHasher(v, 0), &removed, &b.root.n)
	b.root.count -= removed
}

func (b *nodeBuilder) Get(args *eqArgs, el interface{}) *interface{} {
	return b.root.get(args, el)
}

func (b *nodeBuilder) Finish() nodeRoot {
	root := b.root
	if root.n != nil {
		root.n = root.n.canonical(0)
	} else {
		root.n = emptyNode{}
	}
	root.count = b.Count()
	*b = nodeBuilder{}
	return root
}

func nodeAdd(
	n node,
	args *combineArgs,
	v interface{},
	depth int,
	h hasher,
	matches *int,
	out *node,
) {
	*out = n.with(args, v, depth, h, matches)
}

func nodeRemove(n node,
	args *eqArgs,
	v interface{},
	depth int,
	h hasher,
	matches *int,
	out *node,
) {
	*out = n.without(args, v, depth, h, matches)
}
