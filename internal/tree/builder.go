package tree

// Builder provides a more efficient way to build nodes incrementally.
type Builder struct {
	t Tree
}

func NewBuilder(capacity int) *Builder {
	return &Builder{}
}

func (b *Builder) Count() int {
	return b.t.count
}

func (b *Builder) Add(args *CombineArgs, v elementT) {
	matches := 0
	b.t.root = b.t.MutableRoot().Add(args, v, 0, newHasher(v, 0), &matches)
	b.t.count += 1 - matches
}

func (b *Builder) Remove(args *EqArgs, v elementT) {
	removed := 0
	root := b.t.MutableRoot()
	h := newHasher(v, 0)
	b.t.root = root.Remove(args, v, 0, h, &removed)
	b.t.count -= removed
}

func (b *Builder) Get(args *EqArgs, el elementT) *elementT {
	return b.t.Get(args, el)
}

func (b *Builder) Finish() Tree {
	t := Tree{root: b.t.Root(), count: b.t.count}
	*b = Builder{}
	return t
}
