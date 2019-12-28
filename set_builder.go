package frozen

// SetBuilder provides a more efficient way to build sets incrementally.
type SetBuilder struct {
	root          *node
	remover       *composer
	redundantAdds int
	attemptedAdds int
}

// Count returns the count of the Set that will be returned from Finish().
func (b *SetBuilder) Count() int {
	return b.attemptedAdds - b.redundantAdds - b.successfulRemoves()
}

// Add adds el to the Set under construction.
func (b *SetBuilder) Add(v interface{}) {
	var matches int
	b.root, matches = b.root.valueUnion(v, true, true, 0, newHasher(v, 0))
	b.redundantAdds += matches
	b.attemptedAdds++
}

// Remove removes el to the Set under construction.
func (b *SetBuilder) Remove(el interface{}) {
	if b.remover == nil {
		b.remover = newDifferenceComposer(0)
		b.remover.mutate = true
	}
	b.root = b.root.apply(b.remover, el)
}

func (b *SetBuilder) Has(el interface{}) bool {
	return b.root.get(el) != nil
}

// Finish returns a Set containing all elements added since the SetBuilder was
// initialised or the last call to Finish.
func (b *SetBuilder) Finish() Set {
	count := b.Count()
	if count == 0 {
		return Set{}
	}
	root := b.root
	*b = SetBuilder{}
	return Set{root: root, count: count}
}

func (b *SetBuilder) successfulRemoves() int {
	if b.remover == nil {
		return 0
	}
	return b.remover.delta.input
}
