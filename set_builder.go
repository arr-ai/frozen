package frozen

import (
	"fmt"

	"github.com/arr-ai/frozen/internal/pkg/tree"
)

// SetBuilder[T] provides a more efficient way to build sets incrementally.
type SetBuilder[T any] struct {
	b tree.Builder[T]
}

func NewSetBuilder[T any](capacity int) *SetBuilder[T] {
	return &SetBuilder[T]{b: *tree.NewBuilder[T](capacity)}
}

// Count returns the count of the Set that will be returned from Finish().
func (b *SetBuilder[T]) Count() int {
	return b.b.Count()
}

// Add adds el to the Set under construction.
func (b *SetBuilder[T]) Add(v T) {
	b.b.Add(v)
}

// Remove removes el to the Set under construction.
func (b *SetBuilder[T]) Remove(v T) {
	b.b.Remove(v)
}

func (b *SetBuilder[T]) Has(v T) bool {
	return b.b.Get(v) != nil
}

// Finish returns a Set containing all elements added since the SetBuilder[T] was
// initialised or the last call to Finish.
func (b *SetBuilder[T]) Finish() Set[T] {
	return newSet(b.b.Finish())
}

func (b *SetBuilder[T]) borrow() Set[T] {
	return newSet(b.b.Borrow())
}

func (b SetBuilder[T]) String() string {
	return b.borrow().String()
}

func (b SetBuilder[T]) Format(f fmt.State, verb rune) {
	b.borrow().Format(f, verb)
}
