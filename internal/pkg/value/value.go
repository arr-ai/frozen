package value

import (
	"reflect"
	"runtime"
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
	return equalBoxed(i, any(b))
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
	case nil:
		return equalSlow[T]
	}
	// Use reflect.Kind to catch derived types (e.g., type MyFloat float64).
	// Float and string need direct comparison via unsafe to avoid boxing.
	switch reflect.TypeOf(t).Kind() { //nolint:exhaustive
	case reflect.Float32:
		return func(a, b T) bool {
			return *(*float32)(unsafe.Pointer(&a)) == *(*float32)(unsafe.Pointer(&b))
		}
	case reflect.Float64:
		return func(a, b T) bool {
			return *(*float64)(unsafe.Pointer(&a)) == *(*float64)(unsafe.Pointer(&b))
		}
	case reflect.String:
		return func(a, b T) bool {
			return *(*string)(unsafe.Pointer(&a)) == *(*string)(unsafe.Pointer(&b))
		}
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
	return equalBoxed(i, any(b))
}

// equalBoxed compares two already-boxed values with ==, which is only valid
// when their dynamic type (once the two sides agree on one) is comparable.
// T being a statically comparable/interface type does not guarantee that: a
// generic parameter such as `any` accepts dynamic types (e.g. structs or
// slices embedding a slice/map) that panic on ==. Neither Equaler nor Samer
// caught a.'s type above, so falling back here is a last resort; treat
// mismatched or uncomparable dynamic types as unequal rather than panicking.
//
// reflect.Type.Comparable is a static check and is not sufficient on its own:
// a struct whose fields are interfaces (or an array of interfaces) is
// comparable by type, yet == still panics at runtime when those interface
// fields hold uncomparable dynamic values. Only the runtime can decide that,
// so the == itself runs under a recover that turns the comparison panic into
// "unequal". Callers only use this answer to decide whether an existing
// value can be reused, so unequal is always the safe answer.
func equalBoxed(a, b any) (eq bool) {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	// == on interfaces only panics when both dynamic types match and that
	// type is uncomparable; mismatched dynamic types safely compare unequal.
	if ta := reflect.TypeOf(a); ta == reflect.TypeOf(b) && !ta.Comparable() {
		return false
	}
	defer func() {
		if r := recover(); r != nil {
			if _, is := r.(runtime.Error); !is {
				panic(r)
			}
			eq = false
		}
	}()
	return a == b
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
