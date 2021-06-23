package tree

// Builder provides a more efficient way to build nodes incrementally.
type Builder struct {
	t unTree
}

func NewBuilder(capacity int) *Builder {
	return &Builder{}
}

func (b *Builder) Count() int {
	return b.t.count
}

func (b *Builder) Add(args *CombineArgs, v interface{}) {
	matches := 0
	nodeAdd(b.t.Root(), args, v, 0, newHasher(v, 0), &matches, &b.t.root)
	b.t.count += 1 - matches
}

func (b *Builder) Remove(args *EqArgs, v interface{}) {
	removed := 0
	root := b.t.Root()
	h := newHasher(v, 0)
	nodeRemove(root, args, v, 0, h, &removed, &b.t.root)
	b.t.count -= removed
}

func (b *Builder) Get(args *EqArgs, el interface{}) *interface{} {
	return b.t.Get(args, el)
}

func (b *Builder) Finish() Tree {
	t := Tree{root: b.t.Root().Freeze(), count: b.t.count}
	*b = Builder{}
	return t
}

func nodeAdd(
	n unNode,
	args *CombineArgs,
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
	args *EqArgs,
	v interface{},
	depth int,
	h hasher,
	matches *int,
	out *unNode,
) {
	*out = n.Remove(args, v, depth, h, matches)
}
