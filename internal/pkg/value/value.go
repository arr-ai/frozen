package value

import (
	"github.com/arr-ai/hash"
)

// Equaler supports equality comparison with values of the same type.
type Equaler[T any] interface {
	Equal(t T) bool
}

// Samer supports equality comparison with values of any type. It is the
// non-generic counterpart of Equaler.
type Samer interface {
	Same(a any) bool
}

func equalEqualer[T any](a, b T) bool {
	var i any = a
	return i.(Equaler[T]).Equal(b)
}

func equalSamer[T any](a, b T) bool {
	var i any = a
	return i.(Samer).Same(b)
}

func equalComparable[T any](a, b T) bool {
	return any(a) == any(b)
}

func equalAny[T any](a, b T) bool {
	var i any = a
	switch a := i.(type) {
	case Equaler[T]:
		return a.Equal(b)
	case Samer:
		return a.Same(b)
	}
	return i == any(b)
}

// EqualFunc returns an equality tester optimised for T.
func EqualFunc[T any]() func(a, b T) bool {
	var t T
	var i any = t
	switch i.(type) {
	case Equaler[T]:
		return equalEqualer[T]
	case Samer:
		return equalSamer[T]
	case nil:
		return equalAny[T]
	}
	if func() (comp bool) {
		defer recover()
		_ = map[any]struct{}{i: {}}
		return true
	}() {
		return equalComparable[T]
	}
	return equalAny[T]
}

// Key represents a type that can be used as a key in a Map or a Set.
type Key[T any] interface {
	Equaler[T]
	hash.Hashable
}

// Equal returns true iff a == b. If a or b implements Equaler, that is used
// to perform the test.
func Equal[T any](a, b T) bool {
	var i, j any = a, b
	if a, ok := i.(Equaler[T]); ok {
		return a.Equal(b)
	}
	return i == j
}
