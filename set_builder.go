package frozen

// SetBuilder provides a more efficient way to build sets incrementally.
type SetBuilder struct {
	nb nodeBuilder
}

func NewSetBuilder(capacity int) *SetBuilder {
	return &SetBuilder{nb: *newNodeBuilder(capacity)}
}

// Count returns the count of the Set that will be returned from Finish().
func (b *SetBuilder) Count() int {
	return b.nb.Count()
}

// Add adds el to the Set under construction.
func (b *SetBuilder) Add(v interface{}) {
	b.nb.Add(defaultNPCombineArgs, v)
}

// Remove removes el to the Set under construction.
func (b *SetBuilder) Remove(v interface{}) {
	b.nb.Remove(defaultNPEqArgs, v)
}

func (b *SetBuilder) Has(v interface{}) bool {
	return b.nb.Get(defaultNPEqArgs, v) != nil
}

// Finish returns a Set containing all elements added since the SetBuilder was
// initialised or the last call to Finish.
func (b *SetBuilder) Finish() Set {
	return newSet(b.nb.Finish())
}
