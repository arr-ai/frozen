package frozen

// nodeBuilder provides a more efficient way to build nodes incrementally.
type nodeBuilder struct {
	t unTree
}

func newNodeBuilder(capacity int) *nodeBuilder {
	return &nodeBuilder{}
}

func (b *nodeBuilder) Count() int {
	return b.t.count
}

func (b *nodeBuilder) Add(args *combineArgs, v interface{}) {
	matches := 0
	nodeAdd(b.t.Root(), args, v, 0, newHasher(v, 0), &matches, &b.t.root)
	b.t.count += 1 - matches
}

func (b *nodeBuilder) Remove(args *eqArgs, v interface{}) {
	removed := 0
	nodeRemove(b.t.Root(), args, v, 0, newHasher(v, 0), &removed, &b.t.root)
	b.t.count -= removed
}

func (b *nodeBuilder) Get(args *eqArgs, el interface{}) *interface{} {
	return b.t.Get(args, el)
}

func (b *nodeBuilder) Finish() tree {
	t := tree{root: b.t.Root().Freeze(), count: b.t.count}
	*b = nodeBuilder{}
	return t
}

func nodeAdd(
	n unNode,
	args *combineArgs,
	v interface{},
	depth int,
	h hasher,
	matches *int,
	out *unNode,
) {
	*out = n.Add(args, v, depth, h, matches)
}

func nodeRemove(
	n unNode,
	args *eqArgs,
	v interface{},
	depth int,
	h hasher,
	matches *int,
	out *unNode,
) {
	*out = n.Remove(args, v, depth, h, matches)
}
