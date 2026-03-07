package value

import (
	"unsafe"
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

// Equal returns true if a and b are equal. It dispatches directly through
// type assertions to avoid the overhead of a sync.Map cache lookup, which
// costs more than the dispatch itself.
func Equal[T any](a, b T) bool {
	var i any = a
	switch a := i.(type) {
	case Equaler[T]:
		return a.Equal(b)
	case Samer:
		return a.Same(b)
	}
	return i == any(b)
}

// EqualFuncFor returns an equality tester optimised for T.
func EqualFuncFor[T any]() func(a, b T) bool {
	var t T
	var i any = t
	switch i.(type) {
	case Equaler[T]:
		return equalEqualer[T]
	case Samer:
		return equalSamer[T]
	case float32:
		return func(a, b T) bool {
			return *(*float32)(unsafe.Pointer(&a)) == *(*float32)(unsafe.Pointer(&b))
		}
	case float64:
		return func(a, b T) bool {
			return *(*float64)(unsafe.Pointer(&a)) == *(*float64)(unsafe.Pointer(&b))
		}
	case string:
		return func(a, b T) bool {
			return *(*string)(unsafe.Pointer(&a)) == *(*string)(unsafe.Pointer(&b))
		}
	case nil:
		return equalSlow[T]
	}
	if f := equalScalar[T](); f != nil {
		return f
	}
	if func() (comp bool) {
		defer recover() //nolint:errcheck
		_ = map[any]struct{}{i: {}}
		return true
	}() {
		return equalComparable[T]
	}
	return equalSlow[T]
}

// equalSlow is the fallback path that always boxes.
func equalSlow[T any](a, b T) bool {
	var i any = a
	switch a := i.(type) {
	case Equaler[T]:
		return a.Equal(b)
	case Samer:
		return a.Same(b)
	}
	return i == any(b)
}

// equalScalar returns a non-boxing equality function for types that fit in a
// machine word. Returns nil if T is not a scalar-sized type.
func equalScalar[T any]() func(T, T) bool {
	switch unsafe.Sizeof(*new(T)) {
	case 1:
		return func(a, b T) bool {
			return *(*uint8)(unsafe.Pointer(&a)) == *(*uint8)(unsafe.Pointer(&b))
		}
	case 2:
		return func(a, b T) bool {
			return *(*uint16)(unsafe.Pointer(&a)) == *(*uint16)(unsafe.Pointer(&b))
		}
	case 4:
		return func(a, b T) bool {
			return *(*uint32)(unsafe.Pointer(&a)) == *(*uint32)(unsafe.Pointer(&b))
		}
	case 8:
		return func(a, b T) bool {
			return *(*uint64)(unsafe.Pointer(&a)) == *(*uint64)(unsafe.Pointer(&b))
		}
	}
	return nil
}
