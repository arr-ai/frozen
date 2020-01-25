package frozen

// SetBuilder provides a more efficient way to build sets incrementally.
type SetBuilder struct {
	root          *node
	prepared      *node
	attemptedAdds int
	redundantAdds int
	removals      int
	cloner        *cloner
}

func NewSetBuilder(capacity int) *SetBuilder {
	return &SetBuilder{cloner: newCloner(true, capacity)}
}

func (b *SetBuilder) getCloner() *cloner {
	if b.cloner == nil {
		b.cloner = theMutator
	}
	return b.cloner
}

// Count returns the count of the Set that will be returned from Finish().
func (b *SetBuilder) Count() int {
	return b.attemptedAdds - b.redundantAdds - b.removals
}

// Add adds el to the Set under construction.
func (b *SetBuilder) Add(v interface{}) {
	b.root = b.root.with(v, useRHS, 0, newHasher(v, 0), &b.redundantAdds, b.getCloner(), &b.prepared)
	b.attemptedAdds++
}

// Remove removes el to the Set under construction.
func (b *SetBuilder) Remove(v interface{}) {
	b.root = b.root.without(v, 0, newHasher(v, 0), &b.removals, b.getCloner(), &b.prepared)
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
