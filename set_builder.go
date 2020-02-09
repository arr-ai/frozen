package frozen

import "github.com/arr-ai/frozen/internal/tree"

// SetBuilder provides a more efficient way to build sets incrementally.
type SetBuilder struct {
	root          *tree.Node
	prepared      *tree.Node
	attemptedAdds int
	redundantAdds int
	removals      int
	cloner        *tree.Cloner
}

func NewSetBuilder(capacity int) *SetBuilder {
	return &SetBuilder{cloner: tree.NewCloner(true, capacity)}
}

func (b *SetBuilder) getCloner() *tree.Cloner {
	if b.cloner == nil {
		b.cloner = tree.Mutator
	}
	return b.cloner
}

// Count returns the count of the Set that will be returned from Finish().
func (b *SetBuilder) Count() int {
	return b.attemptedAdds - b.redundantAdds - b.removals
}

// Add adds el to the Set under construction.
func (b *SetBuilder) Add(v interface{}) {
	b.root = b.root.With(v, tree.UseRHS, 0, tree.NewHasher(v, 0), &b.redundantAdds, b.getCloner(), &b.prepared)
	b.attemptedAdds++
}

// AddAll adds elts to the Set under construction.
func (b *SetBuilder) AddSet(elts Set) {
	b.root = b.root.Union(elts.root, tree.UseRHS, 0, &b.redundantAdds, b.getCloner())
	b.attemptedAdds += elts.Count()
}

// Remove removes el to the Set under construction.
func (b *SetBuilder) Remove(v interface{}) {
	b.root = b.root.Without(v, 0, tree.NewHasher(v, 0), &b.removals, b.getCloner(), &b.prepared)
}

func (b *SetBuilder) Has(el interface{}) bool {
	return b.root.Get(el) != nil
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
