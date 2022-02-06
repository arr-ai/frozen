package value

import (
	"github.com/arr-ai/hash"
)

// Equatable represents a type that can be compared for equality with another
// value.
type Equatable[T any] interface {
	Equal(T) bool
}

// Key represents a type that can be used as a key in a Map or a Set.
type Key[T any] interface {
	Equatable[T]
	hash.Hashable
}

// Equal returns true iff a == b. If a or b implements Equatable, that is used
// to perform the test.
func Equal[T any](a, b T) bool {
	var i, j any = a, b
	if a, ok := i.(Equatable[T]); ok {
		return a.Equal(b)
	}
	return i == j
}
