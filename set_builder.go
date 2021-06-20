package frozen

import "github.com/arr-ai/frozen/internal/tree"

// SetBuilder provides a more efficient way to build sets incrementally.
type SetBuilder struct {
	b tree.Builder
}

func NewSetBuilder(capacity int) *SetBuilder {
	return &SetBuilder{b: *tree.NewBuilder(capacity)}
}

// Count returns the count of the Set that will be returned from Finish().
func (b *SetBuilder) Count() int {
	return b.b.Count()
}

// Add adds el to the Set under construction.
func (b *SetBuilder) Add(v interface{}) {
	b.b.Add(tree.DefaultNPCombineArgs, v)
}

// Remove removes el to the Set under construction.
func (b *SetBuilder) Remove(v interface{}) {
	b.b.Remove(tree.DefaultNPEqArgs, v)
}

func (b *SetBuilder) Has(v interface{}) bool {
	return b.b.Get(tree.DefaultNPEqArgs, v) != nil
}

// Finish returns a Set containing all elements added since the SetBuilder was
// initialised or the last call to Finish.
func (b *SetBuilder) Finish() Set {
	return newSet(b.b.Finish())
}
