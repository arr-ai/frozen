package value

import (
	"reflect"
	"runtime"
	"sync"
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
	// Same dispatch as the function EqualFuncFor returns for the slow path;
	// keeping one implementation means both the Map/Set operations and the
	// builders (which compare through mapEntry.Equal) agree on equality.
	return equalSlow(a, b)
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
	// == is only safe for T when it is comparable for *every* value of T. A
	// type containing an interface (an array or struct of them) is
	// statically comparable but panics at runtime once that interface holds
	// an uncomparable value, so those go to the checked slow path.
	if !containsInterface(reflect.TypeOf(t)) && isComparable(i) {
		return equalComparable[T]
	}
	return equalSlow[T]
}

// isComparable reports whether values of i's dynamic type can be used as map
// keys, which is Go's own definition of comparable. The probe panics for
// uncomparable types, so it runs under a real recover: `defer recover()`
// does not stop a panic, because recover must be *called by* a deferred
// function rather than be one.
func isComparable(i any) (comp bool) {
	defer func() {
		if r := recover(); r != nil {
			comp = false
		}
	}()
	_ = map[any]struct{}{i: {}}
	return true
}

// containsInterface reports whether t is, or transitively contains, an
// interface — the case where == compiles but can panic at runtime.
func containsInterface(t reflect.Type) bool {
	return containsInterfaceDepth(t, 0)
}

func containsInterfaceDepth(t reflect.Type, depth int) bool {
	// Recursive types are possible; a struct cannot nest deeply enough for
	// this bound to matter in practice, and exceeding it fails safe.
	const maxDepth = 16
	if t == nil || depth > maxDepth {
		return true
	}
	switch t.Kind() { //nolint:exhaustive
	case reflect.Interface:
		return true
	case reflect.Array:
		return containsInterfaceDepth(t.Elem(), depth+1)
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			if containsInterfaceDepth(t.Field(i).Type, depth+1) {
				return true
			}
		}
	}
	return false
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
	j := any(b)
	// When T is an interface, the element's own Equal method may take an
	// interface that T does not name — e.g. T is `any` but the values are
	// arr.ai Values with Equal(Value) bool. Neither Equaler[T] nor Samer
	// matches, yet the type does define equality, and falling through to ==
	// would report two equal values as unequal whenever their type is not
	// comparable. Dispatch on the dynamic type instead.
	if ta := reflect.TypeOf(i); ta != nil && ta == reflect.TypeOf(j) {
		if f := dynamicEqualFunc(ta); f != nil {
			return f(i, j)
		}
	}
	return equalBoxed(i, j)
}

// equalMethodCache memoises the reflective Equal lookup per dynamic type.
var equalMethodCache sync.Map // reflect.Type -> equalMethod

type equalMethod struct {
	f func(a, b any) bool
}

// dynamicEqualFunc returns a function calling t's own Equal method, or nil
// if t has no method of the form Equal(X) bool with a value of t assignable
// to X.
func dynamicEqualFunc(t reflect.Type) func(a, b any) bool {
	if m, ok := equalMethodCache.Load(t); ok {
		return m.(equalMethod).f //nolint:forcetypeassert
	}
	var f func(a, b any) bool
	if m, ok := t.MethodByName("Equal"); ok &&
		m.Type.NumIn() == 2 && m.Type.NumOut() == 1 &&
		m.Type.Out(0).Kind() == reflect.Bool &&
		t.AssignableTo(m.Type.In(1)) {
		method := m.Func
		f = func(a, b any) bool {
			return method.Call([]reflect.Value{reflect.ValueOf(a), reflect.ValueOf(b)})[0].Bool()
		}
	}
	equalMethodCache.Store(t, equalMethod{f: f})
	return f
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
