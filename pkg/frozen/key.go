package frozen

// Equatable represents a type that can be compared for equality with another
// value.
type Equatable interface {
	Equal(interface{}) bool
}

// Hashable represents a type that can evaluate its own hash.
type Hashable interface {
	Hash() uint64
}

// Key represents an Equatable and Hashable type.
type Key interface {
	Equatable
	Hashable
}

func equal(a, b interface{}) bool {
	if a, ok := a.(Equatable); ok {
		return a.Equal(b)
	}
	if b, ok := b.(Equatable); ok {
		return b.Equal(a)
	}
	return a == b
}
