package frozen

import "github.com/arr-ai/hash"

// Equatable represents a type that can be compared for equality with another
// value.
type Equatable interface {
	Equal(interface{}) bool
}

// Key represents a type that can be used as a key in a Map or a Set.
type Key interface {
	Equatable
	hash.Hashable
}

// Equal returns true iff a == b. If a or b implements Equatable, that is used
// to perform the test.
func Equal(a, b interface{}) bool {
	if a, ok := a.(Equatable); ok {
		return a.Equal(b)
	}
	if b, ok := b.(Equatable); ok {
		return b.Equal(a)
	}
	return a == b
}
