package value

import (
	"github.com/arr-ai/hash"
)

// Equatable represents a type that can be compared for equality with another
// value.
type Equatable[T comparable] interface {
	Equal(T) bool
}

// Key represents a type that can be used as a key in a Map or a Set.
type Key[T comparable] interface {
	Equatable[T]
	hash.Hashable
}

// Equal returns true iff a == b. If a or b implements Equatable, that is used
// to perform the test.
func Equal[T comparable](a, b T) bool {
	var i, j interface{} = a, b
	if a, ok := i.(Equatable[T]); ok {
		return a.Equal(b)
	}
	return i == j
}
